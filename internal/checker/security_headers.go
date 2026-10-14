package checker

import (
	"net/http"
	"strings"
)

type SecurityReport struct {
	Grade       string            `json:"grade"`
	Score       int               `json:"score"`
	Headers     map[string]string `json:"headers"`
	Missing     []string          `json:"missing"`
	Warnings    []string          `json:"warnings"`
}

type SecurityAuditor struct{}

func NewSecurityAuditor() *SecurityAuditor {
	return &SecurityAuditor{}
}

func (a *SecurityAuditor) Audit(headers http.Header) *SecurityReport {
	score := 100
	var missing []string
	var warnings []string
	captured := make(map[string]string)

	checkHeader := func(key, desc string, weight int) {
		val := headers.Get(key)
		if val == "" {
			score -= weight
			missing = append(missing, key)
		} else {
			captured[key] = val
		}
	}

	hsts := headers.Get("Strict-Transport-Security")
	if hsts == "" {
		score -= 25
		missing = append(missing, "Strict-Transport-Security")
	} else {
		captured["Strict-Transport-Security"] = hsts
		if !strings.Contains(hsts, "includeSubDomains") {
			warnings = append(warnings, "HSTS missing includeSubDomains directive")
		}
	}

	checkHeader("Content-Security-Policy", "CSP protects against XSS", 25)
	checkHeader("X-Frame-Options", "Protects against clickjacking", 15)
	checkHeader("X-Content-Type-Options", "Prevents MIME sniffing", 15)
	checkHeader("Referrer-Policy", "Controls referrer disclosure", 10)
	checkHeader("Permissions-Policy", "Controls browser features", 10)

	if score < 0 {
		score = 0
	}

	grade := "F"
	switch {
	case score >= 95:
		grade = "A+"
	case score >= 85:
		grade = "A"
	case score >= 70:
		grade = "B"
	case score >= 50:
		grade = "C"
	}

	return &SecurityReport{
		Grade:    grade,
		Score:    score,
		Headers:  captured,
		Missing:  missing,
		Warnings: warnings,
	}
}
