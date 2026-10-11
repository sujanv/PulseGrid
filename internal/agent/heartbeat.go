package agent

import (
	"sync"
	"time"
)

type AgentInfo struct {
	ID            string    `json:"id"`
	Region        string    `json:"region"`
	Version       string    `json:"version"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	ActiveJobs    int       `json:"active_jobs"`
}

type Coordinator struct {
	mu     sync.RWMutex
	agents map[string]*AgentInfo
	queue  *RegionalQueue
}

func NewCoordinator() *Coordinator {
	return &Coordinator{
		agents: make(map[string]*AgentInfo),
		queue:  NewRegionalQueue(),
	}
}

func (c *Coordinator) RegisterAgent(info *AgentInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()
	info.LastHeartbeat = time.Now().UTC()
	c.agents[info.ID] = info
}

func (c *Coordinator) Heartbeat(agentID string, activeJobs int) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	a, ok := c.agents[agentID]
	if !ok {
		return false
	}
	a.LastHeartbeat = time.Now().UTC()
	a.ActiveJobs = activeJobs
	return true
}

func (c *Coordinator) ActiveAgents() []*AgentInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()
	cutoff := time.Now().UTC().Add(-60 * time.Second)

	var list []*AgentInfo
	for _, a := range c.agents {
		if a.LastHeartbeat.After(cutoff) {
			list = append(list, a)
		}
	}
	return list
}
