package model

type AuthTokens struct {
	DeviceToken string
	UserToken   string
}

type DeviceTokenRequest struct {
	Code       string `json:"code"`
	DeviceDesc string `json:"deviceDesc"`
	DeviceId   string `json:"deviceID"`
}
