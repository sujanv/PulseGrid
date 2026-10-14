package test

import (
	"testing"
	"time"

	"github.com/sujanv/PulseGrid/internal/tracing"
)

func TestTracerSpanCreation(t *testing.T) {
	tr := tracing.NewTracer("pulsegrid-test")
	span := tr.StartSpan(nil, "http_probe", "")

	time.Sleep(5 * time.Millisecond)
	span.SetAttribute("http.status_code", 200)
	span.AddEvent("dns_done")
	span.End()

	if span.DurationMs <= 0 {
		t.Errorf("expected positive duration")
	}

	tr.RecordSpan("tgt-trace", span)
	spans := tr.GetSpans("tgt-trace")
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
}

func TestTraceparentFormatting(t *testing.T) {
	tc := &tracing.TraceContext{
		TraceID:  "4bf92f3577b34da6a3ce929d0e0e4736",
		ParentID: "00f067aa0ba902b7",
		Sampled:  true,
	}

	header := tracing.FormatTraceparent(tc)
	parsed, err := tracing.ParseTraceparent(header)
	if err != nil {
		t.Fatalf("failed to parse formatted traceparent: %v", err)
	}

	if parsed.TraceID != tc.TraceID || parsed.ParentID != tc.ParentID || !parsed.Sampled {
		t.Errorf("roundtrip mismatch")
	}
}
