package agent

import (
	"sync"
)

type RegionalQueue struct {
	mu     sync.Mutex
	queues map[string][]CheckJob // region -> jobs
}

func NewRegionalQueue() *RegionalQueue {
	return &RegionalQueue{
		queues: make(map[string][]CheckJob),
	}
}

func (q *RegionalQueue) Enqueue(region string, job CheckJob) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.queues[region] = append(q.queues[region], job)
}

func (q *RegionalQueue) Dequeue(region string, maxCount int) []CheckJob {
	q.mu.Lock()
	defer q.mu.Unlock()

	items := q.queues[region]
	if len(items) == 0 {
		return nil
	}

	if len(items) <= maxCount {
		q.queues[region] = nil
		return items
	}

	batch := items[:maxCount]
	q.queues[region] = items[maxCount:]
	return batch
}
