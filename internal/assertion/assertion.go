package assertion

import (
	"net/http"
)

type RuleType string

const (
	RuleJSONPath RuleType = "JSONPATH"
	RuleRegex    RuleType = "REGEX"
	RuleHeader   RuleType = "HEADER"
)

type AssertionRule struct {
	Type     RuleType `json:"type"`
	Property string   `json:"property"` // JSONPath expression, Header key, or Body
	Operator Operator `json:"operator"` // equals, contains, matches
	Expected string   `json:"expected"`
}

type Engine struct {
	jsonPath *JSONPathEvaluator
	regex    *RegexEvaluator
}

func NewEngine() *Engine {
	return &Engine{
		jsonPath: NewJSONPathEvaluator(),
		regex:    NewRegexEvaluator(),
	}
}

func (e *Engine) EvaluateAll(body []byte, headers http.Header, rules []AssertionRule) (bool, string) {
	for _, rule := range rules {
		switch rule.Type {
		case RuleJSONPath:
			ok, errMsg, err := e.jsonPath.Evaluate(body, rule.Property, rule.Operator, rule.Expected)
			if err != nil || !ok {
				if errMsg != "" {
					return false, errMsg
				}
				return false, "JSONPath assertion failed"
			}
		case RuleRegex:
			matched, err := e.regex.Matches(rule.Expected, string(body))
			if err != nil || !matched {
				return false, "Body regex assertion failed for pattern: " + rule.Expected
			}
		case RuleHeader:
			ok, errMsg, err := e.regex.EvaluateHeader(headers, rule.Property, rule.Expected)
			if err != nil || !ok {
				if errMsg != "" {
					return false, errMsg
				}
				return false, "Header assertion failed for " + rule.Property
			}
		}
	}
	return true, ""
}
