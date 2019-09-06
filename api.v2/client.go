package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

const defaultUserAgent = "rmapi"
const defaultBaseURL = "https://document-storage-production-dot-remarkable-production.appspot.com"

// A Client manages communication with the Remarkable API.
type Client struct {
	// By making the base URL configurable we can make it
	// testable by passing the URL of a httptest.Server.
	// That also means that the Client is acting upon a same base URL for all requests.
	BaseURL *url.URL

	UserAgent string

	// The api package does not directly handle authentication.
	// Instead, when creating a new client, pass an http.Client that
	// can handle authentication for you.
	// The easiest and recommended way to do this is using the auth package.
	httpClient *http.Client
}

// NewClient instanciates and configures the default URL and
// user agent for the http client.
func NewClient(httpClient *http.Client) *Client {
	url, _ := url.Parse(defaultBaseURL)

	return &Client{
		httpClient: httpClient,
		UserAgent:  defaultUserAgent,
		BaseURL:    url,
	}
}

// newRequest creates an http.Request with a method, a relative url path
// and a payload. Query string parameters are not handled.
func (c *Client) newRequest(method, path string, body interface{}) (*http.Request, error) {
	rel := &url.URL{Path: path}
	url := c.BaseURL.ResolveReference(rel)

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, errors.Wrap(err, "can't encode payload")
		}
	}

	req, err := http.NewRequest(method, url.String(), buf)
	if err != nil {
		return nil, errors.Wrapf(err, "can't create request: %s", url.String())
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	return req, nil
}

// do proceeds to the execution of the request and fills v with the answer.
func (c *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "can't execute request")
	}
	defer resp.Body.Close()

	if v != nil {
		if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
			return nil, errors.Wrap(err, "can't decode response content")
		}
	}

	return resp, nil
}
