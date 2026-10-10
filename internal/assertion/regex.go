package assertion

import (
	"fmt"
	"net/http"
	"regexp"
)

type RegexEvaluator struct{}

func NewRegexEvaluator() *RegexEvaluator {
	return &RegexEvaluator{}
}

func (r *RegexEvaluator) Matches(pattern string, content string) (bool, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false, fmt.Errorf("invalid regex pattern '%s': %w", pattern, err)
	}
	return re.MatchString(content), nil
}

func (r *RegexEvaluator) EvaluateHeader(headers http.Header, headerName string, pattern string) (bool, string, error) {
	val := headers.Get(headerName)
	if val == "" {
		return false, fmt.Sprintf("header '%s' is missing or empty", headerName), nil
	}
	matched, err := r.Matches(pattern, val)
	if err != nil {
		return false, "", err
	}
	if !matched {
		return false, fmt.Sprintf("header '%s' value '%s' does not match pattern '%s'", headerName, val, pattern), nil
	}
	return true, "", nil
}
