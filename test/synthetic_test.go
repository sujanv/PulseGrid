package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sujanv/PulseGrid/internal/assertion"
	"github.com/sujanv/PulseGrid/internal/synthetic"
)

func TestSyntheticWorkflowExecution(t *testing.T) {
	// Mock server that requires token from step 1 for step 2
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"token": "secret-session-token", "status": "authenticated"}`))
			return
		}

		if r.URL.Path == "/profile" {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "Bearer secret-session-token" {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error": "unauthorized"}`))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"user": "admin", "active": true}`))
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	runner := synthetic.NewRunner(5 * time.Second)
	wf := &synthetic.Workflow{
		ID:   "wf-auth-test",
		Name: "Auth & Profile Flow",
		Steps: []synthetic.Step{
			{
				Name:   "Login Step",
				Method: "POST",
				URL:    ts.URL + "/login",
				Assertions: []assertion.AssertionRule{
					{
						Type:     assertion.RuleJSONPath,
						Property: "$.status",
						Operator: assertion.OpEquals,
						Expected: "authenticated",
					},
				},
				Extractions: []synthetic.ExtractionRule{
					{
						VariableName: "authToken",
						Type:         assertion.RuleJSONPath,
						Property:     "$.token",
					},
				},
			},
			{
				Name:   "Profile Step",
				Method: "GET",
				URL:    ts.URL + "/profile",
				Headers: map[string]string{
					"Authorization": "Bearer {{authToken}}",
				},
				Assertions: []assertion.AssertionRule{
					{
						Type:     assertion.RuleJSONPath,
						Property: "$.user",
						Operator: assertion.OpEquals,
						Expected: "admin",
					},
				},
			},
		},
	}

	checkRes, execRes := runner.RunWorkflow(context.Background(), wf)
	if !execRes.Passed {
		t.Fatalf("expected workflow to pass, but failed: %s", execRes.ErrorMessage)
	}

	if len(execRes.StepResults) != 2 {
		t.Errorf("expected 2 step results, got %d", len(execRes.StepResults))
	}

	if checkRes.Status != "UP" {
		t.Errorf("expected check result status UP, got %s", checkRes.Status)
	}
}
