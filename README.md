# ⚡ PulseGrid

> **Enterprise Distributed Synthetic & Uptime Monitoring Engine**  
> Engineered in Go 1.22+ with Bounded-Concurrency Goroutine Pools, Google SRE Error Budget Tracking, W3C HAR 1.2 Exporter, Automated Canary Analysis, Chaos Fault Injection, and Kubernetes Native Automation.

[![Go Version](https://img.shields.io/badge/go-1.22+-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/docker-ready-blue.svg)](deploy/docker-compose.yml)
[![Kubernetes Operator](https://img.shields.io/badge/k8s-operator-326ce5.svg)](deploy/helm/pulsegrid)
[![OpenTelemetry](https://img.shields.io/badge/opentelemetry-tracing-f5a800.svg)](internal/tracing)

---

## 📖 Table of Contents

- [Overview](#-overview)
- [System Architecture](#-system-architecture)
- [Key Enterprise Features](#-key-enterprise-features)
  - [High-Throughput Bounded Concurrency Worker Pool](#high-throughput-bounded-concurrency-worker-pool)
  - [Multi-Protocol Probing Engines](#multi-protocol-probing-engines)
  - [X.509 Certificate Chain & SAN Inspector](#x509-certificate-chain--san-inspector)
  - [Mutual TLS (mTLS) Zero-Trust Authentication](#mutual-tls-mtls-zero-trust-authentication)
  - [Browser-Like Multi-Step Synthetic Scenarios](#browser-like-multi-step-synthetic-scenarios)
  - [W3C HAR 1.2 HTTP Archive Exporter](#w3c-har-12-http-archive-exporter)
  - [Automated Canary Analysis (ACA) & Progressive Rollouts](#automated-canary-analysis-aca--progressive-rollouts)
  - [Google SRE Multi-Window Multi-Burn-Rate SLO Engine](#google-sre-multi-window-multi-burn-rate-slo-engine)
  - [Dynamic Adaptive Sampling](#dynamic-adaptive-sampling)
  - [Alert De-duplication & Failure Domain Correlation](#alert-de-duplication--failure-domain-correlation)
  - [Bi-directional PagerDuty v2 Sync & Chat Webhooks](#bi-directional-pagerduty-v2-sync--chat-webhooks)
  - [Geo-Matrix Consensus & Exponential Flap Dampening](#geo-matrix-consensus--exponential-flap-dampening)
  - [Chaos Engineering & Synthetic Fault Injection](#chaos-engineering--synthetic-fault-injection)
  - [Automated SRE Incident Post-Mortem & Timeline Generator](#automated-sre-incident-post-mortem--timeline-generator)
  - [Hop-by-Hop Traceroute & Anycast Route Shifts](#hop-by-hop-traceroute--anycast-route-shifts)
  - [Kubernetes Operator & Helm Deployment](#kubernetes-operator--helm-deployment)
- [Getting Started](#-getting-started)
- [CLI Reference](#-cli-reference)
- [REST API Matrix](#-rest-api-matrix)
- [Testing & Quality Assurance](#-testing--quality-assurance)
- [License](#-license)

---

## 🔭 Overview

**PulseGrid** is an enterprise synthetic and uptime monitoring service built in Go. It empowers Site Reliability Engineering (SRE) and platform teams to schedule, probe, diagnose, and auto-remediate thousands of distributed targets with strictly bounded resource usage.

PulseGrid eliminates monitoring blind spots by combining multi-region quorum consensus, synthetic browser user journeys with cookie persistence, W3C HAR 1.2 exporting, deep X.509 certificate authority inspection, mTLS zero-trust probing, hop-by-hop network path diagnosis, and Google SRE multi-window burn rate alerts.

---

## 🏛 System Architecture

```mermaid
flowchart TD
    subgraph Ingestion & Scheduling
        Cron[Distributed Scheduler / Cron] --> Q[Redis Priority Queue]
        K8s[Kubernetes Ingress & CRD Watcher] -->|Sync Targets| PG[(PostgreSQL)]
        AM[Prometheus Alertmanager Hook] --> Correlator[Infra Correlator]
    end

    subgraph Probing Engine [Bounded Concurrency Worker Pool]
        Q --> WP[Worker Pool]
        WP --> ChkHTTP[HTTP/HTTPS Assertion]
        WP --> ChkGRPC[gRPC Health Check Protocol]
        WP --> ChkMTLS[mTLS Zero-Trust Probing]
        WP --> ChkCert[X.509 SAN & Expiration Inspector]
        WP --> ChkTCP[TCP / Socket Probe]
        WP --> ChkUDP[UDP Reachability Probe]
        WP --> ChkDNS[DNS & DNSSEC Signature]
        WP --> ChkTLS[TLS & OCSP Stapling]
        WP --> ChkScen[Synthetic Scenario Runner]
        WP --> ChkTrace[Hop-by-Hop Traceroute]
        WP --> Chaos[Chaos Fault Injector]
    end

    subgraph Consensus & Analysis
        WP --> GeoQuorum[Geo-Matrix Consensus]
        WP --> FlapDampener[Flap Dampener - Hysteresis]
        WP --> Sampler[Adaptive Sampling Engine]
        WP --> OTel[OpenTelemetry Span Collector]
        WP --> SLO[Google SRE Burn Rate Evaluator]
        WP --> Canary[Automated Canary Analyzer]
        WP --> RCA[RCA & Post-Mortem Generator]
        WP --> Bus[Pub/Sub Event Bus]
    end

    subgraph Delivery & Action
        Bus --> SSE[SSE /api/v1/stream/results]
        Bus --> WS[WebSocket Live Feed]
        Bus --> Rem[Self-Healing Remediation]
        Bus --> Dedup[Alert Deduplication & Correlation]
        Dedup --> PagerDuty[PagerDuty v2 Auto-Resolve]
        Dedup --> Teams[Microsoft Teams Adaptive Cards]
        Dedup --> Discord[Discord Rich Embeds]
        Dedup --> Slack[Slack / SMS / Email]
        Bus --> HAR[HAR 1.2 Export]
        Bus --> StatusPage[Public Status Page & RSS/Atom]
    end
```

---

## 🚀 Key Enterprise Features

### High-Throughput Bounded Concurrency Worker Pool
- Bounded concurrency worker pool preventing goroutine leaks or out-of-memory crashes under massive target load.
- Backpressure rejection and graceful shutdown with Go `context.Context` cancellation.

### Multi-Protocol Probing Engines
- **HTTP / HTTPS**: Full response assertion (body regex, status codes, headers, and latency bounds).
- **gRPC Health Checking Protocol**: Standard `grpc.health.v1.Health/Check` status verification (`SERVING`, `NOT_SERVING`, `UNKNOWN`).
- **TCP / Port**: Connection handshake and socket ping.
- **UDP**: Reachability probe with ICMP port-unreachable detection.
- **DNS & DNSSEC**: Record lookup, cryptographic RRSIG signature validation, and root-trust chain checking.
- **TLS & OCSP Stapling**: SSL expiration alerts, certificate revocation verification via OCSP stapling and CRL.
- **HTTP Security Headers Auditor**: Automated audit and A+ to F grading for HSTS, CSP, X-Frame-Options, and Permissions-Policy.

### X.509 Certificate Chain & SAN Inspector
- Deep inspection of peer certificates:
  - Subject Alternative Names (SAN) matching, including wildcard patterns (`*.example.com`).
  - Warning tiers for expiration: 30 days, 14 days, 7 days, and expired.
  - Intermediate Certificate Authority (CA) chain completeness.
  - Flags weak cryptographic signature algorithms (SHA-1, MD5).

### Mutual TLS (mTLS) Zero-Trust Authentication
- Support for internal zero-trust microservice probing using client certificates, private keys, and custom CA trust bundles (`internal/checker/mtls.go`).

### Browser-Like Multi-Step Synthetic Scenarios
- Multi-step transaction runner (`internal/synthetic/scenario.go`) with session cookie jar persistence across steps.
- Automatic redirect chain tracking (up to 10 hops) and step-by-step performance waterfall metrics (DNS, Connect, TLS, TTFB, and Transfer).
- DOM & HTML selector matcher (`internal/synthetic/dom_matcher.go`) supporting `#id`, `.class`, tags, and regex extraction.

### W3C HAR 1.2 HTTP Archive Exporter
- Export synthetic scenario runs as standard W3C HAR 1.2 JSON files (`internal/synthetic/har_exporter.go`) ready for immediate import into Chrome DevTools, Charles Proxy, and Postman.

### Automated Canary Analysis (ACA) & Progressive Rollouts
- Automated canary evaluator (`internal/canary/evaluator.go`) comparing error rate delta and P95/P99 latency regressions between baseline and canary targets, delivering clear `PROMOTE`, `HALT`, or `ROLLBACK` rollout recommendations.

### Google SRE Multi-Window Multi-Burn-Rate SLO Engine
Implements the Google SRE Handbook Chapter 5 standard multi-window burn rate alerts:
- **1-Hour Window (14.4x Burn Rate)**: Consumes 2% error budget in 1 hour &rarr; PagerDuty Critical paging.
- **6-Hour Window (6.0x Burn Rate)**: Consumes 5% error budget in 6 hours &rarr; High severity alert.
- **24-Hour / 3-Day Window (3.0x / 1.0x Burn Rate)**: 10% budget consumed &rarr; Engineering ticket alert.
- Computes remaining downtime budget in minutes and real-time **Time-to-Exhaustion** in hours.

### Dynamic Adaptive Sampling
- Automatically ramps sampling rate to 100% when targets fail or latency degrades. Downsamples steady green targets to configurable baseline (e.g. 20%) to conserve network and database storage.

### Alert De-duplication & Failure Domain Correlation
- Groups firing alerts by shared failure domains (cloud provider, availability zone, cluster, gateway) within a sliding time window, preventing alert fatigue by consolidating 50 alerts into a single incident cluster alert.

### Bi-directional PagerDuty v2 Sync & Chat Webhooks
- PagerDuty Events API v2 client with automatic incident triggers and automatic resolve events upon health check recovery.
- Rich Microsoft Teams Adaptive Cards and Discord Embed formatters.

### Geo-Matrix Consensus & Exponential Flap Dampening
- **Geo-Matrix Evaluator**: Distinguishes between localized CDN edge degradation and global origin outages.
- **Flap Dampener**: Exponential decay hysteresis suppresses alert flapping storms while services rapidly oscillate during recovery.

### Chaos Engineering & Synthetic Fault Injection
- Synthetic fault injection engine simulates latency spikes, packet loss, and HTTP 500/503 status overrides.
- **Blast Radius Protection**: Strict environment checks prevent running chaos experiments against production targets; enforces max percentage limits.

### Automated SRE Incident Post-Mortem & Timeline Generator
- Auto-generates GitHub-flavored Markdown post-mortem documents aggregating root causes, affected regions, chronological timeline events, and action items.
- Reliability metrics calculator: Calculates MTTD (Mean Time to Detect), MTTR (Mean Time to Resolve), total downtime in minutes, and error budget impact.

---

## ⚡ Getting Started

### 1. Launch with Docker Compose

```bash
git clone https://github.com/sujanv/PulseGrid.git
cd PulseGrid
docker-compose up -d
```

### 2. Verify Services

- **Web Dashboard**: [http://localhost:8080](http://localhost:8080)
- **Public Status Page**: [http://localhost:8080/status](http://localhost:8080/status)
- **Prometheus Metrics**: [http://localhost:8080/metrics](http://localhost:8080/metrics)
- **Kubernetes Health**: [http://localhost:8080/healthz](http://localhost:8080/healthz)

---

## 💻 CLI Reference

PulseGrid includes a command-line interface for probes, synthetic journeys, network diagnostics, and post-mortems:

```bash
# Execute immediate single probe
./bin/pulsegrid-cli check https://example.com --timeout=5

# Deep inspect SSL/TLS certificate and SAN chain
./bin/pulsegrid-cli cert-inspect api.pulsegrid.io

# Export synthetic scenario to W3C HAR 1.2 JSON
./bin/pulsegrid-cli har-export ./examples/user-journey.json

# Evaluate progressive canary rollout vs baseline
./bin/pulsegrid-cli canary-eval baseline-api canary-api

# Execute hop-by-hop traceroute network diagnostic
./bin/pulsegrid-cli traceroute example.com

# Calculate Google SRE error budget and time-to-exhaustion
./bin/pulsegrid-cli slo-calc tgt-payment-service

# Generate automated incident post-mortem markdown
./bin/pulsegrid-cli postmortem-gen INC-1001

# Evaluate multi-region quorum consensus
./bin/pulsegrid-cli geo-quorum tgt-payment-service

# Run SSL and HTTP security headers audit (A+ to F grade)
./bin/pulsegrid-cli audit https://api.pulsegrid.io

# Export 30-day SLA performance report to CSV
./bin/pulsegrid-cli export
```

---

## 📡 REST API Matrix

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/api/v1/targets` | List all monitored targets |
| `POST` | `/api/v1/targets` | Register a new target |
| `GET` | `/api/v1/targets/:id/results` | Historical check results |
| `GET` | `/api/v1/targets/:id/stats` | P50, P95, P99 Latency & Uptime % |
| `POST` | `/api/v1/targets/har/export` | Export synthetic scenario to W3C HAR 1.2 |
| `POST` | `/api/v1/canary/evaluate` | Automated Canary Analysis (ACA) decision |
| `GET` | `/api/v1/network/traceroute?host=...` | Hop-by-hop network path diagnosis |
| `GET` | `/api/v1/slo/burn-rate?target_id=...` | Multi-window burn-rate error budget |
| `POST` | `/api/v1/scenarios/run` | Execute browser-like multi-step scenario |
| `GET` | `/api/v1/incidents/postmortem?id=...` | Generate SRE incident post-mortem markdown |
| `POST` | `/api/v1/integrations/alertmanager` | Ingest Prometheus Alertmanager webhook |
| `GET` | `/api/v1/stream/results` | Server-Sent Events (SSE) live results |
| `POST` | `/api/v1/remediation/:id` | Configure automated remediation action |
| `GET` | `/api/v1/statuspage/components` | Component status summary |
| `GET` | `/status/feed.rss` | RSS 2.0 incident feed |
| `GET` | `/status/feed.atom` | Atom XML incident feed |
| `GET` | `/badge/:target_id.svg` | Dynamic SVG Uptime Badge |
| `GET` | `/healthz` | Kubernetes liveness probe |
| `GET` | `/metrics` | Prometheus metrics scrape endpoint |

---

## 🧪 Testing & Quality Assurance

```bash
# Run all unit and integration tests
go test -v ./...

# Run tests with race condition detection
go test -v -race ./...

# Run concurrency benchmarks
go test -v -bench=. ./test/stress_test.go
```

---

## 📄 License

MIT License. See [LICENSE](LICENSE) for details.
