package scurl

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAttackerRate(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
	)
	defer server.Close()

	atk := &attacker{}

	req, _ := NewTarget(server.URL)
	rate := &Rate{Freq: 100, Per: time.Second}

	hits := 0
	for range atk.Attack(req, rate, time.Second) {
		hits++
	}

	if hits != rate.Freq {
		t.Fatalf("got: %v, want: %v", hits, rate.Freq)
	}

}

func TestStopAttackWhenOrdered(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		}),
	)

	req, _ := NewTarget(server.URL)
	rate := &Rate{Freq: 1, Per: time.Second}

	atk := &attacker{}

	result := atk.Attack(req, rate, 1*time.Hour)

	go func() {
		atk.Stop()
	}()

	for range result {
		// if the attack does not stop this test will never finish
	}
}
