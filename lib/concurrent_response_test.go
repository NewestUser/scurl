package scurl

import (
	"testing"
	"net/http"
	"github.com/stretchr/testify/assert"
)

func TestCanCloseEmptyResponse(t *testing.T) {

	resp := &MultiResponse{}
	resp.Close()
}

func TestAddResponse(t *testing.T) {

	resp := &MultiResponse{}

	resp.Add(&Response{Response: &http.Response{}})

	assert.Equal(t, 1, len(resp.Responses))

}
