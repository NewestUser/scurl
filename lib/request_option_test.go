package scurl

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWrongFormatOfHeaderOption(t *testing.T) {
	req := &Target{}
	opt := HeaderOption("wrong-format")

	err := opt(req)

	assert.NotNil(t, err)
}

func TestEmptyHeaderOption(t *testing.T) {
	req := &Target{}
	opt := HeaderOption("content-type:")

	err := opt(req)

	assert.NotNil(t, err)
}

func TestHeaderOption(t *testing.T) {

	req := &Target{}
	opt := HeaderOption("content-type: application/json")

	opt(req)

	got := req.Header["content-type"][0]

	assert.Equal(t, `application/json`, got)
}

func TestMethodOption(t *testing.T) {

	req := &Target{}
	opt := MethodOption("POST")

	opt(req)

	got := req.Method

	assert.Equal(t, `POST`, got)
}
