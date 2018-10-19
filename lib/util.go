package scurl

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

func copyBody(req *http.Request) ([]byte, error) {

	if req.Body == nil {
		return nil, nil
	}

	allBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	req.Body = ioutil.NopCloser(bytes.NewReader(allBytes))

	return allBytes, nil
}

func copyReq(r *http.Request) *http.Request {

	body, err := copyBody(r)
	if err != nil {
		panic(fmt.Errorf("could not copy http body err: %s", err.Error()))
	}

	req, err := http.NewRequest(r.Method, r.URL.String(), bytes.NewReader(body))
	if err != nil {
		panic(fmt.Errorf("could not construct http request err: %s", err.Error()))
	}

	for k, vs := range r.Header {
		req.Header[k] = make([]string, len(vs))
		copy(req.Header[k], vs)
	}

	if host := req.Header.Get("Host"); host != "" {
		req.Host = host
	}

	return req
}
