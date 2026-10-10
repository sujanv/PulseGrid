package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sujanv/PulseGrid/internal/checker"
	"github.com/sujanv/PulseGrid/internal/models"
)

func TestHTTPCheckerSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("pong-response"))
	}))
	defer ts.Close()

	c := checker.NewHTTPChecker(5 * time.Second)
	target := &models.Target{
		Name:                "Test-HTTP",
		URL:                 ts.URL,
		Type:                models.TargetHTTP,
		Timeout:             5 * time.Second,
		ExpectedStatusCodes: []int{200},
		ExpectedKeyword:     "pong",
	}

	res, err := c.Check(context.Background(), target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Status != models.StatusUp {
		t.Errorf("expected StatusUp, got: %s (err: %s)", res.Status, res.ErrorMessage)
	}
	if res.StatusCode != 200 {
		t.Errorf("expected 200, got: %d", res.StatusCode)
	}
}

func TestHTTPCheckerKeywordFailure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("unexpected-data"))
	}))
	defer ts.Close()

	c := checker.NewHTTPChecker(5 * time.Second)
	target := &models.Target{
		Name:                "Test-HTTP",
		URL:                 ts.URL,
		Type:                models.TargetHTTP,
		Timeout:             5 * time.Second,
		ExpectedStatusCodes: []int{200},
		ExpectedKeyword:     "missing-keyword",
	}

	res, err := c.Check(context.Background(), target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Status != models.StatusDown {
		t.Errorf("expected StatusDown due to missing keyword, got: %s", res.Status)
	}
}
