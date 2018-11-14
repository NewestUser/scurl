package scurl

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestSingleRequest(t *testing.T) {

	respHandler := func(w http.ResponseWriter, r *http.Request) {

		w.WriteHeader(http.StatusOK)
	}

	fs := httptest.NewServer(http.HandlerFunc(respHandler))
	defer fs.Close()

	req, _ := NewTarget(fs.URL)

	client := NewConcurrentClient(
		FanOutOpt(1),
		RateOpt(&Rate{Freq: 1, Per: 1 * time.Second}),
		DurationOpt(1*time.Second),
	)

	hits := 0
	for range client.DoReq(req) {
		hits++
	}

	assert.Equal(t, 1, hits)
}

func TestReturnFirstResponseIfSecondFails(t *testing.T) {

	var invocationCounter int32 = 0

	blockingHandler := func(w http.ResponseWriter, r *http.Request) {

		if invocationCounter == 0 {
			atomic.AddInt32(&invocationCounter, 1)
			return
		}

		// according to http spec there should be a location header for a redirect
		// this will cause the response to fail
		w.WriteHeader(http.StatusMovedPermanently)
		w.Header().Del(`Location`)
	}

	fs := httptest.NewServer(http.HandlerFunc(blockingHandler))
	defer fs.Close()

	req, _ := NewTarget(fs.URL)

	hits := 0
	for range NewConcurrentClient(FanOutOpt(2)).DoReq(req) {
		hits++
	}

	assert.Equal(t, 1, hits)
}

func TestCancelAllRequestsIfOrdered(t *testing.T) {

	blockingHandler := func(w http.ResponseWriter, r *http.Request) {
	}

	fs := httptest.NewServer(http.HandlerFunc(blockingHandler))
	defer fs.Close()

	request, _ := NewTarget(fs.URL)

	client := NewConcurrentClient(FanOutOpt(2), DurationOpt(1*time.Hour))
	response := client.DoReq(request)

	client.Stop()

	resp, ok := <-response

	assert.Nil(t, resp)
	assert.False(t, ok)

}
