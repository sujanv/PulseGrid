package streaming

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type SSEBroker struct {
	clients map[chan []byte]bool
	newClients chan chan []byte
	closingClients chan chan []byte
	messages chan []byte
	mu sync.Mutex
}

func NewSSEBroker() *SSEBroker {
	b := &SSEBroker{
		clients: make(map[chan []byte]bool),
		newClients: make(chan chan []byte),
		closingClients: make(chan chan []byte),
		messages: make(chan []byte, 100),
	}
	go b.listen()
	return b
}

func (b *SSEBroker) listen() {
	for {
		select {
		case s := <-b.newClients:
			b.mu.Lock()
			b.clients[s] = true
			b.mu.Unlock()
		case s := <-b.closingClients:
			b.mu.Lock()
			delete(b.clients, s)
			close(s)
			b.mu.Unlock()
		case msg := <-b.messages:
			b.mu.Lock()
			for s := range b.clients {
				select {
				case s <- msg:
				default:
				}
			}
			b.mu.Unlock()
		}
	}
}

func (b *SSEBroker) Broadcast(data interface{}) {
	bytes, err := json.Marshal(data)
	if err == nil {
		b.messages <- bytes
	}
}

func (b *SSEBroker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	msgChan := make(chan []byte, 10)
	b.newClients <- msgChan
	defer func() {
		b.closingClients <- msgChan
	}()

	notify := r.Context().Done()
	for {
		select {
		case <-notify:
			return
		case msg := <-msgChan:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		}
	}
}
