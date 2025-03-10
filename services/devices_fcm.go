package services

import (
	"context"
	"fcm/common/log"
	"fcm/models"
	"fcm/pkgs/oauth"
	"fcm/repositories"
	"fmt"
)

type (
	IDevicesFcm interface {
		RegisterDeviceToken(ctx context.Context, request models.RegisterDeviceTokenRequest) (err error)
		UnRegisterDeviceToken(ctx context.Context, request models.RegisterDeviceTokenRequest) (err error)
	}

	DevicesFcm struct {
		Devices
	}
)

func NewDevicesFcm(oauth2Client oauth.IOAuth2) IDevicesFcm {
	return &DevicesFcm{
		Devices: Devices{
			OAuth2Client: oauth2Client,
		},
	}
}

func (s *DevicesFcm) RegisterDeviceToken(ctx context.Context, request models.RegisterDeviceTokenRequest) (err error) {
	// Check token exist in store and match with device
	deviceId, err := repositories.DevicesFcmRepo.GetDevicesInStore(ctx, request.FcmToken)
	if err != nil {
		log.Error(err)
		return err
	} else if len(deviceId) > 0 {
		// Add to store
		if err = repositories.DevicesFcmRepo.AddDevicesTokenToStore(ctx, request.FcmToken, deviceId); err != nil {
			log.Error(err)
			return err
		}
	}

	// Insert token to device
	if err = repositories.DevicesFcmRepo.InsertToken(ctx, deviceId, request.FcmToken); err != nil {
		log.Error(err)
		return err
	}

	return
}

func (s *DevicesFcm) UnRegisterDeviceToken(ctx context.Context, request models.RegisterDeviceTokenRequest) (err error) {
	// Check token exist in store and match with device
	deviceId, err := repositories.DevicesFcmRepo.GetDevicesInStore(ctx, request.FcmToken)
	if err != nil {
		log.Error(err)
		return err
	} else if len(deviceId) == 0 {
		err = fmt.Errorf("device not found")
		return
	}

	// Remove from store
	if err = repositories.DevicesFcmRepo.RemoveTokenFromDevice(ctx, deviceId, request.FcmToken); err != nil {
		log.Error(err)
		return err
	}

	return
}
