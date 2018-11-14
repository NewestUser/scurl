package scurl

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

var DefaultMethod = http.MethodGet


type Target struct {
	Method string      `json:"method"`
	URL    string      `json:"url"`
	Body   []byte      `json:"body,omitempty"`
	Header http.Header `json:"header,omitempty"`
}

// Request creates an *http.Request with the provided context.Context out of Target and returns it along with an
// error in case of failure.
func (t *Target) RequestWithContext(c context.Context) (*http.Request, error) {
	req, err := http.NewRequest(t.Method, t.URL, bytes.NewReader(t.Body))
	if err != nil {
		return nil, err
	}

	for k, vs := range t.Header {
		req.Header[k] = make([]string, len(vs))
		copy(req.Header[k], vs)
	}
	if host := req.Header.Get("Host"); host != "" {
		req.Host = host
	}

	return req.WithContext(c), nil
}

func NewTarget(urlStr string, opts ...ReqOption) (*Target, error) {
	r, rErr := defaultTarget(urlStr)

	if rErr != nil {
		return nil, rErr
	}

	for _, opt := range opts {
		if err := opt(r); err != nil {
			return nil, err
		}
	}

	return r, nil
}

type ReqOption func(r *Target) error

func MethodOption(method string) ReqOption {
	return func(req *Target) error {
		if method == "" {
			method = http.MethodGet
		}

		req.Method = method
		return nil
	}
}

func HeaderOption(headers ...string) ReqOption {
	return func(req *Target) error {

		for _, v := range headers {
			parts := strings.Split(v, `:`)

			if len(parts) != 2 {
				return fmt.Errorf(`header '%s' has a wrong format`, v)
			}

			key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			if key == `` || value == `` {
				return fmt.Errorf(`header cannot be empty %s`, v)
			}

			if req.Header == nil {
				req.Header = http.Header{}
			}

			req.Header[key] = append(req.Header[key], value) // preserve the case of the passed header
		}

		return nil
	}
}

func BodyOption(body string) ReqOption {
	return func(req *Target) error {
		if len(body) == 0 {
			return nil
		}

		req.Body = []byte(body)
		return nil
	}
}

func defaultTarget(target string) (*Target, error) {

	if _, err := url.ParseRequestURI(target); err != nil {
		return nil, err
	}

	return &Target{Method: `GET`, URL: target}, nil
}
