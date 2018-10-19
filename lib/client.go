package scurl

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func NewTimedClient() *Client {

	return &Client{http.DefaultClient}
}

type Client struct {
	*http.Client
}

func (c *Client) Do(r *http.Request) (*Response, error) {
	if c.Client == nil {
		c.Client = http.DefaultClient
	}

	start := time.Now()
	httpResp, err := c.Client.Do(r)

	if err != nil {
		return nil, err
	}

	duration := time.Since(start)

	return &Response{Response: httpResp, Time: duration}, nil
}

type Response struct {
	*http.Response
	Time       time.Duration
	TotalBytes int
}

func (r *Response) String() string {
	return fmt.Sprintf("{code=%s, time=%s}", r.Status, r.Time)
}
func (r *Response) ReadAndDiscard() {
	if bytes, err :=ioutil.ReadAll(r.Body); err == nil{
		r.TotalBytes = len(bytes);
	}

	r.Body.Close()
}
