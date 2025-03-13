package fcm

/// Reference:  https://github.com/appleboy/gorush/blob/master/notify/notification_fcm.go

import (
	// "encoding/base64"
	"context"
	"encoding/base64"
	"errors"
	"fcm/common/constant"
	"fcm/common/log"
	"fcm/common/util"
	"fmt"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

type (
	FCMCredential struct {
		Credential  *google.Credentials
		ProjectId   string
		AccessToken string
		ExpiredAt   time.Time
	}

	FCMConfig struct {
		FCMVersion       string `mapstructure:"fcm_version,omitempty" json:"fcm_version,omitempty"`
		AppId            string `mapstructure:"app_id,omitempty" json:"app_id,omitempty"`
		CredentialBase64 string `mapstructure:"credential_base64,omitempty" json:"credential_base64,omitempty"`
	}
)

// FCM Credential for HTTP V1 Push
var FCMV1Credentials = make(map[string]*FCMCredential)

var (
	FcmConfigList map[string]*FCMConfig
)

// / init FcmConfigList
func InitFCMConfig(fcmCfg *FCMConfig, isForce bool) {
	// check if fcmConfig exists in FcmClientList
	// if not exist, insert to list
	fcmConfigElement := GetFCMConfigOfAppId(fcmCfg.AppId)
	if fcmConfigElement == nil || fcmConfigElement.AppId == "" || isForce {
		FcmConfigList[fcmCfg.AppId] = fcmCfg
	}
}

func (cfg *FCMConfig) InitFCMCredential(ctx context.Context, credentialBase64 string) (fcmCredential *FCMCredential, err error) {
	if credentialBase64 == "" || cfg.CredentialBase64 == "" {
		err = errors.New("missing credential json")
		return
	}
	scopes := []string{
		"https://www.googleapis.com/auth/firebase.messaging",
	}
	appId := cfg.AppId
	var ok bool
	fcmCredential, ok = FCMV1Credentials[appId]
	if ok && fcmCredential != nil {
		// check if expired
		// if expired, need to reinit
		if fcmCredential.ExpiredAt.After(time.Now()) {
			return
		}
	}
	credentialJson, err := base64.StdEncoding.DecodeString(credentialBase64)
	if err != nil {
		return
	}
	cred, err := google.CredentialsFromJSON(ctx, credentialJson, scopes...)
	if err != nil {
		return
	}
	tokenSource := cred.TokenSource
	token, err := tokenSource.Token()
	if err != nil {
		return
	}
	// init FCM v1
	InitFCMConfig(cfg, true)
	fcmCredential = &FCMCredential{
		Credential:  cred,
		ProjectId:   cred.ProjectID,
		AccessToken: token.AccessToken,
		ExpiredAt:   token.Expiry,
	}
	FCMV1Credentials[appId] = fcmCredential
	log.Infof("[FCM] init fcm v1 client ~ app_id: %s", appId)
	return
}

// Unregistered checks if the device token is unregistered,
// according to response from FCM server. Useful to determine
// if app is uninstalled.
func IsUnregistered(err error) bool {
	switch err {
	case ErrNotRegistered, ErrMismatchSenderID, ErrMissingRegistration, ErrInvalidRegistration:
		return true

	default:
		return false
	}
}

func (cred *FCMCredential) SendWithContext(ctx context.Context, msg *PushNotification) (response *messaging.BatchResponse, err error) {
	app, err := firebase.NewApp(ctx, &firebase.Config{
		ProjectID: cred.ProjectId,
	}, option.WithTokenSource(oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: cred.AccessToken,
	})))
	if err != nil {
		return
	}
	appMessaging, err := app.Messaging(ctx)
	if err != nil {
		return
	}
	ttl := time.Duration(5) * time.Second
	if msg.TimeToLive != nil {
		ttl = time.Duration(*msg.TimeToLive) * time.Second
	}
	message := &messaging.MulticastMessage{
		Android: &messaging.AndroidConfig{
			Priority:    constant.NORMAL,
			TTL:         &ttl,
			CollapseKey: msg.CollapseKey,
		},
		APNS: &messaging.APNSConfig{
			Headers: map[string]string{
				"apns-priority": "10",
				"sound":         "",
			},
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					MutableContent:   true,
					ContentAvailable: true,
				},
			},
		},
		Notification: msg.Notification,
	}

	if len(msg.Tokens) > 0 {
		message.Tokens = msg.Tokens
	}
	if msg.Priority == constant.HIGH || msg.Priority == constant.NORMAL {
		message.Android.Priority = msg.Priority
	}
	if len(msg.Data) > 0 {
		message.Data = make(map[string]string)
		for k, v := range msg.Data {
			message.Data[k] = fmt.Sprintf("%v", v)
		}
	}

	response, err = appMessaging.SendEachForMulticast(ctx, message)
	return
}

