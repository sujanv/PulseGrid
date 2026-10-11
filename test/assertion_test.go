package test

import (
	"net/http"
	"testing"

	"github.com/sujanv/PulseGrid/internal/assertion"
)

func TestJSONPathEvaluator(t *testing.T) {
	eval := assertion.NewJSONPathEvaluator()
	jsonData := []byte(`{
		"status": "healthy",
		"version": 2,
		"metrics": {
			"load": 0.45,
			"workers": 10
		},
		"nodes": ["node-1", "node-2"]
	}`)

	// 1. String Equality
	ok, msg, err := eval.Evaluate(jsonData, "$.status", assertion.OpEquals, "healthy")
	if err != nil || !ok {
		t.Errorf("expected status == healthy, got error: %v, msg: %s", err, msg)
	}

	// 2. Numeric Greater Than
	ok, _, _ = eval.Evaluate(jsonData, "$.metrics.workers", assertion.OpGreaterThan, "5")
	if !ok {
		t.Errorf("expected metrics.workers > 5")
	}

	// 3. Array Index
	ok, _, _ = eval.Evaluate(jsonData, "$.nodes[0]", assertion.OpEquals, "node-1")
	if !ok {
		t.Errorf("expected nodes[0] == node-1")
	}

	// 4. Exists
	ok, _, _ = eval.Evaluate(jsonData, "$.metrics.load", assertion.OpExists, "true")
	if !ok {
		t.Errorf("expected metrics.load to exist")
	}
}

func TestRegexEvaluator(t *testing.T) {
	eval := assertion.NewRegexEvaluator()
	body := "API server v2.4.1 running normally"

	matched, err := eval.Matches("v2\.[0-9]+\.[0-9]+", body)
	if err != nil || !matched {
		t.Errorf("expected regex match on version pattern")
	}

	headers := http.Header{}
	headers.Set("Content-Type", "application/json; charset=utf-8")

	ok, _, _ := eval.EvaluateHeader(headers, "Content-Type", "application/json")
	if !ok {
		t.Errorf("expected header match")
	}
}
