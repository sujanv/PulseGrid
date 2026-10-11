package checker

import (
	"net/http"
	"strings"
)

type SecurityReport struct {
	Grade       string            `json:"grade"` // A+, A, B, C, F
	Score       int               `json:"score"` // 0 - 100
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

	// 1. Strict-Transport-Security (HSTS)
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

	// 2. Content-Security-Policy (CSP)
	checkHeader("Content-Security-Policy", "CSP protects against XSS", 25)

	// 3. X-Frame-Options
	checkHeader("X-Frame-Options", "Protects against clickjacking", 15)

	// 4. X-Content-Type-Options
	checkHeader("X-Content-Type-Options", "Prevents MIME sniffing", 15)

	// 5. Referrer-Policy
	checkHeader("Referrer-Policy", "Controls referrer disclosure", 10)

	// 6. Permissions-Policy
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
