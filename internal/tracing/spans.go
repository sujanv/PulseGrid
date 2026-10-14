package tracing

import (
	"github.com/sujanv/PulseGrid/internal/models"
)

func CreatePhaseSpans(tracer *Tracer, target *models.Target, res *models.CheckResult) *Span {
	root := tracer.StartSpan(nil, "health_check:"+target.Name, "")
	root.SetAttribute("target.id", target.ID)
	root.SetAttribute("target.url", target.URL)
	root.SetAttribute("target.type", string(target.Type))
	root.SetAttribute("http.status_code", res.StatusCode)
	root.SetAttribute("check.status", string(res.Status))

	if res.Breakdown.DNSLookup > 0 {
		root.AddEvent("dns_lookup_complete")
		root.SetAttribute("timing.dns_ms", float64(res.Breakdown.DNSLookup.Milliseconds()))
	}
	if res.Breakdown.TCPConnect > 0 {
		root.AddEvent("tcp_connect_complete")
		root.SetAttribute("timing.tcp_ms", float64(res.Breakdown.TCPConnect.Milliseconds()))
	}
	if res.Breakdown.TLSHandshake > 0 {
		root.AddEvent("tls_handshake_complete")
		root.SetAttribute("timing.tls_ms", float64(res.Breakdown.TLSHandshake.Milliseconds()))
	}
	if res.Breakdown.TTFB > 0 {
		root.AddEvent("ttfb_received")
		root.SetAttribute("timing.ttfb_ms", float64(res.Breakdown.TTFB.Milliseconds()))
	}

	root.End()
	tracer.RecordSpan(target.ID, root)
	return root
}
