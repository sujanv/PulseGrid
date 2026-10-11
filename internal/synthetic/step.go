package synthetic

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sujanv/PulseGrid/internal/assertion"
)

type StepResult struct {
	StepIndex  int           `json:"step_index"`
	StepName   string        `json:"step_name"`
	Duration   time.Duration `json:"duration"`
	StatusCode int           `json:"status_code"`
	Success    bool          `json:"success"`
	Error      string        `json:"error,omitempty"`
}

func (r *Runner) ExecuteStep(ctx context.Context, step *Step, execCtx *ExecutionContext) (*StepResult, error) {
	startTime := time.Now()

	resolvedURL := execCtx.Interpolate(step.URL)
	resolvedBody := execCtx.Interpolate(step.Body)

	var bodyReader io.Reader
	if resolvedBody != "" {
		bodyReader = bytes.NewBufferString(resolvedBody)
	}

	req, err := http.NewRequestWithContext(ctx, step.Method, resolvedURL, bodyReader)
	if err != nil {
		return &StepResult{
			StepName: step.Name,
			Success:  false,
			Error:    fmt.Sprintf("failed to create request: %v", err),
			Duration: time.Since(startTime),
		}, err
	}

	for k, v := range step.Headers {
		req.Header.Set(k, execCtx.Interpolate(v))
	}
	req.Header.Set("User-Agent", "PulseGrid-Synthetic-Agent/1.0")

	resp, err := r.client.Do(req)
	if err != nil {
		return &StepResult{
			StepName: step.Name,
			Success:  false,
			Error:    fmt.Sprintf("HTTP request error: %v", err),
			Duration: time.Since(startTime),
		}, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return &StepResult{
			StepName:   step.Name,
			StatusCode: resp.StatusCode,
			Success:    false,
			Error:      fmt.Sprintf("failed to read body: %v", err),
			Duration:   time.Since(startTime),
		}, err
	}

	// 1. Evaluate Assertions
	if len(step.Assertions) > 0 {
		ok, errMsg := r.assertionEngine.EvaluateAll(respBytes, resp.Header, step.Assertions)
		if !ok {
			return &StepResult{
				StepName:   step.Name,
				StatusCode: resp.StatusCode,
				Success:    false,
				Error:      fmt.Sprintf("assertion failed: %s", errMsg),
				Duration:   time.Since(startTime),
			}, nil
		}
	}

	// 2. Extract Variables into Context
	for _, ext := range step.Extractions {
		if ext.Type == assertion.RuleJSONPath {
			evaluator := assertion.NewJSONPathEvaluator()
			val, extractErr := evaluator.Extract(respBytes, ext.Property)
			if extractErr == nil && val != nil {
				execCtx.Set(ext.VariableName, fmt.Sprintf("%v", val))
			}
		} else if ext.Type == assertion.RuleHeader {
			hdrVal := resp.Header.Get(ext.Property)
			if hdrVal != "" {
				execCtx.Set(ext.VariableName, hdrVal)
			}
		}
	}

	return &StepResult{
		StepName:   step.Name,
		StatusCode: resp.StatusCode,
		Success:    true,
		Duration:   time.Since(startTime),
	}, nil
}
