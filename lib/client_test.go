package scurl

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExecuteGetRequest(t *testing.T) {

	respHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	fs := httptest.NewServer(http.HandlerFunc(respHandler))

	req, _ := http.NewRequest(`GET`, fs.URL, nil)

	resp, _ := NewTimedClient().Do(req)

	assert.Equal(t, resp.StatusCode, http.StatusOK)
}

func TestExecuteRequestWithHeaders(t *testing.T) {
	respHandler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application-json", r.Header.Get("content-type"))

		w.WriteHeader(http.StatusOK)
	}

	fs := httptest.NewServer(http.HandlerFunc(respHandler))

	req, _ := http.NewRequest(`GET`, fs.URL, nil)
	req.Header.Add(`content-type`, `application-json`)

	resp, _ := NewTimedClient().Do(req)

	assert.Equal(t, resp.StatusCode, http.StatusOK)
}

func TestExecuteAndFail(t *testing.T) {
	respHandler := func(w http.ResponseWriter, r *http.Request) {
		// according to http spec there should be a location header for a redirect
		// this will cause the response to fail
		w.WriteHeader(http.StatusMovedPermanently)
		w.Header().Del(`Location`)
	}

	fs := httptest.NewServer(http.HandlerFunc(respHandler))

	req, _ := http.NewRequest(`GET`, fs.URL, nil)

	resp, respErr := NewTimedClient().Do(req)

	assert.NotNil(t, respErr)
	assert.Nil(t, resp)
}
