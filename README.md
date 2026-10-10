# PulseGrid ⚡

**PulseGrid** is a high-performance, bounded-concurrency uptime monitoring and alerting engine written in Go. It schedules and evaluates thousands of endpoint checks (HTTP, HTTPS, TCP, DNS, TLS) while keeping memory and goroutine footprints strictly controlled.

---

## 🌟 Key Features

- **Bounded-Concurrency Worker Pool**: Fixed-size goroutine pool backed by a buffered queue to monitor thousands of endpoints without risking runaway concurrency or memory exhaustion.
- **Microsecond Phase Latency Breakdown**: Powered by `net/http/httptrace` to record DNS lookup, TCP handshake, TLS negotiation, and Time-to-First-Byte (TTFB).
- **Incident State Machine**: Debounced transition logic with configurable failure/recovery thresholds and anti-flapping suppression to prevent alert fatigue.
- **Multi-Protocol Probes**:
  - **HTTP/HTTPS**: Status assertions, keyword/regex matching, custom request payloads.
  - **TCP**: Socket connect probe and banner latency verification.
  - **DNS**: Recursive resolution times and IP record verification.
  - **TLS/SSL**: Certificate expiration inspection with proactive warning triggers.
- **Multi-Channel Alerts**: Webhook (with HMAC-SHA256 signatures), Slack Block Kit, Discord rich embeds, and SMTP email.
- **Production Observability**: Native Prometheus `/metrics` exposing latency histograms, check counters, and worker pool gauges.
- **Embedded Web UI**: Single-page dashboard embedded into the binary for status tracking.

---

## 🏗 Concurrency & Architecture Model

```
                    ┌────────────────────────────┐
                    │    Scheduler with Jitter   │
                    │ (Prevents Thundering Herd) │
                    └─────────────┬──────────────┘
                                  │
                                  ▼
                    ┌────────────────────────────┐
                    │  Per-Host Rate Limiter     │
                    │  (Token Bucket per Domain) │
                    └─────────────┬──────────────┘
                                  │
                                  ▼
                    ┌────────────────────────────┐
                    │  Buffered Job Queue (Ch)   │
                    └─────────────┬──────────────┘
                                  │
           ┌──────────────────────┼──────────────────────┐
           ▼                      ▼                      ▼
    ┌──────────────┐       ┌──────────────┐       ┌──────────────┐
    │  Worker 1    │       │  Worker 2    │       │  Worker N    │  (Fixed Pool Size: 50-500)
    └──────┬───────┘       └──────┬───────┘       └──────┬───────┘
           │                      │                      │
           └──────────────────────┼──────────────────────┘
                                  │
                                  ▼
           ┌─────────────────────────────────────────────┐
           │        Incident Detector & Flap Dampener    │
           └──────────────────────┬──────────────────────┘
                                  │
                                  ▼
           ┌─────────────────────────────────────────────┐
           │     Async Alert Dispatcher (Worker Pool)    │
           │  (Slack / Discord / Webhook / SMTP Email)   │
           └─────────────────────────────────────────────┘
```

---

## 🚀 Quickstart

### Running with Docker Compose

Spin up PulseGrid, PostgreSQL 16, Redis 7, and Prometheus:

```bash
docker-compose -f deploy/docker-compose.yml up -d
```

- **Dashboard UI**: [http://localhost:8080](http://localhost:8080)
- **Prometheus Metrics**: [http://localhost:8080/metrics](http://localhost:8080/metrics)
- **Health Probes**: [http://localhost:8080/healthz](http://localhost:8080/healthz) and `/readyz`

### Local Development

```bash
# Build binaries
make build

# Run unit and race tests
make test

# Execute benchmarks
make bench

# Probe an endpoint instantly via CLI
./bin/pulsegrid-cli check https://google.com
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
| `GET` | `/api/v1/incidents` | List incident history |
| `GET` | `/api/v1/incidents/active` | Current active outages |
| `GET` | `/healthz` | Liveness probe |
| `GET` | `/readyz` | Readiness probe (checks DB, Redis, WorkerPool) |
| `GET` | `/metrics` | Prometheus metrics scrape target |

---

## ⚙️ Configuration

| Variable | Default | Description |
|---|---|---|
| `PULSEGRID_PORT` | `8080` | HTTP server port |
| `PULSEGRID_WORKER_COUNT` | `50` | Max concurrent worker goroutines |
| `PULSEGRID_QUEUE_SIZE` | `5000` | Buffer capacity of job channel |
| `PULSEGRID_HOST_RATE_LIMIT` | `5.0` | Max requests/sec per target host |
| `PULSEGRID_SCHEDULER_JITTER` | `0.10` | Random jitter spread (10%) |
| `PULSEGRID_DB_URL` | `postgres://...` | PostgreSQL connection string |
| `PULSEGRID_REDIS_ADDR` | `localhost:6379` | Redis host:port for locks |

---

## 📄 License

MIT License. See [LICENSE](LICENSE) for details.
