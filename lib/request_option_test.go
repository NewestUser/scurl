package scurl

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/url"
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

	_ = opt(req)

	got := req.Header["content-type"][0]

	assert.Equal(t, `application/json`, got)
}

func TestMethodOption(t *testing.T) {
	req := &Target{}
	opt := MethodOption("POST")

	_ = opt(req)

	got := req.Method

	assert.Equal(t, `POST`, got)
}

func TestBodyOptionMultipartFormData(t *testing.T) {
	req := &Target{}

	form := map[string][]string{
		"foo": {"bar"},
		"Zar": {"gar", "dar"},
	}

	opt := MultipartFormBodyOption(form)

	_ = opt(req)

	expected := url.Values{
		"foo": {"bar"},
		"Zar": {"gar", "dar"},
	}.Encode()

	actual, _ := ioutil.ReadAll(req.getBody())

	assert.Equal(t, expected, string(actual))
}
