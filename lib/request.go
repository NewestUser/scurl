package scurl

import (
	"strings"
	"net/http"
	"fmt"
	"io/ioutil"
	"net/url"
)

func NewRequest(urlStr string, opts ...ReqOption) (*http.Request, error) {
	r, rErr := defaultHttpReq(urlStr)

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

type ReqOption func(r *http.Request) error

func MethodOption(method string) ReqOption {
	return func(req *http.Request) error {
		if method == "" {
			method = http.MethodGet
		}

		req.Method = method
		return nil
	}
}

func HeaderOption(headers ...string) ReqOption {
	return func(req *http.Request) error {

		for _, v := range headers {
			parts := strings.Split(v, `:`)

			if len(parts) != 2 {
				return fmt.Errorf(`header '%s' has a wrong format`, v)
			}

			key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			if key == `` || value == `` {
				return fmt.Errorf(`header cannot be empty %s`, v)
			}

			req.Header[key] = append(req.Header[key], value) // preserve the case of the passed header
		}

		return nil
	}
}

func BodyOption(body string) ReqOption {
	return func(req *http.Request) error {
		if body == "" {
			return nil
		}

		req.Body = ioutil.NopCloser(strings.NewReader(body))
		return nil
	}
}

func defaultHttpReq(target string) (*http.Request, error) {

	if _, err := url.ParseRequestURI(target); err != nil {
		return nil, err
	}

	return http.NewRequest(`GET`, target, nil)
}
