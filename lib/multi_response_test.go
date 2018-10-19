package scurl

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
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
