package scurl

import (
	"testing"
	"net/http/httptest"
	"net/http"
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"io/ioutil"
	"fmt"
)

func TestSingleRequest(t *testing.T) {

	respHandler := func(w http.ResponseWriter, r *http.Request) {

		w.WriteHeader(http.StatusOK)
	}

	fs := httptest.NewServer(http.HandlerFunc(respHandler))
	defer fs.Close()

	req, _ := NewRequest(fs.URL)

	resp, _ := NewConcurrentClient(1).Do(req)

	assert.Equal(t, 1, resp.Trips)
}

func TestMultipleConcurrentRequests(t *testing.T) {

	routineBlocker := newRoutineBlocker()

	respHandler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("entered")
		routineBlocker.blockCount(4)
		routineBlocker.onceReleaseAll(4)

		w.WriteHeader(http.StatusOK)
	}

	fs := httptest.NewServer(http.HandlerFunc(respHandler))
	defer fs.Close()

	req, _ := NewRequest(fs.URL)
	resp, _ := NewConcurrentClient(5).Do(req)

	assert.Equal(t, 5, resp.Trips)
}

func TestCancelAllRequestsOnFailure(t *testing.T) {
	routineBlocker := newRoutineBlocker()

	blockingHandler := func(w http.ResponseWriter, r *http.Request) {

		routineBlocker.blockCount(2)

		w.WriteHeader(http.StatusMovedPermanently)
		w.Header().Del(`Location`) // according to http spec there should be a location header for a redirect
		routineBlocker.onceReleaseAll(2)
	}

	fs := httptest.NewServer(http.HandlerFunc(blockingHandler))
	defer fs.Close()

	req, _ := NewRequest(fs.URL)

	response, _ := NewConcurrentClient(3).Do(req)

	assert.Equal(t, response.Trips, 0)
}

func TestAllServedRequestsToHaveTheExpectedBody(t *testing.T) {

	routineBlocker := newRoutineBlocker()

	blockingHandler := func(w http.ResponseWriter, r *http.Request) {

		routineBlocker.blockCount(1)

		bytes, e := ioutil.ReadAll(r.Body)
		assert.Nil(t, e)
		assert.Equal(t, `body`, string(bytes))

		routineBlocker.onceReleaseAll(1)

		w.WriteHeader(http.StatusOK)
	}

	fs := httptest.NewServer(http.HandlerFunc(blockingHandler))
	defer fs.Close()

	req, _ := NewRequest(fs.URL, MethodOption(`POST`), BodyOption(`body`))

	response, e := NewConcurrentClient(2).Do(req)

	assert.Nil(t, e)
	assert.Equal(t, 2, response.Trips)
}

func TestReturnFirstResponseIfSecondFails(t *testing.T) {

	var invocationCounter int32 = 0

	blockingHandler := func(w http.ResponseWriter, r *http.Request) {

		if invocationCounter == 0 {
			atomic.AddInt32(&invocationCounter, 1)
			return
		}

		w.WriteHeader(http.StatusMovedPermanently)
		w.Header().Del(`Location`) // according to http spec there should be a location header for a redirec
	}

	fs := httptest.NewServer(http.HandlerFunc(blockingHandler))
	defer fs.Close()

	req, _ := NewRequest(fs.URL)

	response, e := NewConcurrentClient(2).Do(req)

	assert.Nil(t, e)
	assert.Equal(t, 1, response.Trips)
}

func TestCancelAllRequestsIfOrdered(t *testing.T) {

	blocker := newRoutineBlocker()

	blockingHandler := func(w http.ResponseWriter, r *http.Request) {
		blocker.blockCount(2)
	}

	fs := httptest.NewServer(http.HandlerFunc(blockingHandler))
	defer fs.Close()

	request, _ := NewRequest(fs.URL)

	client := NewConcurrentClient(2)
	response := client.DoReq(request)

	client.Stop()

	resp, ok := <-response

	assert.Nil(t, resp)
	assert.False(t, ok)

}

func newRoutineBlocker() *blocker {

	return &blocker{
		blockChan:    make(chan int),
		blockCounter: 0,
		released:     false}
}

type blocker struct {
	blockChan    chan int
	blockCounter int32
	released     bool
}

func (b *blocker) blockCount(count int32) {
	if b.blockCounter < count { // block on thee first n go routines
		atomic.AddInt32(&b.blockCounter, 1)
		<-b.blockChan
	}
}

func (b *blocker) onceReleaseAll(count int) {

	if !b.released { // make sure that the released go routines don't get blocked
		b.released = true
		for i := 0; i < count; i++ {
			b.blockChan <- i
		}
	}
}
