package alert

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/sujanv/PulseGrid/internal/models"
)

type FailureDomain struct {
	Provider string `json:"provider"` // e.g. "AWS", "GCP", "Cloudflare"
	Region   string `json:"region"`   // e.g. "us-east-1"
	Zone     string `json:"zone"`     // e.g. "us-east-1a"
	Service  string `json:"service"`  // e.g. "ingress-gateway"
}

type CorrelatedAlertGroup struct {
	Fingerprint string        `json:"fingerprint"`
	Domain      FailureDomain `json:"domain"`
	TargetIDs   []string      `json:"target_ids"`
	FirstSeen   time.Time     `json:"first_seen"`
	LastSeen    time.Time     `json:"last_seen"`
	Count       int           `json:"count"`
	IsCluster   bool          `json:"is_cluster"`
}

type AlertDeduplicator struct {
	mu            sync.Mutex
	window        time.Duration
	groups        map[string]*CorrelatedAlertGroup
	clusterMinCnt int
}

func NewAlertDeduplicator(window time.Duration, clusterMinCount int) *AlertDeduplicator {
	if window <= 0 {
		window = 5 * time.Minute
	}
	if clusterMinCount <= 1 {
		clusterMinCount = 3
	}
	return &AlertDeduplicator{
		window:        window,
		groups:        make(map[string]*CorrelatedAlertGroup),
		clusterMinCnt: clusterMinCount,
	}
}

func (d *AlertDeduplicator) Deduplicate(event *models.AlertEvent, domain FailureDomain) (isNewOrEscalated bool, group *CorrelatedAlertGroup) {
	d.mu.Lock()
	defer d.mu.Unlock()

	fp := d.computeFingerprint(domain)
	now := time.Now().UTC()

	grp, exists := d.groups[fp]
	if !exists || now.Sub(grp.LastSeen) > d.window {
		// New alert group window
		newGrp := &CorrelatedAlertGroup{
			Fingerprint: fp,
			Domain:      domain,
			TargetIDs:   []string{event.TargetID},
			FirstSeen:   now,
			LastSeen:    now,
			Count:       1,
			IsCluster:   false,
		}
		d.groups[fp] = newGrp
		return true, newGrp
	}

	// Existing alert group within sliding window
	grp.LastSeen = now
	grp.Count++
	found := false
	for _, id := range grp.TargetIDs {
		if id == event.TargetID {
			found = true
			break
		}
	}
	if !found {
		grp.TargetIDs = append(grp.TargetIDs, event.TargetID)
	}

	// If threshold reached, escalate as correlated cluster incident
	if grp.Count >= d.clusterMinCnt && !grp.IsCluster {
		grp.IsCluster = true
		return true, grp
	}

	return false, grp
}

func (d *AlertDeduplicator) computeFingerprint(domain FailureDomain) string {
	raw := fmt.Sprintf("%s:%s:%s:%s", domain.Provider, domain.Region, domain.Zone, domain.Service)
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:8])
}
