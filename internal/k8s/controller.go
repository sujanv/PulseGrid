package k8s

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sujanv/PulseGrid/internal/models"
	"github.com/sujanv/PulseGrid/internal/scheduler"
)

type Controller struct {
	scheduler *scheduler.Scheduler
	mu        sync.RWMutex
	targets   map[string]string // "ns/name" -> targetID
}

func NewController(s *scheduler.Scheduler) *Controller {
	return &Controller{
		scheduler: s,
		targets:   make(map[string]string),
	}
}

func (c *Controller) Reconcile(pt *ProbeTarget) *models.Target {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := pt.Metadata.Namespace + "/" + pt.Metadata.Name
	targetID, exists := c.targets[key]
	if !exists {
		targetID = uuid.New().String()
		c.targets[key] = targetID
	}

	interval := time.Duration(pt.Spec.IntervalSeconds) * time.Second
	if interval < 5*time.Second {
		interval = 30 * time.Second
	}

	target := &models.Target{
		ID:                  targetID,
		Name:                "k8s-" + pt.Metadata.Name,
		URL:                 pt.Spec.URL,
		Type:                models.TargetHTTPS,
		Interval:            interval,
		Timeout:             time.Duration(pt.Spec.TimeoutSeconds) * time.Second,
		ExpectedStatusCodes: pt.Spec.ExpectedStatusCodes,
		FailureThreshold:    pt.Spec.FailureThreshold,
		Active:              pt.Spec.Active,
		CreatedAt:           time.Now().UTC(),
		UpdatedAt:           time.Now().UTC(),
	}

	if c.scheduler != nil {
		c.scheduler.RegisterTarget(target)
	}
	return target
}
