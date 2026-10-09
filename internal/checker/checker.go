package checker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sujanv/PulseGrid/internal/models"
)

type Checker interface {
	Check(ctx context.Context, target *models.Target) (*models.CheckResult, error)
}

type Registry struct {
	mu       sync.RWMutex
	checkers map[models.TargetType]Checker
}

func NewRegistry() *Registry {
	return &Registry{
		checkers: make(map[models.TargetType]Checker),
	}
}

func (r *Registry) Register(targetType models.TargetType, c Checker) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.checkers[targetType] = c
}

func (r *Registry) Get(targetType models.TargetType) (Checker, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.checkers[targetType]
	return c, ok
}

func (r *Registry) Execute(ctx context.Context, target *models.Target) (*models.CheckResult, error) {
	c, ok := r.Get(target.Type)
	if !ok {
		return &models.CheckResult{
			ID:           uuid.New().String(),
			TargetID:     target.ID,
			Timestamp:    time.Now().UTC(),
			Status:       models.StatusDown,
			ErrorMessage: fmt.Sprintf("unsupported target type: %s", target.Type),
		}, nil
	}

	checkCtx, cancel := context.WithTimeout(ctx, target.Timeout)
	defer cancel()

	start := time.Now()
	res, err := c.Check(checkCtx, target)
	if err != nil {
		return &models.CheckResult{
			ID:           uuid.New().String(),
			TargetID:     target.ID,
			Timestamp:    time.Now().UTC(),
			Status:       models.StatusDown,
			Latency:      time.Since(start),
			ErrorMessage: err.Error(),
		}, nil
	}

	if res.ID == "" {
		res.ID = uuid.New().String()
	}
	if res.TargetID == "" {
		res.TargetID = target.ID
	}
	if res.Timestamp.IsZero() {
		res.Timestamp = time.Now().UTC()
	}

	return res, nil
}
