package synthetic

import (
	"net/http"
	"time"
	"github.com/sujanv/PulseGrid/internal/assertion"
)

type ExtractionRule struct {
	VariableName string                 `json:"variable_name"`
	Type         assertion.RuleType     `json:"type"` // JSONPATH or HEADER
	Property     string                 `json:"property"`
}

type Step struct {
	Name        string                    `json:"name"`
	Method      string                    `json:"method"`
	URL         string                    `json:"url"`
	Headers     map[string]string         `json:"headers,omitempty"`
	Body        string                    `json:"body,omitempty"`
	Timeout     time.Duration             `json:"timeout"`
	Assertions  []assertion.AssertionRule `json:"assertions,omitempty"`
	Extractions []ExtractionRule          `json:"extractions,omitempty"`
}

type Workflow struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Steps       []Step        `json:"steps"`
	Timeout     time.Duration `json:"timeout"`
}

type Runner struct {
	client          *http.Client
	assertionEngine *assertion.Engine
}

func NewRunner(timeout time.Duration) *Runner {
	return &Runner{
		client: &http.Client{
			Timeout: timeout,
		},
		assertionEngine: assertion.NewEngine(),
	}
}
