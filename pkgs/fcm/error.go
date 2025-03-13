package fcm

import "errors"

var (
	// ErrMissingRegistration occurs if registration token is not set.
	ErrMissingRegistration = errors.New("missing registration token")

	// ErrInvalidRegistration occurs if registration token is invalid.
	ErrInvalidRegistration = errors.New("invalid registration token")

	// ErrNotRegistered occurs when application was deleted from device and
	// token is not registered in FCM.
	ErrNotRegistered = errors.New("unregistered device")

	// ErrInvalidPackageName occurs if package name in message is invalid.
	ErrInvalidPackageName = errors.New("invalid package name")

	// ErrMismatchSenderID occurs when application has a new registration token.
	ErrMismatchSenderID = errors.New("mismatched sender id")

	// ErrMessageTooBig occurs when message is too big.
	ErrMessageTooBig = errors.New("message is too big")

	// ErrInvalidDataKey occurs if data key is invalid.
	ErrInvalidDataKey = errors.New("invalid data key")

	// ErrInvalidTTL occurs when message has invalid TTL.
	ErrInvalidTTL = errors.New("invalid time to live")

	// ErrDeviceMessageRateExceeded occurs when client sent to many requests to
	// the device.
	ErrDeviceMessageRateExceeded = errors.New("device message rate exceeded")

	// ErrTopicsMessageRateExceeded occurs when client sent to many requests to
	// the topics.
	ErrTopicsMessageRateExceeded = errors.New("topics message rate exceeded")

	// ErrInvalidParameters occurs when provided parameters have the right name and type
	ErrInvalidParameters = errors.New("check that the provided parameters have the right name and type")

	// ErrUnknown for unknown error type
	ErrUnknown = errors.New("unknown error type")

	// ErrInvalidApnsCredential for Invalid APNs credentials
	ErrInvalidApnsCredential = errors.New("invalid APNs credentials")
)
