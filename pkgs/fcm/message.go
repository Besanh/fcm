package fcm

import (
	"errors"

	"firebase.google.com/go/v4/messaging"
)

type (
	// PushNotification is single notification request
	PushNotification struct {
		// Common
		ID               string         `json:"notif_id,omitempty"`
		Tokens           []string       `json:"tokens" binding:"required"`
		Platform         int            `json:"platform" binding:"required"`
		Message          string         `json:"message,omitempty"`
		Title            string         `json:"title,omitempty"`
		Image            string         `json:"image,omitempty"`
		Priority         string         `json:"priority,omitempty"`
		ContentAvailable bool           `json:"content_available,omitempty"`
		MutableContent   bool           `json:"mutable_content,omitempty"`
		Sound            any            `json:"sound,omitempty"`
		Data             map[string]any `json:"data,omitempty"`
		Retry            int            `json:"retry,omitempty"`

		// Android
		To                    string                  `json:"to,omitempty"`
		CollapseKey           string                  `json:"collapse_key,omitempty"`
		DelayWhileIdle        bool                    `json:"delay_while_idle,omitempty"`
		TimeToLive            *uint                   `json:"time_to_live,omitempty"`
		RestrictedPackageName string                  `json:"restricted_package_name,omitempty"`
		DryRun                bool                    `json:"dry_run,omitempty"`
		Condition             string                  `json:"condition,omitempty"`
		Notification          *messaging.Notification `json:"notification,omitempty"`

		//AppId
		AppId string `json:"app_id"`

		ClientTokens []ClientToken `json:"client_token"`
	}

	ClientToken struct {
		Token   string `json:"token,omitempty"`
		AppMode string `json:"app_mode,omitempty"`
	}
)

const (
	// PlatFormIos constant is 1 for iOS
	PlatFormIos = iota + 1
	// PlatFormAndroid constant is 2 for Android
	PlatFormAndroid
	// PlatFormHuawei constant is 3 for Huawei
	PlatFormHuawei
)

// CheckMessage for check request message
func (e *PushNotification) CheckMessage() error {
	var msg string

	// ignore send topic mesaage from FCM

	if len(e.Tokens) == 0 && e.To == "" {
		msg = "the message must specify at least one registration ID"
		return errors.New(msg)
	}

	if len(e.Tokens) == PlatFormIos && e.Tokens[0] == "" {
		msg = "the token must not be empty"
		return errors.New(msg)
	}

	if e.Platform == PlatFormAndroid && len(e.Tokens) > 1000 {
		msg = "the message may specify at most 1000 registration IDs"
		return errors.New(msg)
	}

	// ref: https://firebase.google.com/docs/cloud-messaging/http-server-ref
	if e.Platform == PlatFormAndroid && e.TimeToLive != nil && *e.TimeToLive > uint(2419200) {
		msg = "the message's TimeToLive field must be an integer " +
			"between 0 and 2419200 (4 weeks)"
		return errors.New(msg)
	}

	return nil
}
