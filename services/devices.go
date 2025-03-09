package services

import (
	"context"
	"fcm/models"
	"time"
)

type (
	IDevices interface {
		GetDeviceCode(ctx context.Context, clientId, scope, deviceEndpoint string) (result models.DeviceCodeResponse, err error)
	}

	Devices struct{}
)

func NewDevices() IDevices {
	return &Devices{}
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
