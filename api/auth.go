package api

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/juruen/rmapi/config"
	"github.com/juruen/rmapi/log"
	"github.com/juruen/rmapi/model"
	"github.com/juruen/rmapi/transport"
)

const (
	defaultDeviceDesc string = "desktop-linux"
)

func AuthHttpCtx(reAuth, nonInteractive bool) *transport.HttpClientCtx {
	configPath, err := config.ConfigPath()
	if err != nil {
		log.Error.Fatal("failed to get config path")
	}
	authTokens := config.LoadTokens(configPath)
	httpClientCtx := transport.CreateHttpClientCtx(authTokens)

	if authTokens.DeviceToken == "" {
		if nonInteractive {
			log.Error.Fatal("missing token, not asking, aborting")
		}
		deviceToken, err := newDeviceToken(&httpClientCtx, readCode())

		if err != nil {
			log.Error.Fatal("failed to crete device token from on-time code")
		}

		log.Trace.Println("device token", deviceToken)

		authTokens.DeviceToken = deviceToken
		httpClientCtx.Tokens.DeviceToken = deviceToken

		config.SaveTokens(configPath, authTokens)
	}

	if authTokens.UserToken == "" || reAuth {
		userToken, err := newUserToken(&httpClientCtx)

		if err == transport.ErrUnauthorized {
			log.Trace.Println("Invalid deviceToken, resetting")
			authTokens.DeviceToken = ""
		} else if err != nil {
			log.Error.Fatalln("failed to create user token from device token", err)
		}

		log.Trace.Println("user token:", userToken)

		authTokens.UserToken = userToken
		httpClientCtx.Tokens.UserToken = userToken

		config.SaveTokens(configPath, authTokens)
	}

	return &httpClientCtx
}

func readCode() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter one-time code (go to https://my.remarkable.com/device/desktop/connect): ")
	code, _ := reader.ReadString('\n')

	code = strings.TrimSuffix(code, "\n")
	code = strings.TrimSuffix(code, "\r")

	if len(code) != 8 {
		log.Error.Println("Code has the wrong length, it should be 8")
		return readCode()
	}

	return code
}

func newDeviceToken(http *transport.HttpClientCtx, code string) (string, error) {
	uuid := uuid.New()

	req := model.DeviceTokenRequest{code, defaultDeviceDesc, uuid.String()}

	resp := transport.BodyString{}
	err := http.Post(transport.EmptyBearer, config.NewTokenDevice, req, &resp)

	if err != nil {
		log.Error.Fatal("failed to create a new device token")
		return "", err
	}

	return resp.Content, nil
}

func newUserToken(http *transport.HttpClientCtx) (string, error) {
	resp := transport.BodyString{}
	err := http.Post(transport.DeviceBearer, config.NewUserDevice, nil, &resp)

	if err != nil {
		return "", err
	}

	return resp.Content, nil
}
