package transport

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"time"

	"github.com/juruen/rmapi/log"
	"github.com/juruen/rmapi/model"
	"github.com/juruen/rmapi/util"
)

type AuthType int

type BodyString struct {
	Content string
}

var UnAuthorizedError = errors.New("401 Unauthorized Error")

var RmapiUserAGent = "rmapi"

const (
	EmptyBearer AuthType = iota
	DeviceBearer
	UserBearer
)

const (
	EmptyBody string = ""
)

type HttpClientCtx struct {
	Client *http.Client
	Tokens model.AuthTokens
}

func CreateHttpClientCtx(tokens model.AuthTokens) HttpClientCtx {
	var httpClient = &http.Client{Timeout: 5 * 60 * time.Second}

	return HttpClientCtx{httpClient, tokens}
}

func (ctx HttpClientCtx) addAuthorization(req *http.Request, authType AuthType) {
	var header string

	switch authType {
	case EmptyBearer:
		header = "Bearer"
	case DeviceBearer:
		header = fmt.Sprintf("Bearer %s", ctx.Tokens.DeviceToken)
	case UserBearer:
		header = fmt.Sprintf("Bearer %s", ctx.Tokens.UserToken)
	}

	req.Header.Add("Authorization", header)
}

func (ctx HttpClientCtx) Get(authType AuthType, url string, body interface{}, target interface{}) error {
	bodyReader, err := util.ToIOReader(body)

	if err != nil {
		log.Error.Println("failed to serialize body", err)
		return err
	}

	response, err := ctx.Request(authType, http.MethodGet, url, bodyReader)

	if response != nil {
		defer response.Body.Close()
	}

	if err != nil {
		return err
	}

	return json.NewDecoder(response.Body).Decode(target)
}

func (ctx HttpClientCtx) GetStream(authType AuthType, url string) (io.ReadCloser, error) {
	response, err := ctx.Request(authType, http.MethodGet, url, strings.NewReader(""))

	var respBody io.ReadCloser
	if response != nil {
		respBody = response.Body
	}

	return respBody, err
}

func (ctx HttpClientCtx) Post(authType AuthType, url string, reqBody, resp interface{}) error {
	return ctx.httpRawReq(authType, http.MethodPost, url, reqBody, resp)
}

func (ctx HttpClientCtx) Put(authType AuthType, url string, reqBody, resp interface{}) error {
	return ctx.httpRawReq(authType, http.MethodPut, url, reqBody, resp)
}

func (ctx HttpClientCtx) PutStream(authType AuthType, url string, reqBody io.Reader) error {
	return ctx.httpRawReq(authType, http.MethodPut, url, reqBody, nil)
}

func (ctx HttpClientCtx) Delete(authType AuthType, url string, reqBody, resp interface{}) error {
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

	response, err := ctx.Request(authType, verb, url, contentBody)

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

func (ctx HttpClientCtx) Request(authType AuthType, verb, url string, body io.Reader) (*http.Response, error) {
	request, err := http.NewRequest(verb, url, body)
	if err != nil {
		return nil, err
	}

	ctx.addAuthorization(request, authType)
	request.Header.Add("User-Agent", RmapiUserAGent)

	if log.TracingEnabled {
		drequest, err := httputil.DumpRequest(request, true)
		log.Trace.Printf("request: %s %v", string(drequest), err)
	}

	response, err := ctx.Client.Do(request)

	if err != nil {
		log.Error.Println("http request failed with", err)
		return nil, err
	}

	if log.TracingEnabled {
		defer response.Body.Close()
		dresponse, err := httputil.DumpResponse(response, true)
		log.Trace.Printf("%s %v", string(dresponse), err)
	}

	if response.StatusCode != 200 {
		log.Trace.Printf("request failed with status %d\n", response.StatusCode)
	}

	switch response.StatusCode {
	case http.StatusOK:
		return response, nil
	case http.StatusUnauthorized:
		return response, UnAuthorizedError
	default:
		return response, errors.New(fmt.Sprintf("request failed with status %d", response.StatusCode))
	}
}

func (ctx HttpClientCtx) GetBlobStream(url string) (io.ReadCloser, int64, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, err
	}
	client := &http.Client{}
	response, err := client.Do(req)

	if err != nil {
		return nil, 0, err
	}
	if response.StatusCode == http.StatusNotFound {
		return nil, 0, ErrNotFound
	}
	if response.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("GetBlobStream, status code not ok %d", response.StatusCode)
	}
	var gen int64
	if response.Header != nil {
		genh := response.Header.Get(HeaderGeneration)
		if genh != "" {
			log.Trace.Println("got generation header: ", genh)
			gen, err = strconv.ParseInt(genh, 10, 64)
		}
	}

	return response.Body, gen, err
}

