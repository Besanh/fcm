package models

// DeviceCodeResponse represents the response from the device authorization endpoint.
type (
	DeviceCodeResponse struct {
		DeviceCode      string `json:"device_code"`
		UserCode        string `json:"user_code"`
		VerificationURL string `json:"verification_uri"` // or "verification_url" based on provider
		ExpiresIn       int    `json:"expires_in"`       // in seconds
		Interval        int    `json:"interval"`         // in seconds, suggested polling interval
	}

	// TokenResponse represents the token endpoint response.
	TokenResponse struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token,omitempty"`
		Error        string `json:"error,omitempty"`
		ErrorDesc    string `json:"error_description,omitempty"`
	}
)

// Fcm
type (
	RegisterDeviceTokenRequest struct {
		FcmToken string `json:"fcm_token" form:"fcm_token" required:"true"`
	}
)
