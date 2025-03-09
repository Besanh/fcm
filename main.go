package main

import (
	"errors"
	v1 "fcm/apis/v1"
	"fcm/common/cache"
	"fcm/common/env"
	"fcm/common/util"
	"fcm/common/variables"
	messagequeue "fcm/pkgs/message_queue"
	"fcm/pkgs/mongodb"
	"fcm/pkgs/oauth"
	"fcm/pkgs/redis"
	"fcm/repositories"
	"fcm/server"
	"fcm/services"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/alexedwards/scs/v2"
	log "github.com/besanh/logger/logging/slog"
	"golang.org/x/oauth2"

	"github.com/fluent/fluent-logger-golang/fluent"
	"github.com/joho/godotenv"
)

var (
	DB             mongodb.IMongoDBClient
	sessionManager *scs.SessionManager
)

func init() {
	if err := godotenv.Load(".env"); err != nil {
		panic(err)
	}

	variables.API_VERSION = env.GetStringENV("API_VERSION", "v1.0")
	variables.API_SERVICE_NAME = env.GetStringENV("API_SERVICE_NAME", "fcm")

	initLogger()
	initRedis()
	initMongoDb()
	initNatsJetstream()
	initFcm()
}

func initLogger() {
	logFile := "tmp/console.log"
	logLevel := log.LEVEL_DEBUG
	switch env.GetStringENV("LOG_LEVEL", "error") {
	case "debug":
		logLevel = log.LEVEL_DEBUG
	case "info":
		logLevel = log.LEVEL_INFO
	case "error":
		logLevel = log.LEVEL_ERROR
	case "warn":
		logLevel = log.LEVEL_WARN
	}
	opts := []log.Option{}
	opts = append(opts, log.WithLevel(logLevel),
		log.WithRotateFile(logFile),
		log.WithFileSource(),
		log.WithTraceId(),
		log.WithAttrs(slog.Attr{
			Key: "environment", Value: slog.StringValue(env.GetStringENV("ENVIRONMENT", "local")),
		}),
	)
	if env.GetStringENV("LOG_SERVER", "") != "" {
		// get server and port from env
		arr := strings.Split(env.GetStringENV("LOG_SERVER", ""), ":")
		if len(arr) >= 2 {
			tag := "fcm"
			client, err := fluent.New(fluent.Config{FluentPort: int(util.ParseInt64(arr[1])), FluentHost: arr[0]})
			if err != nil {
				log.Error(err)
			} else {
				opts = append(opts, log.WithFluentd(client, tag))
			}
		}
	}
	log.SetLogger(log.NewSLogger(opts...))
}

func initMongoDb() {
	mongodbConfig := mongodb.MongoDBConfig{
		Username:      env.GetStringENV("MONGODB_USERNAME", "root"),
		Password:      env.GetStringENV("MONGODB_PASSWORD", "anhle@!*2025"),
		Host:          env.GetStringENV("MONGODB_HOST", "localhost"),
		Port:          env.GetIntENV("MONGODB_PORT", 27017),
		Database:      env.GetStringENV("MONGODB_DATABASE", "fcm"),
		DefaultAuthDb: env.GetStringENV("MONGODB_DEFAULT_AUTH_DB", "admin"),
	}

	var err error
	var db mongodb.IMongoDBClient
	db, err = mongodb.NewMongoDBClient(mongodbConfig)
	if err != nil {
		log.Errorf("mongodb connect error: %v", err)
		panic(err)
	}

	DB = db
}

func initRedis() {
	redisClient := &redis.RedisConfig{
		Host:         env.GetStringENV("REDIS_HOST", "localhost"),
		Password:     env.GetStringENV("REDIS_PASSWORD", ""),
		DB:           env.GetIntENV("REDIS_DB", 0),
		PoolSize:     env.GetIntENV("REDIS_POOL_SIZE", 10),
		PoolTimeout:  env.GetIntENV("REDIS_POOL_TIMEOUT", 10),
		ReadTimeout:  env.GetIntENV("REDIS_READ_TIMEOUT", 10),
		WriteTimeout: env.GetIntENV("REDIS_WRITE_TIMEOUT", 10),
	}

	var err error
	if redis.Redis, err = redis.NewRedis(*redisClient); err != nil {
		panic(err)
	}

	cache.RCache = cache.NewRedisCache(redis.Redis.GetClient())
}

func initNatsJetstream() {
	nat := &messagequeue.NatsJetStream{
		Config: messagequeue.Config{
			Host: env.GetStringENV("NATS_JETSTREAM_HOST", "localhost:4222"),
		},
	}
	if err := nat.Connect(); err != nil {
		panic(err)
	}
}

func initFcm() {
	// fcm
}

func main() {
	isOk, err := util.DecryptSecret(env.GetStringENV("SECRET_KEY", ""))
	if err != nil {
		panic(err)
	} else if !isOk {
		panic(errors.New("secret_key was incorrect"))
	}

	// Gin
	envMode := env.GetStringENV("ENV", "debug")
	if slices.Contains([]string{"debug", "test", "release"}, envMode) {
		panic(errors.New("env was incorrect"))
	}

	// Manage session
	sessionManager = scs.New()
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.Cookie.Persist = true
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode

	// Set true if using HTTPS
	sessionManager.Cookie.Secure = false

	server := server.NewServer(envMode, sessionManager)
	services.ENABLE_LOGIN_MULTI_SESSION = env.GetBoolENV("ENABLE_LOGIN_MULTI_SESSION", false)
	services.GOOGLE_URL_USER_INFO = env.GetStringENV("GOOGLE_URL_USER_INFO", "")

	initServices(server)

	server.Start(env.GetStringENV("API_PORT", "8000"))
}

func initServices(server *server.Server) {
	repositories.UserRepo = repositories.NewUser(&DB)

	oau2Scope := env.GetSliceStringENV("OAUTH2_SCOPE", []string{})
	services.OAUTH2CONFIG = &oauth.OAuth2Config{
		ClientId:     env.GetStringENV("OAUTH2_CLIENT_ID", ""),
		ClientSecret: env.GetStringENV("OAUTH2_CLIENT_SECRET", ""),
		Scopes:       oau2Scope,
		Endpoint: oauth2.Endpoint{
			AuthURL:  env.GetStringENV("OAUTH2_ENDPOINT_AUTH_URL", ""),
			TokenURL: env.GetStringENV("OAUTH2_ENDPOINT_TOKEN_URL", ""),
		},
		Redirect: env.GetStringENV("OAUTH2_REDIRECT_URL", ""),
	}

	oAuth2Client := oauth.NewOAuth2(*services.OAUTH2CONFIG)

	services.NewUser(repositories.UserRepo, oAuth2Client)

	// Handler
	v1.NewUser(server.Engine, sessionManager, services.NewUser(repositories.UserRepo, oAuth2Client))
}
