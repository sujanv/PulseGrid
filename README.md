# PulseGrid ⚡

**PulseGrid** is an enterprise-grade, bounded-concurrency uptime and synthetic monitoring platform written in Go. Designed to schedule, execute, and analyze thousands of checks across multiple protocols (HTTP/S, TCP, UDP, DNS, TLS, ICMP Ping, WebSocket, gRPC, and Multi-Step Synthetics) while strictly bounding memory and goroutine footprints.

---

## 🌟 Advanced Features & Capabilities

- **Bounded-Concurrency Worker Pool**: High-performance goroutine pool with dynamic resizing and backpressure rejection handling.
- **Microsecond Phase Latency Breakdown**: Powered by `net/http/httptrace` to record DNS lookup, TCP connect, TLS handshake, and Time-to-First-Byte (TTFB).
- **Multi-Step Synthetic User Journeys**: Execute multi-request workflows (e.g. login -> token extraction -> authenticated API call) with dynamic variable interpolation (`{{authToken}}`).
- **Statistical Anomaly Detection**: Real-time sliding-window Welford algorithm detecting latency spikes ($Z > 3.0$) and diurnal baselines before total outages occur.
- **Intelligent Alert Escalation & Quiet Hours**:
  - Multi-tier delayed escalation policies (Tier 1 immediate -> Tier 2 after 10m -> Tier 3 after 30m).
  - Timezone-aware quiet hours suppression that prevents non-critical notifications from waking on-call engineers.
  - Alert grouping to bundle cascading multi-service outages into unified summaries.
- **Enterprise Alert Integrations**:
  - PagerDuty Events API v2 (auto-deduplication and resolution).
  - Atlassian OpsGenie with priority tagging (P1-P5).
  - Twilio SMS with automated truncation.
  - Telegram Bot API with markdown alert notifications.
  - Microsoft Teams / Office 365 Adaptive Cards.
  - Slack Block Kit and Discord rich embeds.
  - SMTP Email delivery with responsive HTML templates.
- **Multi-Region Quorum Consensus**: Geo-distributed synthetic agents requiring majority consensus (e.g., 2 of 3 regions) to eliminate false positives.
- **Security & DNSSEC Posture Auditing**:
  - HTTP security header analyzer with automated scoring (A+ through F) covering HSTS, CSP, X-Frame-Options, and Permissions-Policy.
  - DNSSEC validation and SSL/TLS revocation check via OCSP Stapling.
- **Maintenance Windows & Silence Rules**: Target-specific or global maintenance windows with scheduled suppression of alerts and SLA pauses.
- **Automated Data Rollups & Retention Pruning**: Background hourly worker aggregates check results into daily summaries and prunes raw time-series records.
- **Observability & Visualizations**:
  - Dynamic Shields.io-compatible SVG uptime badges (`/badge/{id}.svg`).
  - Public status page (`/status`).
  - Terminal TUI Mode (`pulsegrid-cli tui`).
  - Global latency heatmap (`/heatmap`).
  - Native Prometheus metrics (`/metrics`) and pre-configured Grafana dashboards.

---

## 🏗 High-Level Architecture

```
                    ┌───────────────────────────────────┐
                    │      Scheduler with Jitter        │
                    │   (Prevents Thundering Herd)      │
                    └─────────────────┬─────────────────┘
                                      │
                                      ▼
                    ┌───────────────────────────────────┐
                    │      Per-Host Rate Limiter        │
                    │   (Token Bucket per Domain)       │
                    └─────────────────┬─────────────────┘
                                      │
                                      ▼
                    ┌───────────────────────────────────┐
                    │      Buffered Job Queue (Ch)      │
                    └─────────────────┬─────────────────┘
                                      │
           ┌──────────────────────────┼──────────────────────────┐
           ▼                          ▼                          ▼
    ┌──────────────┐           ┌──────────────┐           ┌──────────────┐
    │   Worker 1   │           │   Worker 2   │           │   Worker N   │  (Fixed Pool: 50–500)
    └──────┬───────┘           └──────┬───────┘           └──────┬───────┘
           │                          │                          │
           └──────────────────────────┼──────────────────────────┘
                                      │
                                      ▼
           ┌─────────────────────────────────────────────────────┐
           │     Assertion & Synthetic Transaction Engine        │
           └──────────────────────────┬──────────────────────────┘
                                      │
                                      ▼
           ┌─────────────────────────────────────────────────────┐
           │        Incident Detector & Anomaly Engine           │
           │  (Welford Z-Score, Maintenance, Quorum Consensus)   │
           └──────────────────────────┬──────────────────────────┘
                                      │
                                      ▼
           ┌─────────────────────────────────────────────────────┐
           │   Escalation Router & Multi-Channel Alert Queue     │
           │ (PagerDuty / OpsGenie / Slack / Teams / SMS / Mail) │
           └─────────────────────────────────────────────────────┘
```

