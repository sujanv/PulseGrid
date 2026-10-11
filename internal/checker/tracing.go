package checker

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
)

type Tracer struct{}

func NewTracer() *Tracer {
	return &Tracer{}
}

func (t *Tracer) InjectTraceHeaders(req *http.Request) {
	traceID := make([]byte, 16)
	parentID := make([]byte, 8)
	_, _ = rand.Read(traceID)
	_, _ = rand.Read(parentID)

	// W3C traceparent format: version(00)-traceid(32 hex)-parentid(16 hex)-traceflags(01)
	traceparent := fmt.Sprintf("00-%s-%s-01", hex.EncodeToString(traceID), hex.EncodeToString(parentID))

	req.Header.Set("traceparent", traceparent)
	req.Header.Set("tracestate", "pulsegrid=probe")
	req.Header.Set("X-PulseGrid-Trace", hex.EncodeToString(traceID))
}
