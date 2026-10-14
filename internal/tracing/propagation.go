package tracing

import (
	"fmt"
	"strings"
)

type TraceContext struct {
	TraceID  string
	ParentID string
	Sampled  bool
}

func ParseTraceparent(header string) (*TraceContext, error) {
	parts := strings.Split(header, "-")
	if len(parts) < 4 {
		return nil, fmt.Errorf("invalid traceparent format")
	}
	return &TraceContext{
		TraceID:  parts[1],
		ParentID: parts[2],
		Sampled:  parts[3] == "01",
	}, nil
}

func FormatTraceparent(tc *TraceContext) string {
	sampledStr := "00"
	if tc.Sampled {
		sampledStr = "01"
	}
	return fmt.Sprintf("00-%s-%s-%s", tc.TraceID, tc.ParentID, sampledStr)
}
