package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

type AuthType int

var UnAuthorizedError = errors.New("401 Unauthorized Error")

const (
	EmptyBearer AuthType = iota
	DeviceBearer
	UserBearer
)

const (
	EmptyBody string = ""
)

type AuthTokens struct {
	DeviceToken string
	UserToken   string
}

type HttpClientCtx struct {
	client *http.Client
	tokens AuthTokens
}

func CreateHttpClientCtx(tokens AuthTokens) HttpClientCtx {
	var httpClient = &http.Client{Timeout: 30 * time.Second}

	return HttpClientCtx{httpClient, tokens}
}

func (ctx HttpClientCtx) addAuthorization(req *http.Request, authType AuthType) {
	var header string

	switch authType {
	case EmptyBearer:
		header = "Bearer"
	case DeviceBearer:
		header = fmt.Sprintf("Bearer %s", ctx.tokens.DeviceToken)
	case UserBearer:
		header = fmt.Sprintf("Bearer %s", ctx.tokens.UserToken)
	}

	req.Header.Add("Authorization", header)
}

func (ctx HttpClientCtx) httpGet(authType AuthType, url, body string, target interface{}) error {
	response, err := ctx.httpRequest(authType, http.MethodGet, url, body)

	if response != nil {
		defer response.Body.Close()
	}

	if err != nil {
		return err
	}

	return json.NewDecoder(response.Body).Decode(target)
}

func (ctx HttpClientCtx) httpPostRaw(authType AuthType, url, reqBody string) (string, error) {
	response, err := ctx.httpRequest(authType, http.MethodPost, url, reqBody)

	if response != nil {
		defer response.Body.Close()
	}

	if err != nil {
		return "", err
	}

	respBody, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return "", err
	}

	return string(respBody), nil
}

func (ctx HttpClientCtx) httpRequest(authType AuthType, verb, url, body string) (*http.Response, error) {
	request, _ := http.NewRequest(verb, url, strings.NewReader(body))

	ctx.addAuthorization(request, authType)

	drequest, err := httputil.DumpRequest(request, true)
	Trace.Printf("request: %s", string(drequest))

	response, err := ctx.client.Do(request)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	dresponse, err := httputil.DumpResponse(response, true)
	Trace.Print(string(dresponse))

	if response.StatusCode != 200 {
		Warning.Printf("request failed with status %i\n", response.StatusCode)
	}

	switch response.StatusCode {
	case http.StatusOK:
		return response, nil
	case http.StatusUnauthorized:
		return response, UnAuthorizedError
	default:
		return response, errors.New(fmt.Sprintf("request failed with status %i", response.StatusCode))
	}
}
