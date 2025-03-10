package services

import (
	"context"
	"encoding/json"
	"fcm/common/cache"
	"fcm/common/constant"
	"fcm/common/log"
	"fcm/common/util"
	"fcm/models"
	circuitbreaker "fcm/pkgs/circuit_breaker"
	"fcm/pkgs/oauth"
	"fcm/repositories"
	"fmt"
	"net/url"
	"time"

	"github.com/alexedwards/scs/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/oauth2"
)

type (
	IUsers interface {
		Login(ctx context.Context) (url string)
		OAuth2Callback(ctx context.Context, callbackData *models.OAuth2Callback) (token string, err error)
	}

	Users struct {
		UserRepo       repositories.IUsers
		OAuth2Client   oauth.IOAuth2
		sessionManager *scs.SessionManager
	}
)

func NewUser(userRepo repositories.IUsers, oauth2Client oauth.IOAuth2) IUsers {
	return &Users{
		UserRepo:     userRepo,
		OAuth2Client: oauth2Client,
	}
}

/*
 * Login and return url(google app)
 */
func (s *Users) Login(ctx context.Context) (callbackUrl string) {
	// use PKCE to protect against CSRF attacks
	// https://www.ietf.org/archive/id/draft-ietf-oauth-security-topics-22.html#name-countermeasures-6
	verifier := oauth2.GenerateVerifier()
	// Generate PKCE values.
	challenge := util.GenerateCodeChallenge(verifier)

	// Generate the authorization URL with PKCE parameters.
	authURL, err := url.Parse(s.OAuth2Client.AuthCodeUrl(OAUTH2_STATE, verifier))
	if err != nil {
		log.Error(err)
		return
	}

	q := authURL.Query()
	q.Set("code_challenge", challenge)
	q.Set("code_challenge_method", "S256")
	authURL.RawQuery = q.Encode()

	// Caching
	redisKey := fmt.Sprintf("pkce:%s", OAUTH2_STATE)
	if err := cache.RCache.Set(redisKey, verifier, 3*time.Minute); err != nil {
		log.Error(err)
		return
	}

	// Redirect user to consent page to ask for permission
	// for the scopes specified above.
	callbackUrl = s.OAuth2Client.AuthCodeUrl(OAUTH2_STATE, verifier)
	return
}

func (s *Users) OAuth2Callback(ctx context.Context, callbackData *models.OAuth2Callback) (token string, err error) {
	// s.sessionManager.Put(ctx, fmt.Sprintf("authenticated:%s", callbackData.State), true)
	// s.sessionManager.Put(ctx, "user_id", "user123")
	// s.sessionManager.Put(ctx, "device_id", "deviceXYZ")

	// PKCE
	verifierCache := cache.RCache.Get(fmt.Sprintf("pkce:%s", callbackData.State))
	if verifierCache == nil {
		err = fmt.Errorf("invalid state: %s", callbackData.State)
		return
	}
	verifier := verifierCache.(string)

	// Exchange the code for a token
	userInfo, err := s.OAuth2Client.Exchange(ctx, callbackData.Code, oauth2.SetAuthURLParam("code_verifier", verifier))
	if err != nil {
		log.Error(err)
		return
	}
	// Calculate TTL for the access token based on its expiry time
	ttl := time.Until(userInfo.Expiry)

	refreshTokenEncrypted, err := util.Encrypt(userInfo.RefreshToken)
	if err != nil {
		log.Error(err)
		return
	}

	user := models.User{
		GBase:  models.InitBase(),
		Status: constant.USER_STATUS_ACTIVE,
	}

	// Get user profile
	userProfile, err := s.getProfileUser(ctx, OAuth2Request{
		AccessToken: userInfo.AccessToken,
		Url:         GOOGLE_URL_USER_INFO,
		CBSetting: circuitbreaker.CBSetting{
			CBName:     "GOOGLE_USER_PROFILE",
			MaxRequest: 1,
			Interval:   5 * time.Second,
			TimeOut:    5 * time.Second,
			MaxTripCB:  3,
		},
		Timeout: 5 * time.Second,
	})
	if err != nil {
		log.Debug(err)
		return
	}
	user.UserProfile = userProfile

	// Marshal data and put to redis
	data, err := json.Marshal(user)
	if err != nil {
		log.Error(err)
		return
	}

	// Start transaction redis
	redisTx := cache.RCache.TxPineLine()
	cache.RCache.TxSet(ctx, redisTx, fmt.Sprintf("%s:%s", OAUTH2_TOKEN, userInfo.AccessToken), data, ttl)
	user.RefreshTokenEncrypted = refreshTokenEncrypted

	// Start transaction mongodb
	mongoSession, err := s.UserRepo.StartSession()
	if err != nil {
		log.Error(err)
		return
	}

	defer mongoSession.EndSession(ctx)

	// Transaction mongodb
	if err = mongo.WithSession(ctx, mongoSession, func(sc mongo.SessionContext) (err error) {
		if err = s.UserRepo.StartTransaction(mongoSession); err != nil {
			return
		}

		// Check email
		total, _, err := s.UserRepo.Select(ctx, 1, 0, repositories.Filter{Key: "email", Value: userProfile.Email})
		if err != nil {
			s.UserRepo.AbortTransaction(ctx, mongoSession)
			return err
		} else if total == 0 {
			if err = s.UserRepo.Insert(ctx, &user); err != nil {
				s.UserRepo.AbortTransaction(ctx, mongoSession)
				return err
			}
		}

		// Nothing to do if user already exist

		if err = s.UserRepo.CommitTransaction(ctx, mongoSession); err != nil {
			s.UserRepo.AbortTransaction(ctx, mongoSession)
			return
		}

		return
	}); err != nil {
		log.Error(err)
		return
	}

	// Commit redis
	if _, err = redisTx.Exec(ctx); err != nil {
		log.Error(err)
		return
	}

	token = userInfo.AccessToken

	return
}
