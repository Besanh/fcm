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

// init is the entry point of the application
func init() {
	// Load env file
	if err := godotenv.Load(".env"); err != nil {
		panic(err)
	}

	// Set application version and name
	variables.API_VERSION = env.GetStringENV("API_VERSION", "v1.0")
	variables.API_SERVICE_NAME = env.GetStringENV("API_SERVICE_NAME", "fcm")

	// Initialize logs
	initLogger()

	// Initialize redis
	initRedis()

	// Initialize MongoDB
	initMongoDb()

	// Initialize NATS JetStream
	initNatsJetstream()

	// Initialize FCM
	initFcm()
}

// initLogger initializes the logger with the given log level and log file.
// If a log server is provided, it will send the logs to the server.
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
	// If a log server is provided, send the logs to the server.
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
	// Set the logger with the given options.
	log.SetLogger(log.NewSLogger(opts...))
}

// initMongoDb initializes the MongoDB client.
// It gets the MongoDB connection string from the environment variables MONGODB_HOST, MONGODB_PORT, MONGODB_DATABASE,
// MONGODB_USERNAME, MONGODB_PASSWORD, and MONGODB_DEFAULT_AUTH_DB.
// If the connection string is invalid, it panics.
func initMongoDb() {
	mongodbConfig := mongodb.MongoDBConfig{
		Username:      env.GetStringENV("MONGODB_USERNAME", ""),
		Password:      env.GetStringENV("MONGODB_PASSWORD", ""),
		Host:          env.GetStringENV("MONGODB_HOST", "localhost"),
		Port:          env.GetIntENV("MONGODB_PORT", 27017),
		Database:      env.GetStringENV("MONGODB_DATABASE", "fcm"),
		DefaultAuthDb: env.GetStringENV("MONGODB_DEFAULT_AUTH_DB", "admin"),
	}

	var err error
	var db mongodb.IMongoDBClient
	db, err = mongodb.NewMongoDBClient(mongodbConfig)
	if err != nil {
		// If the connection string is invalid, panic
		log.Errorf("mongodb connect error: %v", err)
		panic(err)
	}

	DB = db
}

// initRedis initializes the Redis client.
// It gets the Redis connection string from the environment variables REDIS_HOST, REDIS_PASSWORD, REDIS_DB, REDIS_POOL_SIZE,
// REDIS_POOL_TIMEOUT, REDIS_READ_TIMEOUT, and REDIS_WRITE_TIMEOUT.
// If the connection string is invalid, it panics.
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
		// If the connection string is invalid, panic
		log.Errorf("redis connect error: %v", err)
		panic(err)
	}

	// Initialize the Redis cache with the Redis client.
	cache.RCache = cache.NewRedisCache(redis.Redis.GetClient())
}

// initNatsJetstream initializes the NATS JetStream client.
// It gets the NATS JetStream connection string from the environment variable NATS_JETSTREAM_HOST.
// If the connection string is invalid, it panics.
func initNatsJetstream() {
	nat := &messagequeue.NatsJetStream{
		Config: messagequeue.Config{
			Host: env.GetStringENV("NATS_JETSTREAM_HOST", "localhost:4222"),
		},
	}

	// Connect to NATS JetStream
	if err := nat.Connect(); err != nil {
		// If the connection string is invalid, panic
		log.Errorf("nats jetstream connect error: %v", err)
		panic(err)
	}
}

func initFcm() {
	// fcm
}

// main is the entry point of the application.
func main() {
	// Decrypt and validate the secret key from environment variables.
	isOk, err := util.DecryptSecret(env.GetStringENV("SECRET_KEY", ""))
	if err != nil {
		panic(err) // Terminate if decryption fails.
	} else if !isOk {
		panic(errors.New("secret_key was incorrect")) // Terminate if the secret key is incorrect.
	}

	// Determine the environment mode for Gin framework.
	// Gin
	envMode := env.GetStringENV("ENV", "debug")
	if slices.Contains([]string{"debug", "test", "release"}, envMode) {
		panic(errors.New("env was incorrect")) // Terminate if the environment mode is invalid.
	}

	// Manage session
	// Initialize session management.
	sessionManager = scs.New()
	sessionManager.Lifetime = 24 * time.Hour // Set session lifetime.
	sessionManager.Cookie.Persist = true     // Make session cookies persistent.
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode

	// Set true if using HTTPS
	// Set to true if using HTTPS to ensure secure cookies.
	sessionManager.Cookie.Secure = false

	server := server.NewServer(envMode, sessionManager)
	// Create a new server instance with the specified environment mode and session manager.
	services.ENABLE_LOGIN_MULTI_SESSION = env.GetBoolENV("ENABLE_LOGIN_MULTI_SESSION", false)

	// Configure service settings from environment variables.
	services.GOOGLE_URL_USER_INFO = env.GetStringENV("GOOGLE_URL_USER_INFO", "")

	initServices(server)
	// Initialize services with the server instance.

	server.Start(env.GetStringENV("API_PORT", "8000"))
}

// initServices initializes services with the server instance.
// It creates new instances of the repositories, services, and handlers, and registers them with the server.
func initServices(server *server.Server) {
	// Repositories
	usersRepo := repositories.NewUsers(&DB)
	devicesFcmRepo := repositories.NewDevicesFcm(*redis.Redis.GetClient())
	devicesNotificationRepo := repositories.NewDevicesNotification(DB)

	// Services
	// Create a new OAuth2 configuration from environment variables.
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

	// Create a new OAuth2 client from the OAuth2 configuration.

	oAuth2Client := oauth.NewOAuth2(*services.OAUTH2CONFIG)

	// Handlers
	// Register the user handlers with the server.
	v1.NewUsers(server.Engine, sessionManager, services.NewUser(usersRepo, oAuth2Client))
	// Register the device handlers with the server.
	v1.NewDevices(server.Engine, services.NewDevices())
	// Register the devices FCM handlers with the server.
	v1.NewDevicesFcm(server.Engine, services.NewDevicesFcm(devicesFcmRepo))
	// Register the devices notification handlers with the server.
	v1.NewDevicesNotification(server.Engine, services.NewDeviceNotification(devicesNotificationRepo))

	// Register the health check endpoint with the server.
	v1.NewHealthCheck(server.Engine)
}
