package api

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/juruen/rmapi/config"
	"github.com/juruen/rmapi/log"
	"github.com/juruen/rmapi/model"
	"github.com/juruen/rmapi/transport"
	uuid "github.com/satori/go.uuid"
)

const (
	defaultDeviceDesc string = "desktop-linux"
)

func AuthHttpCtx(reAuth, nonInteractive bool) *transport.HttpClientCtx {
	configPath := config.ConfigPath()
	authTokens := config.LoadTokens(configPath)
	httpClientCtx := transport.CreateHttpClientCtx(authTokens)

	if authTokens.DeviceToken == "" {
		var oneTimeCode string
		if nonInteractive {
			if code, ok := os.LookupEnv("RMAPI_DEVICE_CODE"); ok && len(code) == 8 {
				oneTimeCode = code
			} else {
				log.Error.Fatal("missing token, not asking, aborting")
			}
		} else {
			oneTimeCode = readCode()
		}

		deviceToken, err := newDeviceToken(&httpClientCtx, oneTimeCode)
		if err != nil {
			log.Error.Fatal("failed to create device token from on-time code")
		}

		log.Trace.Println("device token", deviceToken)

		authTokens.DeviceToken = deviceToken
		httpClientCtx.Tokens.DeviceToken = deviceToken

		config.SaveTokens(configPath, authTokens)
	}

	if authTokens.UserToken == "" || reAuth || userTokenExpires(authTokens.UserToken) {
		userToken, err := newUserToken(&httpClientCtx)

		if err == transport.UnAuthorizedError {
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

func userTokenExpires(token string) bool {
	// if there are parsing errors or the registered claim 'exp' is missing/invalid, consider the token to be expired
	// TODO: check if the v1 API JWT's contain the exp claim... if not, return false for the above mentioned cases
	p := jwt.Parser{}
	res, _, err := p.ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		return true
	}
	claims := res.Claims.(jwt.MapClaims)
	if _, ok := claims["exp"]; !ok {
		return true
	}
	exp := time.Unix(int64(claims["exp"].(float64)), 0).UTC()
	// add 1 hour to the actual time, assuming the max. session duration of AuthHttpCtx is 1 hour...
	if time.Now().UTC().Add(time.Hour).After(exp) {
		return true
	}
	return false
}

func newDeviceToken(http *transport.HttpClientCtx, code string) (string, error) {
	uuid, err := uuid.NewV4()

	if err != nil {
		panic(err)
	}

	req := model.DeviceTokenRequest{code, defaultDeviceDesc, uuid.String()}

	resp := transport.BodyString{}
	err = http.Post(transport.EmptyBearer, config.NewTokenDevice, req, &resp)

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
