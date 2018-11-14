package scurl

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type stopper struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
}

func (s *stopper) Stop() {
	s.cancelFunc()
}

func (s *stopper) Done() <-chan struct{} {
	return s.ctx.Done()
}

func NewStopper() *stopper {
	ctx, cancel := context.WithCancel(context.Background())

	return &stopper{
		ctx:        ctx,
		cancelFunc: cancel,
	}
}

type Rate struct {
	Freq int           // Frequency (number of occurrences) per ...
	Per  time.Duration // Time unit, usually 1s
}

func (r *Rate) Interval() time.Duration {
	return time.Duration(r.Per.Nanoseconds() / int64(r.Freq))
}

func (r *Rate) Hits(duration time.Duration) uint64 {
	return uint64(duration.Nanoseconds() / r.Interval().Nanoseconds())
}

func (r *Rate) RemainingSince(start time.Time, iteration uint64) time.Duration {
	timeSinceStart := start.Add(r.Interval() * time.Duration(iteration))

	return timeSinceStart.Sub(time.Now())
}

func (r *Rate) String() string {
	return fmt.Sprintf("%d/%v", r.Freq, r.Per)
}

func (r *Rate) IsZero() bool {
	return r.Freq == 0 || r.Per == 0
}

type attacker struct {
	workers int
	client  *Client
	stopper *stopper
}

func (a *attacker) Attack(t *Target, r *Rate, du time.Duration) <-chan *Response {
	workers := sync.WaitGroup{}
	results := make(chan *Response)
	ticks := make(chan uint64)

	if a.stopper == nil {
		a.stopper = NewStopper()
	}

	for i := 0; i < a.workers; i++ {
		workers.Add(1)
		go a.attack(t, ticks, &workers, results)
	}

	go func() {
		defer close(results)
		defer workers.Wait()
		defer close(ticks)

		hits := r.Hits(du)
		count := uint64(0)
		began := time.Now()

		for {

			time.Sleep(r.RemainingSince(began, count))

			select {

			case ticks <- count:
				if count++; count == hits {
					return
				}

			case _, ok := <-a.Done():
				if !ok {
					return
				}

			default:
				workers.Add(1)
				go a.attack(t, ticks, &workers, results)
			}
		}

	}()

	return results
}

func (a *attacker) attack(t *Target, ticks <-chan uint64, workers *sync.WaitGroup, result chan *Response) {
	defer workers.Done()

	for {
		select {
		case _, ok := <-ticks:
			if !ok {
				return
			}

			resp := a.hit(t)
			if resp != nil {
				result <- resp
			}

		case _, ok := <-a.Done():
			if !ok {
				return
			}
		}
	}
}

func (a *attacker) hit(t *Target) *Response {
	if a.client == nil {
		a.client = &Client{}
	}

	req, err := t.RequestWithContext(a.stopper.ctx)
	if err != nil {
		a.Stop()
		return nil
	}

	response, e := a.client.Do(req)

	if e != nil {
		a.Stop()
		return response
	}

	return response
}

func (a *attacker) Stop() {

	a.stopper.Stop()
}

func (a *attacker) Done() <-chan struct{} {

	return a.stopper.Done()
}
