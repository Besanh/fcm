package services

import (
	"context"
	"fcm/models"
	"time"
)

type (
	IDevices interface {
		GetDeviceCode(ctx context.Context, clientId, scope, deviceEndpoint string) (result models.DeviceCodeResponse, err error)
		GetPollForToken(ctx context.Context, clientId, deviceCode, tokenEndpoint string, interval, expiresInt int) (result models.TokenResponse, err error)
	}

	Devices struct {
	}
)

func NewDevices() IDevices {
	return &Devices{}
}

func (s *Devices) GetDeviceCode(ctx context.Context, clientId, scope, deviceEndpoint string) (result models.DeviceCodeResponse, err error) {
	result, err = s.getDeviceCode(OAuth2Request{
		ClientId: clientId,
		Scope:    scope,
		Url:      deviceEndpoint,
		Timeout:  3 * time.Second,
	})

	return
}

func (s *Devices) GetPollForToken(ctx context.Context, clientId, deviceCode, tokenEndpoint string, interval, expiresInt int) (result models.TokenResponse, err error) {
	result, err = s.getPollForToken(OAuth2Request{
		ClientId: clientId,
		Url:      tokenEndpoint,
		Timeout:  3 * time.Second,

		DeviceCode: deviceCode,
		Interval:   interval,
		ExpiresIn:  expiresInt,
	})

	return
}