const HeaderGeneration = "x-goog-generation"
const HeaderContentLengthRange = "x-goog-content-length-range"
const HeaderGenerationIfMatch = "x-goog-if-generation-match"
const HeaderContentMD5 = "Content-MD5"

var ErrWrongGeneration = errors.New("wrong generation")
var ErrNotFound = errors.New("not found")

func addSizeHeader(req *http.Request, maxRequestSize int64) {
	if maxRequestSize > 0 {
		req.Header[HeaderContentLengthRange] = []string{fmt.Sprintf("0,%d", maxRequestSize)}
	}
}

func (ctx HttpClientCtx) PutRootBlobStream(url string, gen, maxRequestSize int64, reader io.Reader) (newGeneration int64, err error) {
	req, err := http.NewRequest(http.MethodPut, url, reader)
	if err != nil {
		return
	}
	req.Header.Add("User-Agent", RmapiUserAGent)
	//don't change the header case
	req.Header[HeaderGenerationIfMatch] = []string{strconv.FormatInt(gen, 10)}
	addSizeHeader(req, maxRequestSize)

	if log.TracingEnabled {
		drequest, err := httputil.DumpRequest(req, true)
		log.Trace.Printf("PutRootBlobStream: %s %v", string(drequest), err)
	}
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return
	}

	if log.TracingEnabled {
		defer response.Body.Close()
		dresponse, err := httputil.DumpResponse(response, true)
		log.Trace.Printf("PutRootBlobSteam:Response: %s %v", string(dresponse), err)
	}

	if response.StatusCode == http.StatusPreconditionFailed {
		return 0, ErrWrongGeneration
	}
	if response.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("PutRootBlobStream: got status code %d", response.StatusCode)
	}

	if response.Header == nil {
		return 0, fmt.Errorf("PutRootBlobStream: no response headers")
	}
	generationHeader := response.Header.Get(HeaderGeneration)
	if generationHeader == "" {
		log.Warning.Println("no new generation header")
		return
	}

	log.Trace.Println("new generation header: ", generationHeader)
	newGeneration, err = strconv.ParseInt(generationHeader, 10, 64)
	if err != nil {
		log.Error.Print(err)
	}

	return
}
func (ctx HttpClientCtx) PutBlobStream(url string, reader io.Reader, maxRequestSize int64) (err error) {
	req, err := http.NewRequest(http.MethodPut, url, reader)
	if err != nil {
		return
	}
	req.Header.Add("User-Agent", RmapiUserAGent)
	addSizeHeader(req, maxRequestSize)

	if log.TracingEnabled {
		drequest, err := httputil.DumpRequest(req, true)
		log.Trace.Printf("PutBlobStream: %s %v", string(drequest), err)
	}
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return
	}

	if log.TracingEnabled {
		defer response.Body.Close()
		dresponse, err := httputil.DumpResponse(response, true)
		log.Trace.Printf("PutBlobSteam: Response: %s %v", string(dresponse), err)
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("PutBlobStream: got status code %d", response.StatusCode)
	}

	return nil
}
