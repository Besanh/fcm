package services

import (
	"context"
	"fcm/models"
	"fcm/pkgs/oauth"
	"time"
)

type (
	IDevices interface {
		GetDeviceCode(ctx context.Context, clientId, scope, deviceEndpoint string) (result models.DeviceCodeResponse, err error)
		GetPollForToken(ctx context.Context, clientId, deviceCode, tokenEndpoint string, interval, expiresInt int) (result models.TokenResponse, err error)
	}

	Devices struct {
		OAuth2Client oauth.IOAuth2
	}
)

func NewDevices(oauth2Client oauth.IOAuth2) IDevices {
	return &Devices{
		OAuth2Client: oauth2Client,
	}
}

func (s *Devices) GetDeviceCode(ctx context.Context, clientId, scope, deviceEndpoint string) (result models.DeviceCodeResponse, err error) {
	result, err = s.getDeviceCode(ctx, OAuth2Request{
		ClientId: clientId,
		Scope:    scope,
		Url:      deviceEndpoint,
		Timeout:  3 * time.Second,
	})

	return
}

func (s *Devices) GetPollForToken(ctx context.Context, clientId, deviceCode, tokenEndpoint string, interval, expiresInt int) (result models.TokenResponse, err error) {
	result, err = s.getPollForToken(ctx, OAuth2Request{
		ClientId: clientId,
		Url:      tokenEndpoint,
		Timeout:  3 * time.Second,

		DeviceCode: deviceCode,
		Interval:   interval,
		ExpiresIn:  expiresInt,
	})

	return
}
