package web

import (
	"fmt"
	"net/http"
)

func ServeHeatmap(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>PulseGrid Regional Latency Matrix</title>
  <style>
    body { font-family: sans-serif; background: #0d1117; color: #c9d1d9; padding: 24px; }
    table { width: 100%; border-collapse: collapse; margin-top: 20px; }
    th, td { border: 1px solid #30363d; padding: 12px; text-align: center; }
    th { background: #161b22; }
    .good { background: rgba(46, 160, 67, 0.25); color: #3fb950; font-weight: bold; }
    .fair { background: rgba(210, 153, 34, 0.25); color: #d29922; }
    .slow { background: rgba(248, 81, 73, 0.25); color: #f85149; }
  </style>
</head>
<body>
  <h2>🌍 Multi-Region Latency Heatmap</h2>
  <table>
    <tr><th>Target</th><th>US-East</th><th>US-West</th><th>EU-Central</th><th>AP-South</th><th>AP-East</th></tr>
    <tr><td>Cloudflare 1.1.1.1</td><td class="good">8 ms</td><td class="good">12 ms</td><td class="good">14 ms</td><td class="good">22 ms</td><td class="good">25 ms</td></tr>
    <tr><td>Google Primary DNS</td><td class="good">11 ms</td><td class="good">9 ms</td><td class="good">16 ms</td><td class="good">28 ms</td><td class="good">31 ms</td></tr>
    <tr><td>GitHub API</td><td class="good">35 ms</td><td class="good">22 ms</td><td class="fair">110 ms</td><td class="fair">170 ms</td><td class="fair">165 ms</td></tr>
  </table>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(html))
}
