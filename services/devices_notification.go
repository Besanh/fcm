package services

import (
	"context"
	"errors"
	"fcm/common/log"
	"fcm/models"
	"fcm/repositories"
	"time"
)

type (
	IDevicesNotification interface {
		GetDeviceNotifications(ctx context.Context, authUser *models.UserResponse, param models.DevicesNotificationQueryParam, limit, offset int64) (total int64, entries []*models.DevicesNotification, err error)
		GetTotalUnreadNotifications(ctx context.Context, authUser *models.UserResponse, param models.DevicesNotificationQueryParam) (total int64, err error)
		PutBulkDeviceNotification(ctx context.Context, authUser *models.UserResponse, param models.DevicesNotificationQueryParam, req *models.DevicesNotificationBulkRequest) (err error)
	}

	DevicesNotification struct {
		DevicesNotificationRepo repositories.IDevicesNotification
	}
)

func NewDeviceNotification(devicesNotification repositories.IDevicesNotification) IDevicesNotification {
	return &DevicesNotification{
		DevicesNotificationRepo: devicesNotification,
	}
}

// Get notifications
func (s *DevicesNotification) GetDeviceNotifications(ctx context.Context, authUser *models.UserResponse, param models.DevicesNotificationQueryParam, limit, offset int64) (total int64, entries []*models.DevicesNotification, err error) {
	// TODO: check permission
	total, entries, err = s.DevicesNotificationRepo.SelectInfo(ctx, param, limit, offset)
	if err != nil {
		log.Error(ctx, err)
		return
	}
	return
}

// Get total unread notifications
func (s *DevicesNotification) GetTotalUnreadNotifications(ctx context.Context, authUser *models.UserResponse, param models.DevicesNotificationQueryParam) (total int64, err error) {
	// TODO: check permission
	total, err = s.DevicesNotificationRepo.CountUnreadNotifications(ctx, param)
	if err != nil {
		log.Error(ctx, err)
		return
	}
	return
}

// Put bulk notifications
func (s *DevicesNotification) PutBulkDeviceNotification(ctx context.Context, authUser *models.UserResponse, param models.DevicesNotificationQueryParam, req *models.DevicesNotificationBulkRequest) (err error) {
	// TODO: check permission
	var notifications []*models.DevicesNotification

	if req.IsSelectAll {
		var limit, offset int64 = 100, 0
	LOOP:
		var entries []*models.DevicesNotification
		_, entries, err = s.DevicesNotificationRepo.SelectInfo(ctx, param, limit, offset)
		if err != nil {
			log.Error(ctx, err)
			return
		} else {
			notifications = append(notifications, entries...)
		}
		if len(entries) > 0 {
			offset += limit
			goto LOOP
		}
	} else {
		for _, id := range req.Ids {
			var notification *models.DevicesNotification
			notification, err = s.DevicesNotificationRepo.GetById(ctx, id)
			if err != nil {
				log.Error(err)
				return
			} else if notification != nil {
				notifications = append(notifications, notification)
			}
		}
	}
	current := time.Now()
	switch req.Action {
	case "read":
		for _, n := range notifications {
			n.ReadAt = &current
			if err = s.DevicesNotificationRepo.UpdateById(ctx, n); err != nil {
				log.Error(err)
				return
			}
		}
	case "archive":
		for _, n := range notifications {
			n.ArchivedAt = &current
			if err = s.DevicesNotificationRepo.UpdateById(ctx, n); err != nil {
				log.Error(err)
				return
			}
		}
	case "snooze":
		// TODO: implement snooze
	case "unarchive":
		// TODO: implement unarchive
	default:
		err = errors.New("action was incorrect")
		return
	}

	return
}
