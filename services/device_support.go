package services

import (
	"context"
	"errors"
	"fcm/models"
	circuitbreaker "fcm/pkgs/circuit_breaker"
	"fcm/pkgs/resty"

	"github.com/sony/gobreaker/v2"
)

func (s *Devices) getDeviceCode(ctx context.Context, request OAuth2Request) (result models.DeviceCodeResponse, err error) {
	client := resty.NewResty(resty.RestyConfig{Timeout: request.Timeout})
	defer client.Close()
	client.SetTimeout(request.Timeout)

	cbSetting := circuitbreaker.CBGeneric(request.CBSetting)
	cb := gobreaker.NewCircuitBreaker[models.DeviceCodeResponse](*cbSetting)

	result, err = cb.Execute(func() (res models.DeviceCodeResponse, err error) {
		resp, err := client.R().
			SetHeaders(map[string]string{
				"Authorization": "Bearer " + request.AccessToken,
				"Content-Type":  "application/json",
				"Accept":        "application/json",
			}).
			SetFormData(map[string]string{
				"client_id":  request.ClientId,
				"scope":      request.Scope,
				"grant_type": "urn:ietf:params:oauth:grant-type:device_code",
			}).
			SetResult(&res).
			Post(request.Url)
		if err != nil {
			return
		} else if resp.IsError() || resp.StatusCode() != 200 {
			err = errors.New(resp.Status())
			return
		}

		return
	})

	if err != nil {
		return
	}

	return
}
