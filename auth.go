package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func authHttpCtx() *HttpClientCtx {
	authTokens := loadTokens(configPath())
	httpClientCtx := CreateHttpClientCtx(authTokens)

	if authTokens.DeviceToken == "" {
		deviceToken, err := httpClientCtx.newDeviceToken(readCode())

		if err != nil {
			Error.Fatal("failed to crete device token from on-time code")
		}

		Trace.Println("device token %s", deviceToken)

		authTokens.DeviceToken = deviceToken
		saveTokens(configPath(), authTokens)
	}

	if authTokens.UserToken == "" {
		userToken, err := httpClientCtx.newUserToken()

		if err != nil {
			Error.Fatal("failed to crete user token from device token")
		}

		Trace.Println("user token %s", userToken)

		authTokens.UserToken = userToken
		saveTokens(configPath(), authTokens)
	}

	return &httpClientCtx
}

func readCode() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter one-time code (go to https://my.remkarable.com): ")
	code, _ := reader.ReadString('\n')

	if len(code) != 9 {
		Error.Println("Code has the wrong lenth, it should be 8")
		return readCode()
	}

	return strings.TrimSuffix(code, "\n")
}
