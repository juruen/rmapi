package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/juruen/rmapi/common"
	"github.com/juruen/rmapi/log"
	"github.com/juruen/rmapi/util"
)

type AuthType int

type BodyString struct {
	Content string
}

var UnAuthorizedError = errors.New("401 Unauthorized Error")

const (
	EmptyBearer AuthType = iota
	DeviceBearer
	UserBearer
)

const (
	EmptyBody string = ""
)

type HttpClientCtx struct {
	client *http.Client
	tokens common.AuthTokens
}

func CreateHttpClientCtx(tokens common.AuthTokens) HttpClientCtx {
	var httpClient = &http.Client{Timeout: 60 * time.Second}

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

func (ctx HttpClientCtx) httpGet(authType AuthType, url string, body interface{}, target interface{}) error {
	bodyReader, err := util.ToIOReader(body)

	if err != nil {
		log.Error.Println("failed to serialize body", err)
		return err
	}

	response, err := ctx.httpRequest(authType, http.MethodGet, url, bodyReader)

	if response != nil {
		defer response.Body.Close()
	}

	if err != nil {
		return err
	}

	return json.NewDecoder(response.Body).Decode(target)
}

func (ctx HttpClientCtx) httpGetStream(authType AuthType, url string) (io.ReadCloser, error) {
	response, err := ctx.httpRequest(authType, http.MethodGet, url, strings.NewReader(""))

	var respBody io.ReadCloser
	if response != nil {
		respBody = response.Body
	}

	return respBody, err
}

func (ctx HttpClientCtx) httpPost(authType AuthType, url string, reqBody, resp interface{}) error {
	return ctx.httpRawReq(authType, http.MethodPost, url, reqBody, resp)
}

func (ctx HttpClientCtx) httpPut(authType AuthType, url string, reqBody, resp interface{}) error {
	return ctx.httpRawReq(authType, http.MethodPut, url, reqBody, resp)
}

func (ctx HttpClientCtx) httpPutStream(authType AuthType, url string, reqBody io.Reader) error {
	return ctx.httpRawReq(authType, http.MethodPut, url, reqBody, nil)
}

func (ctx HttpClientCtx) httpDelete(authType AuthType, url string, reqBody, resp interface{}) error {
	return ctx.httpRawReq(authType, http.MethodDelete, url, reqBody, resp)
}

func (ctx HttpClientCtx) httpRawReq(authType AuthType, verb, url string, reqBody, resp interface{}) error {
	var contentBody io.Reader

	switch reqBody.(type) {
	case io.Reader:
		contentBody = reqBody.(io.Reader)
	default:
		c, err := util.ToIOReader(reqBody)

		if err != nil {
			log.Error.Println("failed to serialize body", err)
			return nil
		}

		contentBody = c
	}

	response, err := ctx.httpRequest(authType, verb, url, contentBody)

	if response != nil {
		defer response.Body.Close()
	}

	if err != nil {
		return err
	}

	// We want to ingore the response
	if resp == nil {
		return nil
	}

	switch resp.(type) {
	case *BodyString:
		bodyContent, err := ioutil.ReadAll(response.Body)

		if err != nil {
			return err
		}

		resp.(*BodyString).Content = string(bodyContent)
	default:
		err := json.NewDecoder(response.Body).Decode(resp)

		if err != nil {
			log.Error.Println("failed to deserialize body", err, response.Body)
			return err
		}
	}
	return nil
}

func (ctx HttpClientCtx) httpRequest(authType AuthType, verb, url string, body io.Reader) (*http.Response, error) {
	request, _ := http.NewRequest(verb, url, body)

	ctx.addAuthorization(request, authType)

	drequest, err := httputil.DumpRequest(request, true)
	log.Trace.Printf("request: %s", string(drequest))

	response, err := ctx.client.Do(request)

	if err != nil {
		log.Error.Printf("http request failed with", err)
		return nil, err
	}

	defer response.Body.Close()

	dresponse, err := httputil.DumpResponse(response, true)
	log.Trace.Print(string(dresponse))

	if response.StatusCode != 200 {
		log.Trace.Printf("request failed with status %i\n", response.StatusCode)
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
