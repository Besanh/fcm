package resty

import (
	"time"

	"resty.dev/v3"
)

type RestyConfig struct {
	Timeout time.Duration
}

func NewResty(config RestyConfig) *resty.Client {
	client := resty.New()
	if config.Timeout > 0 {
		client.SetTimeout(config.Timeout)
	}

	return client
}
