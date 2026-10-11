# PulseGrid ⚡

**PulseGrid** is an enterprise-grade, bounded-concurrency uptime and latency monitoring engine written in Go. Designed to reliably schedule and execute thousands of checks across multiple protocols (HTTP/S, TCP, DNS, TLS, ICMP Ping, WebSocket, gRPC) while strictly bounding memory and goroutine concurrency.

---

## 🌟 Key Features & Capabilities

- **Bounded-Concurrency Worker Pool**: Fixed-size goroutine pool backed by a buffered queue to monitor thousands of endpoints without runaway concurrency or memory leaks.
- **Microsecond Phase Latency Breakdown**: Uses `net/http/httptrace` to record DNS lookup, TCP handshake, TLS negotiation, and Time-to-First-Byte (TTFB).
- **JSONPath & Regex Assertion Engine**: Validate complex JSON API responses (`$.data.status == "healthy"`), HTTP response headers, and body regex patterns.
- **Latency SLA Breach Degradation**: Automatic transition to `DEGRADED` status when response times exceed defined SLA thresholds.
- **Multi-Protocol Health Checkers**:
  - **HTTP/HTTPS**: Status assertion, custom headers, payload assertions, keyword matching, W3C TraceContext propagation.
  - **TCP**: Socket connect timeout, banner retrieval.
  - **DNS**: Recursive resolution duration and IP record matching.
  - **TLS/SSL**: Expiration alerts, cipher suite security, and weak protocol detection (TLS 1.0/1.1 warnings).
  - **ICMP Ping**: Packet loss percentage and min/avg/max round-trip latency.
  - **WebSocket**: RFC 6455 handshake verification and frame round-trip timing.
  - **gRPC**: Native `grpc.health.v1.Health` protocol health probing.
- **Incident State Machine & Anti-Flapping**: Debounced consecutive failure/recovery thresholds and sliding-window flap damping to prevent alert storms.
- **Maintenance Windows & Silence Rules**: Schedule recurring or one-off maintenance windows to silence alerts and pause SLA tracking.
- **Multi-Region Quorum Consensus**: Eliminates false positives by requiring majority consensus (e.g., 2 of 3 regions) before triggering downtime alerts.
- **Multi-Channel Alert Dispatcher**:
  - Generic Webhook (HMAC-SHA256 signature verification).
  - Slack Block Kit formatted message cards.
  - Discord rich embed cards.
  - PagerDuty Events API v2 with automatic incident deduplication and resolution.
  - Telegram Bot API with markdown alerts.
  - Microsoft Teams / Office 365 Adaptive Cards.
  - SMTP Email delivery with HTML incident summaries.
- **Public Status Page & SVG Badges**: Clean public-facing status page at `/status` and dynamic SVG uptime badges for READMEs (`/badge/{id}.svg`).
- **Interactive Terminal TUI**: Live-updating terminal dashboard with real-time status and latency sparklines (`pulsegrid-cli tui`).
- **Data Rollup & Pruning Worker**: Automated background worker that aggregates hourly checks into daily rollups and prunes historical data.
- **Production Observability**: Full Prometheus metrics (`/metrics`) and auto-provisioned Grafana dashboards.

---

## 🏗 System Architecture

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
           │     Assertion Engine (JSONPath, Regex, SLA)         │
           └──────────────────────────┬──────────────────────────┘
                                      │
                                      ▼
           ┌─────────────────────────────────────────────────────┐
           │        Incident Detector & Flap Dampener            │
           │  (Maintenance Windows & Multi-Region Quorum)        │
           └──────────────────────────┬──────────────────────────┘
                                      │
                                      ▼
           ┌─────────────────────────────────────────────────────┐
           │        Async Alert Dispatcher (Worker Pool)         │
           │  (Slack / Discord / PagerDuty / Telegram / Teams)   │
           └─────────────────────────────────────────────────────┘
```

---

## 🚀 Quickstart

### Running with Docker Compose (PulseGrid + PostgreSQL + Redis + Prometheus + Grafana)

```bash
docker-compose -f deploy/docker-compose.yml up -d --build
```

Access services:
- **Admin Dashboard**: [http://localhost:8080](http://localhost:8080)
- **Public Status Page**: [http://localhost:8080/status](http://localhost:8080/status)
- **Prometheus Metrics**: [http://localhost:8080/metrics](http://localhost:8080/metrics)
- **Grafana Dashboard**: [http://localhost:3000](http://localhost:3000) (User: `admin` / Password: `admin`)
- **Health Probes**: [http://localhost:8080/healthz](http://localhost:8080/healthz) and `/readyz`

### Local Development

```bash
# Build server and CLI
make build

# Run server
make run

# Seed sample monitors
./bin/pulsegrid-cli seed

# Launch Terminal TUI Mode
./bin/pulsegrid-cli tui

# Probe endpoint with latency breakdown
./bin/pulsegrid-cli check https://github.com

# Export SLA report to CSV
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
| `GET` | `/api/v1/incidents` | Incident history |
| `GET` | `/api/v1/incidents/active` | Current active outages |
| `GET` | `/api/v1/maintenance` | List maintenance windows |
| `POST` | `/api/v1/maintenance` | Create a maintenance window |
| `DELETE` | `/api/v1/maintenance/:id` | Delete a maintenance window |
| `GET` | `/api/v1/reports/sla` | 30-day SLA performance report |
| `GET` | `/api/v1/reports/export.csv` | Export SLA report to CSV |
| `GET` | `/badge/:target_id.svg` | Dynamic SVG Uptime Badge |
| `GET` | `/status` | Public status page |
| `GET` | `/healthz` | Kubernetes liveness probe |
| `GET` | `/readyz` | Kubernetes readiness probe |
| `GET` | `/metrics` | Prometheus metrics scrape endpoint |

---

## 🏷 Adding SVG Badges to README

Add real-time uptime badges to your repository:

```markdown
![Uptime](https://pulsegrid.io/badge/your-target-id.svg)
```

---

## ⚙️ Configuration Reference

| Environment Variable | Default | Description |
|---|---|---|
| `PULSEGRID_PORT` | `8080` | Server HTTP port |
| `PULSEGRID_WORKER_COUNT` | `50` | Concurrent worker pool size |
| `PULSEGRID_QUEUE_SIZE` | `5000` | Buffer capacity of check job channel |
| `PULSEGRID_HOST_RATE_LIMIT` | `5.0` | Max checks per second per hostname |
| `PULSEGRID_SCHEDULER_JITTER` | `0.10` | Jitter offset ratio (10%) |
| `PULSEGRID_DB_URL` | `postgres://...` | PostgreSQL connection URL |
| `PULSEGRID_REDIS_ADDR` | `localhost:6379` | Redis address for distributed locking |
| `PULSEGRID_API_KEY` | `""` | Master admin API key |

---

## 🧪 Testing & Benchmarks

```bash
# Run unit and race condition tests
make test

# Run throughput benchmarks
make bench
```

---

## 📄 License

MIT License. See [LICENSE](LICENSE) for details.
