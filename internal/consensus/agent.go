package consensus

import (
	"sync"
	"time"
)

type Agent struct {
	ID            string    `json:"id"`
	Region        string    `json:"region"` // e.g. "us-east-1", "eu-central-1", "ap-southeast-1"
	LastHeartbeat time.Time `json:"last_heartbeat"`
}

type AgentRegistry struct {
	mu     sync.RWMutex
	agents map[string]*Agent
}

func NewAgentRegistry() *AgentRegistry {
	return &AgentRegistry{
		agents: make(map[string]*Agent),
	}
}

func (r *AgentRegistry) Register(id, region string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.agents[id] = &Agent{
		ID:            id,
		Region:        region,
		LastHeartbeat: time.Now().UTC(),
	}
}

func (r *AgentRegistry) ActiveRegions() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	regions := make(map[string]bool)
	cutoff := time.Now().UTC().Add(-5 * time.Minute)

	for _, a := range r.agents {
		if a.LastHeartbeat.After(cutoff) {
			regions[a.Region] = true
		}
	}

	var res []string
	for reg := range regions {
		res = append(res, reg)
	}
	return res
}
