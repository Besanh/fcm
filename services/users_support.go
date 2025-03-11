package services

import (
	"errors"
	"fcm/models"
	circuitbreaker "fcm/pkgs/circuit_breaker"
	"fcm/pkgs/resty"
	"time"

	"github.com/sony/gobreaker/v2"
)

type (
	OAuth2Request struct {
		AccessToken string
		Url         string
		CBSetting   circuitbreaker.CBSetting
		Timeout     time.Duration

		// Device code
		ClientId   string
		Scope      string
		DeviceCode string
		Interval   int
		ExpiresIn  int
	}
)

/*
 * Combine Circuit Breaker pattern and resty
 */
func (s *Users) getProfileUser(request OAuth2Request) (result models.UserProfile, err error) {
	client := resty.NewResty(resty.RestyConfig{Timeout: request.Timeout})
	defer client.Close()

	cbSetting := circuitbreaker.CBGeneric(request.CBSetting)
	cb := gobreaker.NewCircuitBreaker[models.UserProfile](*cbSetting)

	result, err = cb.Execute(func() (res models.UserProfile, err error) {
		resp, err := client.R().
			SetHeaders(map[string]string{
				"Authorization": "Bearer " + request.AccessToken,
				"Content-Type":  "application/json",
				"Accept":        "application/json",
			}).
			SetResult(&res).
			Get(request.Url)
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
