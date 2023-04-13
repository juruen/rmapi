package model

type AuthTokens struct {
	DeviceToken string `yaml:"devicetoken"`
	UserToken   string `yaml:"usertoken"`
}

type DeviceTokenRequest struct {
	Code       string `json:"code"`
	DeviceDesc string `json:"deviceDesc"`
	DeviceId   string `json:"deviceID"`
}
