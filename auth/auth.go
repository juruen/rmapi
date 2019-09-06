// Package auth has the responsibility to handle the authentication
// to the Remarkable Cloud API.
//
// For this purpose, it provides a *http.Client that can be used with the api package.
// This *http.Client will hold the authentication process that will allow the api package
// to interact with the Remarkable API without worrying about auth.
// We do take advantage of a custom http.Transport that is by default attached to the
// http.Client and that will act as a middleware to attach HTTP auth headers.
//
// This separation means that in the future, another auth could be implemented if
// Remarkable decides to change / improve it. As well, the api package is clearer
// because not cluttered by any auth processes.
package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	uuid "github.com/satori/go.uuid"
)

// ClientTimeout is the timeout set for the http.Client
// that auth is providing.
const ClientTimeout time.Duration = time.Second * 10

const (
	defaultDeviceDesc string = "desktop-windows"
	deviceTokenURL    string = "https://my.remarkable.com/token/json/2/device/new"
	userTokenURL      string = "https://my.remarkable.com/token/json/2/user/new"
)

var defaultTokenStore FileTokenStore

// Auth is a structure containing authentication gears to fetch and hold tokens
// for interacting authenticated with the Remarkable Cloud API.
type Auth struct {
	ts TokenStore

	// Refresh can be used to force a refresh of the UserToken.
	Refresh bool
}

func New() *Auth {
	return NewFromStore(&defaultTokenStore)
}

func NewFromStore(ts TokenStore) *Auth {
	return &Auth{ts, false}
}

// RegisterDevice will make an HTTP call to the Remarkable API using the provided code
// to register a new device. The code should be gathered at https://my.remarkable.com/generator-device.
// The DeviceToken is then attached to the Auth instance.
func (a *Auth) RegisterDevice(code string) error {
	uuid, err := uuid.NewV4()
	if err != nil {
		return err
	}

	body, err := json.Marshal(map[string]string{
		"code":       code,
		"deviceDesc": defaultDeviceDesc,
		"deviceID":   uuid.String(),
	})

	req, err := http.NewRequest("POST", deviceTokenURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("auth: can't register device")
	}

	bearer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	tks := TokenSet{
		DeviceToken: string(bearer),
		UserToken:   "",
	}

	// persist device token and reset user token
	if err := a.ts.Save(tks); err != nil {
		return err
	}

	return nil
}

// renewToken will try to fetch a userToken from a deviceToken.
func renewToken(deviceToken string) (userToken string, err error) {
	req, err := http.NewRequest("POST", userTokenURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+deviceToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("auth: can't renew token (HTTP %d)", resp.StatusCode)
	}

	bearer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(bearer), nil
}

// Token will return a UserToken fetching it before if nil.
func (a *Auth) Token() (string, error) {
	tks, err := a.ts.Load()
	if err != nil {
		return "", err
	}

	if tks.UserToken != "" && !a.Refresh {
		return tks.UserToken, nil
	}

	if tks.DeviceToken == "" {
		return "", errors.New("auth: nil DeviceToken, please register device")
	}

	tks.UserToken, err = renewToken(tks.DeviceToken)
	if err != nil {
		return "", err
	}

	if err := a.ts.Save(tks); err != nil {
		return "", err
	}

	// reset the Refresh flag when the token has been renewed
	a.Refresh = false

	return tks.UserToken, nil
}

// Client returns a configured http.Client that will hold a custom Transport
// with authentication capabilities to the Remarkable Cloud API.
func (a *Auth) Client() *http.Client {
	t := Transport{
		Auth: a,
	}

	c := http.Client{
		Transport: &t,
		Timeout:   ClientTimeout,
	}

	return &c
}
