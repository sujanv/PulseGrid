package streaming

import (
	"fmt"
	"net/http"
	"sync"
)

type WebSocketHub struct {
	mu          sync.RWMutex
	subscribers map[string]http.ResponseWriter
}

func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		subscribers: make(map[string]http.ResponseWriter),
	}
}

func (h *WebSocketHub) BroadcastUpdate(event, payload string) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_ = fmt.Sprintf("event:%s payload:%s", event, payload)
}
