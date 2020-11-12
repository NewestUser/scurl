package scurl

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestReturnErrorForIncorrectUrl(t *testing.T) {
	request, err := NewTarget(`:`)

	assert.Nil(t, request)
	assert.NotNil(t, err)
}

func TestReturnErrorForOptionThatReturnsError(t *testing.T) {
	req, err := NewTarget(`http://fake.com`, HeaderOption(`incorrect`))

	assert.Nil(t, req)
	assert.NotNil(t, err)
}

func TestAddingRespIncreasesNumOfTrips(t *testing.T) {
	resp := &MultiResponse{}

	assert.Equal(t, 0, resp.Trips)
	assert.True(t, resp.Empty())

	resp.Add(&Response{})

	assert.Equal(t, 1, resp.Trips)
	assert.False(t, resp.Empty())
}

func TestRespStatusMap(t *testing.T) {
	firstOk := &Response{Response: &http.Response{StatusCode: http.StatusOK}}
	secondOk := &Response{Response: &http.Response{StatusCode: http.StatusOK}}

	firstBad := &Response{Response: &http.Response{StatusCode: http.StatusBadRequest}}

	resp := &MultiResponse{}

	resp.Add(firstOk)
	resp.Add(firstBad)
	resp.Add(secondOk)

	actual := resp.StatusMap()

	expected := map[int][]*Response{
		http.StatusOK:         {firstOk, secondOk},
		http.StatusBadRequest: {firstBad},
	}

	assert.Equal(t, actual, expected)
}
