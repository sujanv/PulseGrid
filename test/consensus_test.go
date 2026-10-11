package test

import (
	"testing"
	"time"

	"github.com/sujanv/PulseGrid/internal/consensus"
	"github.com/sujanv/PulseGrid/internal/models"
)

func TestQuorumConsensus(t *testing.T) {
	q := consensus.NewQuorumConsensus(60*time.Second, 0.60) // 60% majority

	targetID := "tgt-global"

	// 1st region says DOWN (1/1 = 100% -> confirmed)
	q.Record(targetID, "us-east-1", models.StatusDown)
	if !q.IsOutageConfirmed(targetID) {
		t.Errorf("1/1 down should confirm outage")
	}

	// 2nd region says UP (1/2 = 50% < 60% -> NOT confirmed)
	q.Record(targetID, "eu-central-1", models.StatusUp)
	if q.IsOutageConfirmed(targetID) {
		t.Errorf("1/2 down should not meet 60%% quorum")
	}

	// 3rd region says DOWN (2/3 = 66.7% >= 60% -> confirmed)
	q.Record(targetID, "ap-southeast-1", models.StatusDown)
	if !q.IsOutageConfirmed(targetID) {
		t.Errorf("2/3 down should confirm outage")
	}
}
