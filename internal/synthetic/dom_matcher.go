package synthetic

import (
	"fmt"
	"regexp"
	"strings"
)

type DOMMatcher struct{}

func NewDOMMatcher() *DOMMatcher {
	return &DOMMatcher{}
}

func (m *DOMMatcher) ContainsText(htmlContent, searchStr string) bool {
	return strings.Contains(htmlContent, searchStr)
}

func (m *DOMMatcher) MatchesRegex(htmlContent, regexPattern string) (bool, error) {
	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return false, fmt.Errorf("invalid regex %q: %w", regexPattern, err)
	}
	return re.MatchString(htmlContent), nil
}

func (m *DOMMatcher) ExtractTagContent(htmlContent, tagName string) []string {
	pattern := fmt.Sprintf(`(?i)<%s[^>]*>(.*?)</%s>`, tagName, tagName)
	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(htmlContent, -1)

	var results []string
	for _, match := range matches {
		if len(match) > 1 {
			results = append(results, strings.TrimSpace(match[1]))
		}
	}
	return results
}

func (m *DOMMatcher) HasElementWithID(htmlContent, id string) bool {
	pattern := fmt.Sprintf(`(?i)id=["']%s["']`, regexp.QuoteMeta(id))
	re := regexp.MustCompile(pattern)
	return re.MatchString(htmlContent)
}

func (m *DOMMatcher) HasElementWithClass(htmlContent, className string) bool {
	pattern := fmt.Sprintf(`(?i)class=["'][^"']*\b%s\b[^"']*["']`, regexp.QuoteMeta(className))
	re := regexp.MustCompile(pattern)
	return re.MatchString(htmlContent)
}
