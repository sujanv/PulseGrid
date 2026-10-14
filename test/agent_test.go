package test

import (
	"testing"

	"github.com/sujanv/PulseGrid/internal/agent"
	"github.com/sujanv/PulseGrid/internal/models"
)

func TestRegionalQueue(t *testing.T) {
	q := agent.NewRegionalQueue()

	q.Enqueue("us-east", agent.CheckJob{JobID: "job-1", Target: &models.Target{Name: "T1"}})
	q.Enqueue("us-east", agent.CheckJob{JobID: "job-2", Target: &models.Target{Name: "T2"}})
	q.Enqueue("eu-west", agent.CheckJob{JobID: "job-3", Target: &models.Target{Name: "T3"}})

	batch := q.Dequeue("us-east", 10)
	if len(batch) != 2 {
		t.Errorf("expected 2 jobs for us-east, got %d", len(batch))
	}

	emptyBatch := q.Dequeue("us-east", 10)
	if len(emptyBatch) != 0 {
		t.Errorf("expected empty batch on second dequeue")
	}

	euBatch := q.Dequeue("eu-west", 1)
	if len(euBatch) != 1 {
		t.Errorf("expected 1 job for eu-west")
	}
}

func TestCoordinatorHeartbeat(t *testing.T) {
	coord := agent.NewCoordinator()
	info := &agent.AgentInfo{
		ID:      "agent-nyc-1",
		Region:  "us-east",
		Version: "1.0.0",
	}

	coord.RegisterAgent(info)
	active := coord.ActiveAgents()
	if len(active) != 1 {
		t.Fatalf("expected 1 active agent")
	}

	ok := coord.Heartbeat("agent-nyc-1", 5)
	if !ok {
		t.Errorf("expected heartbeat to succeed")
	}
}
