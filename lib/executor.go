package scurl

import (
	"sync"
	"time"
)

var DefaultRate = &Rate{Freq: 50, Per: 1 * time.Second}

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

func VerboseOpt(verbose bool) func(*ConcurrentClient) {
	return func(client *ConcurrentClient) {
		l := &logger{verbose: verbose}
		client.logger = l
		client.httpClient.logger = l
	}
}

type ConcurrentClient struct {
	logger     *logger
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

func (c *ConcurrentClient) DoReq(t *Target) <-chan *Response {
	if c.rate == nil {
		c.rate = DefaultRate
	}
	if c.logger == nil {
		c.logger = mutedLogger
	}

	workers := sync.WaitGroup{}
	respCh := make(chan *Response)

	c.logger.debug("duration:", c.du)
	c.logger.debug("rate:", c.rate)
	c.logger.debug("fanOut:", c.fanOut)
	c.logger.debug(">", t.Method, t.URL)
	if len(t.Header) != 0 {
		c.logger.debug(">")
		for k, v := range t.Header {
			c.logger.debug("> ", k, ":", v)
		}
	}
	if t.Body != nil {
		c.logger.debug(">")
		c.logger.debug(t.Body)
	}

	for i := 0; i < c.fanOut; i++ {
		atk := attacker{stopper: c.stopper, logger: c.logger}
		c.attackers = append(c.attackers, atk)

		workers.Add(1)

		go func() {
			defer workers.Done()
			for resp := range atk.Attack(t, c.rate, c.du) {
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