// PushToFCM provide send notification to Android server.
func (cfg *FCMConfig) PushToFCMV1(ctx context.Context, req *PushNotification, maxRetry int) (listResponses []PushResponse, err error) {
	log.Debugf("[PUSH_NOTIFY_FCM_V1] id: %s ~ start push notification for android with fcm http v1", req.ID)
	var (
		cred       *FCMCredential
		retryCount = 0
	)

	if req.Retry > 0 && req.Retry < maxRetry {
		maxRetry = req.Retry
	}

	// check message
	err = req.CheckMessage()
	if err != nil {
		log.Errorf("[PUSH_NOTIFY_FCM_V1] id: %s ~ app_id: %s ~ error: %v", req.ID, cfg.AppId, err)
		return
	}

Retry:
	cred, err = cfg.InitFCMCredential(ctx, cfg.CredentialBase64)
	if err != nil {
		// FCM server error
		log.Errorf("[PUSH_NOTIFY_FCM_V1] id: %s ~ app_id: %s ~ tokens: %v ~ error: %v", req.ID, cfg.AppId, util.MustParseAnyToString(req.Tokens), err)
		return
	}

	br, err := cred.SendWithContext(ctx, req)
	if err != nil {
		// Send Message error
		log.Errorf("[PUSH_NOTIFY_FCM_V1] id: %s ~ app_id: %s ~ tokens: %v ~ error: %v", req.ID, cfg.AppId, util.MustParseAnyToString(req.Tokens), err)
		for _, token := range req.Tokens {
			failPush := addPushResp("fail", token, req.Data, err)
			listResponses = append(listResponses, failPush)
		}
		return
	}
	var newTokens []string
	if br.FailureCount > 0 {
		for idx, resp := range br.Responses {
			if !resp.Success {
				token := req.Tokens[idx]
				if !IsUnregistered(resp.Error) {
					newTokens = append(newTokens, token)
				}
				failPush := addPushResp("Fail", token, req.Data, resp.Error)
				listResponses = append(listResponses, failPush)
			}
		}

	}
	if len(newTokens) > 0 && retryCount < maxRetry {
		retryCount++
		// resend fail token
		req.Tokens = newTokens
		time.Sleep(1 * time.Second)
		goto Retry
	}
	for _, r := range listResponses {
		if r.Status == "Fail" {
			log.Errorf("[PUSH_NOTIFY_FCM_V1] id: %s ~ app_id: %s ~ tokens: %v ~ data: %v ~ fail push ~ error: %v", req.ID, cfg.AppId, util.MustParseAnyToString(req.Tokens), util.MustParseAnyToString(r.Data), r.Message)
		} else {
			log.Infof("[PUSH_NOTIFY_FCM_V1] id: %s ~ app_id: %s ~ tokens: %v ~ data: %v ~ success push", req.ID, cfg.AppId, util.MustParseAnyToString(req.Tokens), util.MustParseAnyToString(r.Data))
		}
	}
	return
}

func addPushResp(status string, token string, data map[string]interface{}, err error) PushResponse {
	resp := PushResponse{
		Token:  token,
		Status: status,
		Data:   data,
	}
	if err != nil {
		resp.Message = err.Error()
	}
	return resp
}

type PushResponse struct {
	Token   string         `json:"token"`
	Status  string         `json:"status"`
	Message string         `json:"message"`
	Data    map[string]any `json:"data"`
}

func GetFCMConfigOfAppId(appId string) *FCMConfig {
	if FcmConfigList == nil {
		FcmConfigList = make(map[string]*FCMConfig)
	}
	cfg, ok := FcmConfigList[appId]
	if !ok {
		return new(FCMConfig)
	}
	return cfg
}
