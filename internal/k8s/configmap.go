package k8s

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
)

type ConfigMapWatcher struct {
	mu       sync.Mutex
	lastHash string
}

func NewConfigMapWatcher() *ConfigMapWatcher {
	return &ConfigMapWatcher{}
}

func (w *ConfigMapWatcher) HasChanged(content []byte) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	h := sha256.Sum256(content)
	hashStr := hex.EncodeToString(h[:])

	if hashStr != w.lastHash {
		w.lastHash = hashStr
		return true
	}
	return false
}
