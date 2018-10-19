package scurl

import (
	"net/http"
	"sync"
	"time"
)

var DefaultRate = &Rate{Freq: 50, Per: 1 * time.Second}

func FanOutOpt(num int) func(*ConcurrentClient) {
	return func(client *ConcurrentClient) {
		if num < 1 {
			num = 1
		}

		client.fanOut = num
	}
}

func RateOpt(rate *Rate) func(*ConcurrentClient) {
	return func(client *ConcurrentClient) {
		if rate == nil || rate.IsZero() {
			rate = DefaultRate
		}

		client.rate = rate
	}
}

func DurationOpt(du time.Duration) func(*ConcurrentClient) {
	return func(client *ConcurrentClient) {
		client.du = du
	}
}

func NewConcurrentClient(opts ...func(*ConcurrentClient)) *ConcurrentClient {

	client := &ConcurrentClient{
		httpClient: NewTimedClient(),
		stopper:    NewStopper(),
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

type ConcurrentClient struct {
	fanOut     int
	rate       *Rate
	du         time.Duration
	httpClient *Client
	attackers  []attacker
	stopper    *stopper
}

func (c *ConcurrentClient) Stop() {
	c.stopper.Stop()
}

func (c *ConcurrentClient) DoReq(req *http.Request) <-chan *Response {

	if c.rate == nil {
		c.rate = DefaultRate
	}

	workers := sync.WaitGroup{}
	respCh := make(chan *Response)

	for i := 0; i < c.fanOut; i++ {

		atk := attacker{stopper: c.stopper}
		c.attackers = append(c.attackers, atk)
		reqWithContext := copyReq(req).WithContext(c.stopper.ctx)

		workers.Add(1)

		go func() {
			defer workers.Done()

			for resp := range atk.Attack(reqWithContext, c.rate, c.du) {
				respCh <- resp
			}
		}()
	}

	go func() {
		defer close(respCh)

		workers.Wait()
	}()

	return respCh
}
