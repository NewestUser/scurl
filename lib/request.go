package scurl

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

var DefaultMethod = http.MethodGet

type Target struct {
	Method string
	URL    string
	Body   BodyProvider
	Header http.Header
}

func (t *Target) getBody() io.Reader {
	if t.Body != nil {
		return t.Body.Get()
	}
	return nil
}

type BodyProvider interface {
	Get() io.Reader
}

type StringBody struct {
	value string
}

func (b *StringBody) Get() io.Reader {
	return strings.NewReader(b.value)
}

func (b *StringBody) String() string {
	return b.value
}

type MultipartFormBody struct {
	form          map[string]string
	multipartData []byte
}

func (b *MultipartFormBody) Get() io.Reader {
	return bytes.NewReader(b.multipartData)
}

func (b *MultipartFormBody) String() string {
	return fmt.Sprintf("%s", b.form)
}

// Request creates an *http.Request with the provided context.Context out of Target and returns it along with an
// error in case of failure.
func (t *Target) RequestWithContext(c context.Context) (*http.Request, error) {
	req, err := http.NewRequest(t.Method, t.URL, t.getBody())
	if err != nil {
		return nil, err
	}

	for k, vs := range t.Header {
		req.Header[k] = make([]string, len(vs))
		copy(req.Header[k], vs)
	}
	if host := req.Header.Get("Host"); host != "" {
		req.Host = host
	}

	return req.WithContext(c), nil
}

func NewTarget(urlStr string, opts ...ReqOption) (*Target, error) {
	r, rErr := defaultTarget(urlStr)

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

type ReqOption func(r *Target) error

func MethodOption(method string) ReqOption {
	return func(req *Target) error {
		if method == "" {
			method = http.MethodGet
		}

		req.Method = method
		return nil
	}
}

func HeaderOption(headers ...string) ReqOption {
	return func(req *Target) error {

		for _, v := range headers {
			parts := strings.Split(v, `:`)

			if len(parts) != 2 {
				return fmt.Errorf(`header '%s' has a wrong format`, v)
			}

			key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			if key == `` || value == `` {
				return fmt.Errorf(`header cannot be empty %s`, v)
			}

			if req.Header == nil {
				req.Header = http.Header{}
			}

			req.Header[key] = append(req.Header[key], value) // preserve the case of the passed header
		}

		return nil
	}
}

func StringBodyOption(body string) ReqOption {
	return func(req *Target) error {
		if len(body) == 0 {
			return nil
		}

		req.Body = &StringBody{value: body}
		return nil
	}
}

func MultipartFormBodyOption(formValues map[string]string) ReqOption {
	return func(req *Target) error {
		if len(formValues) == 0 {
			return nil
		}

		// Prepare the multipart form.
		var buff bytes.Buffer
		w := multipart.NewWriter(&buff)
		for key, value := range formValues {
			formField, err := w.CreateFormField(key)
			if err != nil {
				return fmt.Errorf("failed creating multipart form data for key %s and value %s, err: %s", key, value, err)
			}
			if _, err = io.Copy(formField, strings.NewReader(value)); err != nil {
				return fmt.Errorf("failed creating multipart form data for key %s and value %s, err: %s", key, value, err)
			}
		}
		// Close the multipart writer, this will write the terminating boundary.
		if err := w.Close(); err != nil {
			return fmt.Errorf("failed creating multipart form data, err: %s", err)
		}

		formCopy := make([]byte, buff.Len())
		copy(formCopy, buff.Bytes())
		req.Body = &MultipartFormBody{form: formValues, multipartData: formCopy}

		if req.Header == nil {
			req.Header = http.Header{}
		}
		req.Header.Add("Content-Type", w.FormDataContentType())
		return nil
	}
}

func defaultTarget(target string) (*Target, error) {
	if _, err := url.ParseRequestURI(target); err != nil {
		return nil, err
	}

	return &Target{Method: `GET`, URL: target}, nil
}
