package alert

import (
	"context"
	"log"
	"math"
	"sync"
	"time"

	"github.com/sujanv/PulseGrid/internal/models"
)

type Channel interface {
	Send(ctx context.Context, event *models.AlertEvent) error
	Type() models.ChannelType
}

type Dispatcher struct {
	channels   []Channel
	queue      chan *models.AlertEvent
	maxRetries int
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewDispatcher(workers int, queueSize int, maxRetries int) *Dispatcher {
	if workers <= 0 {
		workers = 5
	}
	if queueSize <= 0 {
		queueSize = 500
	}
	if maxRetries <= 0 {
		maxRetries = 3
	}

	ctx, cancel := context.WithCancel(context.Background())
	d := &Dispatcher{
		queue:      make(chan *models.AlertEvent, queueSize),
		maxRetries: maxRetries,
		ctx:        ctx,
		cancel:     cancel,
	}

	for i := 0; i < workers; i++ {
		d.wg.Add(1)
		go d.worker()
	}

	return d
}

func (d *Dispatcher) RegisterChannel(c Channel) {
	d.channels = append(d.channels, c)
}

func (d *Dispatcher) Dispatch(event *models.AlertEvent) {
	select {
	case d.queue <- event:
	default:
		log.Printf("[AlertDispatcher] Warning: Alert queue full, dropping alert for %s", event.TargetName)
	}
}

func (d *Dispatcher) worker() {
	defer d.wg.Done()
	for {
		select {
		case <-d.ctx.Done():
			return
		case event, ok := <-d.queue:
			if !ok {
				return
			}
			d.broadcast(event)
		}
	}
}

func (d *Dispatcher) broadcast(event *models.AlertEvent) {
	for _, ch := range d.channels {
		go d.sendWithRetry(ch, event)
	}
}

func (d *Dispatcher) sendWithRetry(ch Channel, event *models.AlertEvent) {
	var err error
	for attempt := 0; attempt < d.maxRetries; attempt++ {
		ctx, cancel := context.WithTimeout(d.ctx, 10*time.Second)
		err = ch.Send(ctx, event)
		cancel()

		if err == nil {
			return
		}

		backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
		select {
		case <-d.ctx.Done():
			return
		case <-time.After(backoff):
		}
	}
	log.Printf("[AlertDispatcher] Failed to send alert to %s after %d retries: %v", ch.Type(), d.maxRetries, err)
}

func (d *Dispatcher) Stop() {
	d.cancel()
	close(d.queue)
	d.wg.Wait()
}
