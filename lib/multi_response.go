package scurl

import (
	"fmt"
	"time"
)

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

	if len(c.Responses) == 0 {
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
func (c *MultiResponse) TotalBites() uint64 {
	var total = uint64(0)

	if c.Responses == nil {
		return 0
	}

	for _, r := range c.Responses {
		total += uint64(r.TotalBytes)
	}

	return total
}