---

## 🚀 Quickstart

### Running the Full Stack (PulseGrid + PostgreSQL + Redis + Prometheus + Grafana)

```bash
docker-compose -f deploy/docker-compose.yml up -d --build
```

Access services:
- **Admin Dashboard**: [http://localhost:8080](http://localhost:8080)
- **Public Status Page**: [http://localhost:8080/status](http://localhost:8080/status)
- **Regional Latency Heatmap**: [http://localhost:8080/heatmap](http://localhost:8080/heatmap)
- **Prometheus Metrics**: [http://localhost:8080/metrics](http://localhost:8080/metrics)
- **Grafana Dashboard**: [http://localhost:3000](http://localhost:3000) (User: `admin` / Password: `admin`)
- **Health Probes**: [http://localhost:8080/healthz](http://localhost:8080/healthz) and `/readyz`

### Local Development & CLI

```bash
# Build binaries
make build

# Launch Terminal TUI Mode
./bin/pulsegrid-cli tui

# Probe endpoint with full latency phase breakdown
./bin/pulsegrid-cli check https://github.com

# Audit HTTP security headers & calculate posture grade
./bin/pulsegrid-cli audit https://github.com

# Seed sample monitors
./bin/pulsegrid-cli seed

# Export 30-day SLA performance report
./bin/pulsegrid-cli export
```

---

## 📡 REST API Reference

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/api/v1/targets` | List all monitored targets |
| `POST` | `/api/v1/targets` | Register a new target |
| `GET` | `/api/v1/targets/:id` | Get target details |
| `PUT` | `/api/v1/targets/:id` | Update target configuration |
| `DELETE` | `/api/v1/targets/:id` | Remove target |
| `GET` | `/api/v1/targets/:id/results` | Historical check results |
| `GET` | `/api/v1/targets/:id/stats` | Latency percentiles (P50, P95, P99) & Uptime % |
| `GET` | `/api/v1/targets/:id/anomalies` | Latency spike anomaly history |
| `POST` | `/api/v1/synthetics` | Create multi-step synthetic transaction |
| `POST` | `/api/v1/synthetics/run` | Execute on-demand test of synthetic workflow |
| `GET` | `/api/v1/agents` | List registered synthetic edge agents |
| `POST` | `/api/v1/agents/register` | Register new remote edge agent |
| `POST` | `/api/v1/agents/heartbeat` | Edge agent heartbeat |
| `GET` | `/api/v1/incidents` | Incident history |
| `GET` | `/api/v1/incidents/active` | Current active outages |
| `GET` | `/api/v1/maintenance` | List maintenance windows |
| `POST` | `/api/v1/maintenance` | Create a maintenance window |
| `DELETE` | `/api/v1/maintenance/:id` | Delete a maintenance window |
| `GET` | `/api/v1/reports/sla` | 30-day SLA performance report |
| `GET` | `/api/v1/reports/export.csv` | Export SLA report to CSV |
| `GET` | `/badge/:target_id.svg` | Dynamic SVG Uptime Badge |
| `GET` | `/status` | Public status page |
| `GET` | `/heatmap` | Global latency heatmap matrix |
| `GET` | `/healthz` | Kubernetes liveness probe |
| `GET` | `/readyz` | Kubernetes readiness probe |
| `GET` | `/metrics` | Prometheus metrics scrape endpoint |

---

## 🏷 Dynamic README Uptime Badges

Embed real-time uptime status directly into your GitHub README:

```markdown
![Uptime](https://pulsegrid.io/badge/your-target-id.svg)
```

---

## 🧪 Testing & Stress Benchmarks

```bash
# Run unit and race condition tests
make test

# Run high-throughput concurrency benchmarks
make bench
```

---

## 📄 License

MIT License. See [LICENSE](LICENSE) for details.
