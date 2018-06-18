package scurl

import (
	"time"
	"context"
	"net/http"
	"bytes"
	"io/ioutil"
	"fmt"
)

func NewConcurrentClient(fanOut int) *ConcurrentClient {
	ctx, cancel := context.WithCancel(context.Background())

	return &ConcurrentClient{
		fanOut:     fanOut,
		ctx:        ctx,
		cancel:     cancel,
		httpClient: NewTimedClient(),
	}
}

type ConcurrentClient struct {
	fanOut     int
	ctx        context.Context
	cancel     context.CancelFunc
	httpClient *Client
}

func (c *ConcurrentClient) Do(req *http.Request) (*MultiResponse, error) {

	body, err := copyBody(req)
	if err != nil {
		return nil, err
	}

	respCh := make(chan *Response)

	start := time.Now()
	for i := 0; i < c.fanOut; i++ {
		go c.doReq(copyReq(req, body).WithContext(c.ctx), respCh)
	}

	responses := make([]*Response, 0, c.fanOut)

	respFunc := func(resp []*Response) *MultiResponse {
		return &MultiResponse{Responses: responses, Trips: len(responses), StartTime: start}
	}

	for i := 0; i < c.fanOut; i++ {
		select {
		case resp := <-respCh:
			responses = append(responses, resp)
		case <-c.ctx.Done():
			close(respCh)
			return respFunc(responses), nil
		}
	}

	return respFunc(responses), nil
}

func (c *ConcurrentClient) Stop() {
	c.cancel()
}

func (c *ConcurrentClient) DoReq(req *http.Request) <-chan *Response {
	responses := make(chan *Response, c.fanOut)

	body, err := copyBody(req)
	if err != nil {
		panic(err)
	}

	go closeChanWhenDone(c.ctx, responses) // probably this can be done without starting a go routine

	for i := 0; i < c.fanOut; i++ {
		go c.doReq(copyReq(req, body).WithContext(c.ctx), responses)
	}

	return responses
}

func closeChanWhenDone(ctx context.Context, responses chan *Response) {
	select {
	case <-ctx.Done():
		close(responses)
	}
}

func (c *ConcurrentClient) doReq(req *http.Request, respCh chan<- *Response) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.cancel()
		return
	}

	respCh <- resp
}

type MultiResponse struct {
	Responses []*Response
	Trips     int
	StartTime time.Time
}

func (c *MultiResponse) TotalTime() time.Duration {
	return time.Since(c.StartTime)
}

func (c *MultiResponse) AvrTime() time.Duration {

	var avr int64 = 0
	for _, r := range c.Responses {
		avr += r.Time.Nanoseconds()
	}

	if (len(c.Responses) == 0) {
		return 0
	}

	avr = avr / int64(len(c.Responses))

	return time.Duration(avr)

}

func (c *MultiResponse) Slowest() *Response {

	if len(c.Responses) == 0 {
		return nil
	}
	slowest := c.Responses[0]

	for _, resp := range c.Responses {

		if resp.Time > slowest.Time {
			slowest = resp
		}
	}

	return slowest
}

func (c *MultiResponse) Fastest() *Response {
	if len(c.Responses) == 0 {
		return nil
	}
	slowest := c.Responses[0]

	for _, resp := range c.Responses {

		if resp.Time < slowest.Time {
			slowest = resp
		}
	}

	return slowest
}

func (c *MultiResponse) Add(r *Response) {
	if c.Responses == nil {
		c.Responses = make([]*Response, 0)
	}

	c.Responses = append(c.Responses, r)
	c.Trips += 1
}

func (c *MultiResponse) Empty() bool {
	if c.Responses == nil {
		return true;
	}

	return len(c.Responses) == 0
}

func (c *MultiResponse) Close() {
	if c.Responses != nil {
		for _, r := range c.Responses {
			r.Body.Close()
		}
	}
}

func (c *MultiResponse) String() string {
	return fmt.Sprintf("{time=%s, trips=%d, resp=%s}", c.TotalTime(), c.Trips, c.Responses)
}

func (c *MultiResponse) StatusMap() map[int][]*Response {
	statMap := make(map[int][]*Response)

	for _, v := range c.Responses {
		statMap[v.StatusCode] = append(statMap[v.StatusCode], v)
	}

	return statMap
}

func copyBody(req *http.Request) ([]byte, error) {

	if req.Body == nil {
		return nil, nil
	}

	allBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	req.Body = ioutil.NopCloser(bytes.NewReader(allBytes))

	return allBytes, nil
}

func copyReq(r *http.Request, body []byte) *http.Request {

	req, err := http.NewRequest(r.Method, r.URL.String(), bytes.NewReader(body))
	if err != nil {
		panic(err)
	}

	for k, vs := range r.Header {
		req.Header[k] = make([]string, len(vs))
		copy(req.Header[k], vs)
	}

	if host := req.Header.Get("Host"); host != "" {
		req.Host = host
	}

	return req
}
