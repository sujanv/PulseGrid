package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/sujanv/PulseGrid/internal/models"
	"github.com/sujanv/PulseGrid/internal/storage/redis"
)

type DistributedCoordinator struct {
	scheduler *Scheduler
	locker    redis.Locker
	nodeID    string
}

func NewDistributedCoordinator(s *Scheduler, locker redis.Locker, nodeID string) *DistributedCoordinator {
	return &DistributedCoordinator{
		scheduler: s,
		locker:    locker,
		nodeID:    nodeID,
	}
}

func (d *DistributedCoordinator) TryDispatch(ctx context.Context, target *models.Target) (bool, error) {
	if d.locker == nil {
		return true, nil
	}

	lockResource := fmt.Sprintf("target-check:%s", target.ID)
	ttl := target.Interval / 2
	if ttl < 2*time.Second {
		ttl = 2 * time.Second
	}

	acquired, err := d.locker.Acquire(ctx, lockResource, ttl)
	if err != nil {
		return false, err
	}
	return acquired, nil
}
