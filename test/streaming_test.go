package test

import (
	"testing"
	"time"

	"github.com/sujanv/PulseGrid/internal/streaming"
)

func TestEventBusPubSub(t *testing.T) {
	eb := streaming.NewEventBus()
	sub := eb.Subscribe(streaming.EventCheckResult)

	ev := streaming.Event{
		Type:    streaming.EventCheckResult,
		Payload: "payload-test-1",
	}

	eb.Publish(ev)

	select {
	case received := <-sub:
		if received.Payload != "payload-test-1" {
			t.Errorf("payload mismatch: %v", received.Payload)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("timed out waiting for event bus message")
	}
}
