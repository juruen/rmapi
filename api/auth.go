package api

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/juruen/rmapi/config"
	"github.com/juruen/rmapi/log"
	"github.com/juruen/rmapi/model"
	"github.com/juruen/rmapi/transport"
	uuid "github.com/satori/go.uuid"
)

const (
	defaultDeviceDesc string = "desktop-linux"
)

func AuthHttpCtx() *transport.HttpClientCtx {
	authTokens := config.LoadTokens(config.ConfigPath())
	httpClientCtx := transport.CreateHttpClientCtx(authTokens)

	if authTokens.DeviceToken == "" {
		deviceToken, err := newDeviceToken(&httpClientCtx, readCode())

		if err != nil {
			log.Error.Fatal("failed to crete device token from on-time code")
		}

		log.Trace.Println("device token", deviceToken)

		authTokens.DeviceToken = deviceToken
		httpClientCtx.Tokens.DeviceToken = deviceToken

		config.SaveTokens(config.ConfigPath(), authTokens)
	}

	userToken, err := newUserToken(&httpClientCtx)

	if err != nil {
		log.Error.Fatal("failed to create user token from device token")
	}

	log.Trace.Println("user token:", userToken)

	authTokens.UserToken = userToken
	config.SaveTokens(config.ConfigPath(), authTokens)

	return &httpClientCtx
}

func readCode() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter one-time code (go to https://my.remarkable.com/connect/desktop): ")
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
	uuid, err := uuid.NewV4()

	if err != nil {
		panic(err)
	}

	req := model.DeviceTokenRequest{code, defaultDeviceDesc, uuid.String()}

	resp := transport.BodyString{}
	err = http.Post(transport.EmptyBearer, newTokenDevice, req, &resp)

	if err != nil {
		log.Error.Fatal("failed to create a new device token")
		return "", err
	}

	return resp.Content, nil
}

func newUserToken(http *transport.HttpClientCtx) (string, error) {
	resp := transport.BodyString{}
	err := http.Post(transport.DeviceBearer, newUserDevice, nil, &resp)

	if err != nil {
		log.Error.Fatal("failed to create a new user token")

		return "", err
	}

	return resp.Content, nil
}
