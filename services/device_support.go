package services

import (
	"context"
	"errors"
	"fcm/models"
	circuitbreaker "fcm/pkgs/circuit_breaker"
	"fcm/pkgs/resty"
	"fmt"
	"time"

	"github.com/sony/gobreaker/v2"
)

/*
 * Get authorization device code
 * Endpoint: https://accounts.google.com/o/oauth2/device/code
 */
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

func (s *Devices) getPollForToken(ctx context.Context, request OAuth2Request) (result models.TokenResponse, err error) {
	expiryTime := time.Now().Add(time.Duration(request.ExpiresIn) * time.Second)

	for {
		if time.Now().After(expiryTime) {
			err = fmt.Errorf("device code expired")
			return
		}

		time.Sleep(time.Duration(request.Interval) * time.Second)

		client := resty.NewResty(resty.RestyConfig{Timeout: request.Timeout})
		defer client.Close()
		client.SetTimeout(request.Timeout)

		cbSetting := circuitbreaker.CBGeneric(request.CBSetting)
		cb := gobreaker.NewCircuitBreaker[models.TokenResponse](*cbSetting)

		result, err = cb.Execute(func() (res models.TokenResponse, err error) {
			resp, err := client.R().
				SetHeaders(map[string]string{
					// "Authorization": "Bearer " + request.AccessToken,
					"Content-Type": "application/json",
					"Accept":       "application/json",
				}).
				SetFormData(map[string]string{
					"client_id":   request.ClientId,
					"device_code": request.DeviceCode,
					"grant_type":  "urn:ietf:params:oauth:grant-type:device_code",
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

		if result.AccessToken != "" {
			return
		} else if result.Error != "" {
			switch result.Error {
			case "authorization_pending":
				// Continue polling.
			case "slow_down":
				// Increase the interval.
				request.Interval += 3
			default:
				err = fmt.Errorf("token error: %s - %s", result.Error, result.ErrorDesc)
				return
			}
		}
	}
}
