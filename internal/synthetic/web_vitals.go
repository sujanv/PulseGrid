package synthetic

import (
	"time"
)

type MetricRating string

const (
	RatingGood             MetricRating = "GOOD"
	RatingNeedsImprovement MetricRating = "NEEDS_IMPROVEMENT"
	RatingPoor             MetricRating = "POOR"
)

type CoreWebVitalsReport struct {
	URL         string        `json:"url"`
	LCP         time.Duration `json:"lcp"`          // Largest Contentful Paint (<2.5s)
	LCPRating   MetricRating  `json:"lcp_rating"`
	FID         time.Duration `json:"fid"`          // First Input Delay (<100ms)
	FIDRating   MetricRating  `json:"fid_rating"`
	CLS         float64       `json:"cls"`          // Cumulative Layout Shift (<0.1)
	CLSRating   MetricRating  `json:"cls_rating"`
	TTFB        time.Duration `json:"ttfb"`         // Time to First Byte (<800ms)
	OverallPass bool          `json:"overall_pass"`
	EvaluatedAt time.Time     `json:"evaluated_at"`
}

type WebVitalsCalculator struct{}

func NewWebVitalsCalculator() *WebVitalsCalculator {
	return &WebVitalsCalculator{}
}

func (c *WebVitalsCalculator) Calculate(url string, lcp, fid, ttfb time.Duration, cls float64) CoreWebVitalsReport {
	// 1. Evaluate LCP
	lcpRating := RatingGood
	if lcp > 4*time.Second {
		lcpRating = RatingPoor
	} else if lcp > 2500*time.Millisecond {
		lcpRating = RatingNeedsImprovement
	}

	// 2. Evaluate FID / INP
	fidRating := RatingGood
	if fid > 300*time.Millisecond {
		fidRating = RatingPoor
	} else if fid > 100*time.Millisecond {
		fidRating = RatingNeedsImprovement
	}

	// 3. Evaluate CLS
	clsRating := RatingGood
	if cls > 0.25 {
		clsRating = RatingPoor
	} else if cls > 0.10 {
		clsRating = RatingNeedsImprovement
	}

	overallPass := lcpRating == RatingGood && fidRating == RatingGood && clsRating == RatingGood

	return CoreWebVitalsReport{
		URL:         url,
		LCP:         lcp,
		LCPRating:   lcpRating,
		FID:         fid,
		FIDRating:   fidRating,
		CLS:         cls,
		CLSRating:   clsRating,
		TTFB:        ttfb,
		OverallPass: overallPass,
		EvaluatedAt: time.Now().UTC(),
	}
}
