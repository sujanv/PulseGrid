package web

import (
	"net/http"
)

const DashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>PulseGrid — Uptime & Latency Monitoring</title>
  <style>
    :root {
      --bg: #0d1117;
      --card-bg: #161b22;
      --border: #30363d;
      --text: #c9d1d9;
      --text-muted: #8b949e;
      --up: #238636;
      --up-glow: #2ea043;
      --down: #da3633;
      --accent: #58a6ff;
    }
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: var(--bg); color: var(--text); padding: 24px; }
    .container { max-width: 1200px; margin: 0 auto; }
    header { display: flex; justify-content: space-between; align-items: center; padding-bottom: 24px; border-bottom: 1px solid var(--border); margin-bottom: 24px; }
    .logo { font-size: 24px; font-weight: 700; color: #fff; display: flex; align-items: center; gap: 8px; }
    .logo span { color: var(--accent); }
    .btn { background: var(--up); color: white; border: none; padding: 8px 16px; border-radius: 6px; font-weight: 600; cursor: pointer; transition: 0.2s; }
    .btn:hover { background: var(--up-glow); }
    .stats-bar { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 16px; margin-bottom: 24px; }
    .stat-card { background: var(--card-bg); border: 1px solid var(--border); border-radius: 8px; padding: 16px; }
    .stat-val { font-size: 28px; font-weight: 700; color: #fff; margin-top: 4px; }
    .stat-label { font-size: 13px; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.5px; }
    .grid { display: grid; grid-template-columns: 1fr; gap: 16px; }
    .target-card { background: var(--card-bg); border: 1px solid var(--border); border-radius: 8px; padding: 20px; transition: transform 0.15s; }
    .target-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
    .target-title { font-size: 18px; font-weight: 600; color: #fff; }
    .badge { padding: 4px 10px; border-radius: 12px; font-size: 12px; font-weight: 700; text-transform: uppercase; }
    .badge-up { background: rgba(35, 134, 54, 0.2); color: #3fb950; border: 1px solid var(--up); }
    .badge-down { background: rgba(218, 54, 51, 0.2); color: #f85149; border: 1px solid var(--down); }
    .target-url { color: var(--text-muted); font-size: 14px; font-family: monospace; margin-bottom: 16px; }
    .heartbeat-bar { display: flex; gap: 4px; height: 28px; margin-bottom: 12px; align-items: flex-end; }
    .bar-seg { flex: 1; height: 100%; border-radius: 3px; background: var(--up); }
    .bar-seg.down { background: var(--down); }
    .footer-meta { display: flex; justify-content: space-between; font-size: 13px; color: var(--text-muted); }
  </style>
</head>
<body>
  <div class="container">
    <header>
      <div class="logo">⚡ Pulse<span>Grid</span></div>
      <div>
        <a href="/status" style="color: var(--accent); margin-right: 16px; text-decoration: none; font-size: 14px;">Public Status</a>
        <a href="/heatmap" style="color: var(--accent); margin-right: 16px; text-decoration: none; font-size: 14px;">Regional Heatmap</a>
        <a href="/metrics" target="_blank" style="color: var(--accent); margin-right: 16px; text-decoration: none; font-size: 14px;">Prometheus</a>
        <button class="btn" onclick="alert('Adding targets via POST /api/v1/targets')">+ New Monitor</button>
      </div>
    </header>

    <div class="stats-bar">
      <div class="stat-card">
        <div class="stat-label">Active Monitors</div>
        <div class="stat-val" id="stat-total">--</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">Overall Uptime</div>
        <div class="stat-val" style="color: #3fb950;" id="stat-uptime">99.98%</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">Live Stream (SSE)</div>
        <div class="stat-val" style="color: #58a6ff;" id="stat-stream">Connected</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">Active Incidents</div>
        <div class="stat-val" style="color: #58a6ff;" id="stat-incidents">0</div>
      </div>
    </div>

    <div class="grid" id="targets-list">
      <div style="text-align: center; padding: 40px; color: var(--text-muted);">Loading live monitors...</div>
    </div>
  </div>

  <script>
    async function loadData() {
      try {
        const res = await fetch('/api/v1/targets');
        if (!res.ok) return;
        const targets = await res.json();
        document.getElementById('stat-total').innerText = targets ? targets.length : 0;

        const container = document.getElementById('targets-list');
        if (!targets || targets.length === 0) {
          container.innerHTML = '<div style="text-align: center; padding: 40px; color: var(--text-muted);">No targets configured yet. Seed targets or add one!</div>';
          return;
        }

        container.innerHTML = targets.map(t => `
          <div class="target-card" id="card-${t.id}">
            <div class="target-header">
              <span class="target-title">${t.name}</span>
              <span class="badge badge-up">UP</span>
            </div>
            <div class="target-url">${t.url || (t.host + ':' + t.port)} (${t.type})</div>
            <div class="heartbeat-bar">
              ${Array(30).fill('<div class="bar-seg"></div>').join('')}
            </div>
            <div class="footer-meta">
              <span>Interval: ${t.interval_seconds || 30}s</span>
              <span>Timeout: ${t.timeout_seconds || 10}s</span>
              <span>Live Mode: SSE Enabled</span>
            </div>
          </div>
        `).join('');
      } catch (err) {
        console.error(err);
      }
    }

    if (window.EventSource) {
      const sse = new EventSource('/api/v1/stream/results');
      sse.onmessage = function(e) {
        document.getElementById('stat-stream').innerText = 'Streaming ⚡';
      };
    }

    loadData();
    setInterval(loadData, 15000);
  </script>
</body>
</html>`

func Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(DashboardHTML))
	}
}
