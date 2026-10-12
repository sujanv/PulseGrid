package test

import (
	"net/http"
	"testing"

	"github.com/sujanv/PulseGrid/internal/checker"
)

func TestSecurityAuditorScoring(t *testing.T) {
	auditor := checker.NewSecurityAuditor()

	// 1. Excellent Headers (A+)
	hA := http.Header{}
	hA.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
	hA.Set("Content-Security-Policy", "default-src 'self'")
	hA.Set("X-Frame-Options", "DENY")
	hA.Set("X-Content-Type-Options", "nosniff")
	hA.Set("Referrer-Policy", "strict-origin-when-cross-origin")
	hA.Set("Permissions-Policy", "geolocation=()")

	repA := auditor.Audit(hA)
	if repA.Grade != "A+" || repA.Score < 95 {
		t.Errorf("expected A+ grade, got %s (score: %d)", repA.Grade, repA.Score)
	}

	// 2. Missing Everything (F)
	hF := http.Header{}
	repF := auditor.Audit(hF)
	if repF.Grade != "F" || repF.Score != 0 {
		t.Errorf("expected F grade, got %s (score: %d)", repF.Grade, repF.Score)
	}
}
