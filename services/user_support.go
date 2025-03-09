package services

import (
	"context"
	"errors"
	"fcm/models"
	"time"

	"github.com/sony/gobreaker/v2"
	"resty.dev/v3"
)

type (
	OAuth2Request struct {
		AccessToken string
		Url         string
		CBSetting   CBSetting
		Timeout     time.Duration
	}
)

/*
 * Combine Circuit Breaker pattern and resty
 */
func (s *User) getProfileUser(ctx context.Context, request OAuth2Request) (result models.UserProfile, err error) {
	client := resty.New()
	defer client.Close()
	client.SetTimeout(request.Timeout)

	cbSetting := CBGeneric(request.CBSetting)
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
