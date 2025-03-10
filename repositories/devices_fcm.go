package repositories

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type (
	IDevicesFcm interface {
		InsertToken(ctx context.Context, deviceId string, tokens ...string) (err error)
		RemoveTokenFromDevice(ctx context.Context, deviceId string, tokens ...string) (err error)
		GetDeviceTokens(ctx context.Context, deviceId string) (tokens []string, err error)
		AddDevicesTokenToStore(ctx context.Context, token string, deviceId string) (err error)
		GetDevicesInStore(ctx context.Context, token string) (deviceId string, err error)
		RemoveTokenInStore(ctx context.Context, token string) (err error)
	}

	DevicesFcm struct {
		RedisClient   redis.Client
		KeyTokenStore string
	}
)

var DevicesFcmRepo IDevicesFcm

func NewDevicesFcm(redisClient redis.Client) IDevicesFcm {
	return &DevicesFcm{
		RedisClient:   redisClient,
		KeyTokenStore: "devices_token_fcm",
	}
}

func (repo *DevicesFcm) InsertToken(ctx context.Context, deviceId string, tokens ...string) (err error) {
	values := make([]any, 0)
	for _, v := range tokens {
		values = append(values, v)
	}

	if err = repo.RedisClient.LPush(ctx, deviceId, values...).Err(); err != nil {
		return
	}
	return
}

func (repo *DevicesFcm) RemoveTokenFromDevice(ctx context.Context, deviceId string, tokens ...string) (err error) {
	if err = repo.RedisClient.LRem(ctx, deviceId, 0, tokens).Err(); err != nil {
		return
	}
	return
}

func (repo *DevicesFcm) GetDeviceTokens(ctx context.Context, deviceId string) (tokens []string, err error) {
	tokens, err = repo.RedisClient.LRange(ctx, deviceId, 0, -1).Result()
	if err == redis.Nil {
		err = nil
	}
	return
}

// Find token with belong to device in token store
func (repo *DevicesFcm) GetDevicesInStore(ctx context.Context, token string) (deviceId string, err error) {
	deviceId, err = repo.RedisClient.HGet(ctx, repo.KeyTokenStore, token).Result()
	if err == redis.Nil {
		err = nil
	}
	return
}

// Remove all token with belong to device
func (repo *DevicesFcm) RemoveTokenInStore(ctx context.Context, token string) (err error) {
	if err = repo.RedisClient.HDel(ctx, repo.KeyTokenStore, token).Err(); err != nil {
		return
	}
	return
}

// Add token + device to store
func (repo *DevicesFcm) AddDevicesTokenToStore(ctx context.Context, token, deviceId string) (err error) {
	if err = repo.RedisClient.HSet(ctx, repo.KeyTokenStore, token, deviceId).Err(); err != nil {
		return
	}
	return
}
