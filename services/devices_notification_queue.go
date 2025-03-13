package services

import (
	"context"
	"encoding/json"
	"fcm/common/constant"
	"fcm/common/log"
	"fcm/common/util"
	"fcm/models"
	"fcm/pkgs/fcm"
	messagequeue "fcm/pkgs/message_queue"
	"fcm/repositories"
	"time"

	"github.com/adjust/rmq/v5"
)

type (
	IDeviceNotificationQueue interface {
		PushNotificationToQueue(ctx context.Context, payload models.NotificationTask) (err error)
		HandleQueueTasks()
	}

	DeviceNotificationQueue struct {
		QueueName    string
		DefaultAppId string
		DeviceFcm    repositories.IDevicesFcm
	}
)

var DeviceNotificationQueueService IDeviceNotificationQueue

func NewDeviceNotificationQueue(queueName string, defaultAppId string, deviceTokenRepo repositories.IDevicesFcm) IDeviceNotificationQueue {
	return &DeviceNotificationQueue{
		QueueName:    queueName,
		DefaultAppId: defaultAppId,
		DeviceFcm:    deviceTokenRepo,
	}
}

func (s *DeviceNotificationQueue) HandleQueueTasks() {
	isExisted := messagequeue.RMQ.Server.IsHasQueue(s.QueueName)
	if isExisted {
		messagequeue.RMQ.Server.RemoveQueue(s.QueueName)
	}
	numConsumer := 10
	if err := messagequeue.RMQ.Server.AddQueue(s.QueueName, s.handleQueueTasks, numConsumer); err != nil {
		log.Errorf("create queue error: %v", err)
	}
}

func (s *DeviceNotificationQueue) handleQueueTasks(d rmq.Delivery) {
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()
	defer d.Ack()

	log.Debugf("Websocket receive message: %s", d.Payload())

	payload := new(models.NotificationTask)

	if err := util.ParseAnyToAny(d.Payload(), payload); err != nil {
		log.Error(err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// get device tokens of extension and domain
	deviceId := payload.UserId

	// get device tokens
	var tokens []string
	var err error
	tokens, err = s.DeviceFcm.GetDeviceTokens(ctx, deviceId)
	if err != nil {
		log.Error(err)
		return
	} else if len(tokens) < 1 {
		log.Errorf("[PUSH_NOTIFY] id: %s ~ app_id: %s ~ device not found", payload.Id, s.DefaultAppId)
		return
	}
	timeToLive := uint(5) // 5s

	// setup notification
	notify := fcm.PushNotification{
		Tokens:           tokens,
		Priority:         constant.HIGH,
		TimeToLive:       &timeToLive,
		Data:             payload.Data,
		ContentAvailable: true,
		AppId:            s.DefaultAppId,
	}

	fcmConfig := fcm.GetFCMConfigOfAppId(s.DefaultAppId)
	// send notification
	responses, err := fcmConfig.PushToFCMV1(ctx, &notify, 5)
	if err != nil {
		log.Errorf("[PUSH_NOTIFY] id: %s ~ app_id: %s ~ push to android error: %v", payload.Id, s.DefaultAppId, err)

	}
	for _, e := range responses {
		if e.Status == "fail" {
			messageErr := e.Message
			tmp := map[string]any{
				"token":  e.Token,
				"app_id": fcmConfig.AppId,
				"error":  messageErr,
			}
			log.Errorf("[PUSH_NOTIFY] id: %s ~ app_id: %s ~ push to android error: %v", payload.Id, s.DefaultAppId, tmp)
		}
	}
}

func (q *DeviceNotificationQueue) PushNotificationToQueue(ctx context.Context, payload models.NotificationTask) (err error) {
	// push payload to queue
	var b []byte
	if b, err = json.Marshal(payload); err != nil {
		log.Error(err)
		return
	}
	if err = messagequeue.RMQ.Client.PublishBytes(q.QueueName, b); err != nil {
		log.Error(err)
		return
	}
	return
}
