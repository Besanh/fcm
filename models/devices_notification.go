package models

import (
	"errors"
	"slices"
	"time"
)

type (
	DevicesNotification struct {
		*GBase           `bson:",inline"`
		DeviceId         string         `json:"device_id" bson:"device_id"`
		NotificationType string         `json:"notification_type" bson:"notification_type"`
		Time             time.Time      `json:"time" bson:"time"`
		Title            string         `json:"title" bson:"title"`
		Message          string         `json:"message" bson:"message"`
		Sender           string         `json:"sender" bson:"sender"`
		ReadAt           *time.Time     `json:"read_at" bson:"read_at"`
		SnoozedTill      *time.Time     `json:"snoozed_till,omitempty" bson:"snoozed_till"`
		ArchivedAt       *time.Time     `json:"archived_at,omitempty" bson:"archived_at"`
		ReceiverId       string         `json:"receiver_id" bson:"receiver_id"`
		TriggeredById    string         `json:"triggered_by_id,omitempty" bson:"triggered_by_id"`
		UpdatedById      string         `json:"updated_by_id,omitempty" bson:"updated_by_id"`
		Detail           map[string]any `json:"detail" bson:"detail"`
	}

	DevicesNotificationStatsResponse struct {
		TotalUnread int64 `json:"total_unreads"`
	}
)

type (
	DevicesNotificationQueryParam struct {
		DeviceId_Eq string    `json:"device_id_eq"`
		Time_Gte    time.Time `json:"time_gte"`
		Time_Lte    time.Time `json:"time_lte"`
		Type_Eq     string    `json:"type_eq"`
		Sorts       []string  `json:"sorts"`
	}

	DevicesNotificationBulkRequest struct {
		Ids         []string `json:"ids"`
		Action      string   `json:"action" form:"action" required:"true"`
		IsSelectAll bool     `json:"is_select_all" form:"is_select_all" required:"false"`
	}
)

func (m *DevicesNotificationBulkRequest) Validate() (err error) {
	if !slices.Contains([]string{"read", "archive", "snooze", "unarchive"}, m.Action) {
		return errors.New("action was incorrect")
	}

	return
}
