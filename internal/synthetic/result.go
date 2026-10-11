package synthetic

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sujanv/PulseGrid/internal/models"
)

type WorkflowExecutionResult struct {
	WorkflowID    string        `json:"workflow_id"`
	WorkflowName  string        `json:"workflow_name"`
	TotalDuration time.Duration `json:"total_duration"`
	StepResults   []StepResult  `json:"step_results"`
	Passed        bool          `json:"passed"`
	FailedStep    int           `json:"failed_step,omitempty"`
	ErrorMessage  string        `json:"error_message,omitempty"`
}

func (r *Runner) RunWorkflow(ctx context.Context, wf *Workflow) (*models.CheckResult, *WorkflowExecutionResult) {
	startTime := time.Now()
	execCtx := NewExecutionContext()
	var stepResults []StepResult

	passed := true
	var failureMsg string
	failedStepIdx := -1

	for i := range wf.Steps {
		step := &wf.Steps[i]
		stepRes, err := r.ExecuteStep(ctx, step, execCtx)
		stepRes.StepIndex = i + 1
		stepResults = append(stepResults, *stepRes)

		if err != nil || !stepRes.Success {
			passed = false
			failedStepIdx = i + 1
			if stepRes.Error != "" {
				failureMsg = fmt.Sprintf("Step %d (%s) failed: %s", i+1, step.Name, stepRes.Error)
			} else {
				failureMsg = fmt.Sprintf("Step %d (%s) failed with status %d", i+1, step.Name, stepRes.StatusCode)
			}
			break
		}
	}

	totalDuration := time.Since(startTime)
	execRes := &WorkflowExecutionResult{
		WorkflowID:    wf.ID,
		WorkflowName:  wf.Name,
		TotalDuration: totalDuration,
		StepResults:   stepResults,
		Passed:        passed,
		FailedStep:    failedStepIdx,
		ErrorMessage:  failureMsg,
	}

	status := models.StatusUp
	if !passed {
		status = models.StatusDown
	}

	checkRes := &models.CheckResult{
		ID:           uuid.New().String(),
		TargetID:     wf.ID,
		Timestamp:    time.Now().UTC(),
		Status:       status,
		Latency:      totalDuration,
		ErrorMessage: failureMsg,
		Breakdown: models.LatencyBreakdown{
			Total: totalDuration,
		},
	}

	return checkRes, execRes
}

type SyntheticChecker struct {
	runner *Runner
}

func NewSyntheticChecker(timeout time.Duration) *SyntheticChecker {
	return &SyntheticChecker{
		runner: NewRunner(timeout),
	}
}

func (c *SyntheticChecker) Check(ctx context.Context, target *models.Target) (*models.CheckResult, error) {
	var wf Workflow
	if err := json.Unmarshal([]byte(target.WorkflowJSON), &wf); err != nil {
		return nil, fmt.Errorf("invalid synthetic workflow JSON: %w", err)
	}
	wf.ID = target.ID
	wf.Name = target.Name

	checkRes, _ := c.runner.RunWorkflow(ctx, &wf)
	return checkRes, nil
}
