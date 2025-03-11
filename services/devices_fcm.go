package services

import (
	"context"
	"fcm/common/log"
	"fcm/models"
	"fcm/repositories"
	"fmt"
)

type (
	IDevicesFcm interface {
		RegisterDeviceToken(ctx context.Context, request models.RegisterDeviceTokenRequest) (err error)
		UnRegisterDeviceToken(ctx context.Context, request models.RegisterDeviceTokenRequest) (err error)
	}

	DevicesFcm struct {
		DevicesFcmRepo repositories.IDevicesFcm
	}
)

func NewDevicesFcm(devicesFcm repositories.IDevicesFcm) IDevicesFcm {
	return &DevicesFcm{
		DevicesFcmRepo: devicesFcm,
	}
}

func (s *DevicesFcm) RegisterDeviceToken(ctx context.Context, request models.RegisterDeviceTokenRequest) (err error) {
	// Check token exist in store and match with device
	deviceId, err := s.DevicesFcmRepo.GetDevicesInStore(ctx, request.FcmToken)
	if err != nil {
		log.Error(err)
		return err
	} else if len(deviceId) > 0 {
		// Add to store
		if err = s.DevicesFcmRepo.AddDevicesTokenToStore(ctx, request.FcmToken, deviceId); err != nil {
			log.Error(err)
			return err
		}
	}

	// Insert token to device
	if err = s.DevicesFcmRepo.InsertToken(ctx, deviceId, request.FcmToken); err != nil {
		log.Error(err)
		return err
	}

	return
}

func (s *DevicesFcm) UnRegisterDeviceToken(ctx context.Context, request models.RegisterDeviceTokenRequest) (err error) {
	// Check token exist in store and match with device
	deviceId, err := s.DevicesFcmRepo.GetDevicesInStore(ctx, request.FcmToken)
	if err != nil {
		log.Error(err)
		return err
	} else if len(deviceId) == 0 {
		err = fmt.Errorf("device not found")
		return
	}

	// Remove from store
	if err = s.DevicesFcmRepo.RemoveTokenFromDevice(ctx, deviceId, request.FcmToken); err != nil {
		log.Error(err)
		return err
	}

	return
}
