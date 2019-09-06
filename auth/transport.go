package auth

import (
	"net/http"
	"sync"
)

// Transport is an http.RoundTripper that makes requests to
// the Remarkable Cloud API wrapping a base RoundTripper and
// adding an Authorization header with a token from the supplied Auth
type Transport struct {
	// Auth supplies the token to add to outgoing requests'
	// Authorization headers.
	Auth *Auth

	// Base is the base RoundTripper used to make HTTP requests.
	// If nil, http.DefaultTransport is used.
	Base http.RoundTripper

	// Guard for avoiding multi authentications.
	mu sync.Mutex
}

// RoundTrip authorizes and authenticates the request with an
// access token from Transport's Auth.
//
// RoundTrip makes sure req.Body is closed anyway.
// RoundTrip is cloning the original request to respect the RoundTripper contract.
//
// In order to avoid having two authenticating requests at the same time
// we make use of a Mutex.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.mu.Lock()
	token, err := t.Auth.Token()
	if err != nil {
		if req.Body != nil {
			req.Body.Close()
		}
		return nil, err
	}
	t.mu.Unlock()

	req2 := cloneRequest(req) // to respect to RoundTripper contract
	req2.Header.Set("Authorization", "Bearer "+token)

	res, err := t.base().RoundTrip(req2)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (t *Transport) base() http.RoundTripper {
	if t.Base != nil {
		return t.Base
	}
	return http.DefaultTransport
}

// cloneRequest returns a clone of the provided *http.Request.
// The clone is a shallow copy of the struct and its Header map.
// In order to respect the http.RoundTripper interface contract,
// we should normally be doing a full deep copy.
func cloneRequest(req *http.Request) *http.Request {
	// shallow copy of the struct
	copy := new(http.Request)
	*copy = *req
	// deep copy of the Header
	copy.Header = make(http.Header, len(req.Header))
	for k, s := range req.Header {
		copy.Header[k] = append([]string(nil), s...)
	}
	return copy
}
