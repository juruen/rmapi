package api

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/juruen/rmapi/common"
	"github.com/juruen/rmapi/config"
	"github.com/juruen/rmapi/log"
	uuid "github.com/satori/go.uuid"
)

const (
	defaultDeviceDesc string = "desktop-linux"
)

func AuthHttpCtx() *HttpClientCtx {
	authTokens := config.LoadTokens(config.ConfigPath())
	httpClientCtx := CreateHttpClientCtx(authTokens)

	if authTokens.DeviceToken == "" {
		deviceToken, err := httpClientCtx.newDeviceToken(readCode())

		if err != nil {
			log.Error.Fatal("failed to crete device token from on-time code")
		}

		log.Trace.Println("device token %s", deviceToken)

		authTokens.DeviceToken = deviceToken
		config.SaveTokens(config.ConfigPath(), authTokens)
	}

	userToken, err := httpClientCtx.newUserToken()

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
	fmt.Print("Enter one-time code (go to https://my.remkarable.com): ")
	code, _ := reader.ReadString('\n')

	if len(code) != 9 {
		log.Error.Println("Code has the wrong lenth, it should be 8")
		return readCode()
	}

	return strings.TrimSuffix(code, "\n")
}

func (httpCtx *HttpClientCtx) newDeviceToken(code string) (string, error) {
	uuid, err := uuid.NewV4()

	if err != nil {
		panic(err)
	}

	req := common.DeviceTokenRequest{code, defaultDeviceDesc, uuid.String()}

	resp := BodyString{}
	err = httpCtx.httpPost(EmptyBearer, newTokenDevice, req, &resp)

	if err != nil {
		log.Error.Fatal("failed to create a new device token")
		return "", err
	}

	return resp.Content, nil
}

func (httpCtx *HttpClientCtx) newUserToken() (string, error) {
	resp := BodyString{}
	err := httpCtx.httpPost(DeviceBearer, newUserDevice, nil, &resp)

	if err != nil {
		log.Error.Fatal("failed to create a new user token")

		return "", err
	}

	return resp.Content, nil
}
