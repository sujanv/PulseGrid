package k8s

import (
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/sujanv/PulseGrid/internal/models"
)

type IngressAnnotationParser struct{}

func NewIngressAnnotationParser() *IngressAnnotationParser {
	return &IngressAnnotationParser{}
}

func (p *IngressAnnotationParser) ParseAnnotations(name string, host string, annotations map[string]string) (*models.Target, bool) {
	enabled := annotations["pulsegrid.io/monitor"]
	if enabled != "true" && enabled != "yes" {
		return nil, false
	}

	intervalSec := 30
	if invStr := annotations["pulsegrid.io/interval"]; invStr != "" {
		if inv, err := strconv.Atoi(invStr); err == nil && inv > 0 {
			intervalSec = inv
		}
	}

	path := annotations["pulsegrid.io/path"]
	if path == "" {
		path = "/healthz"
	}

	url := fmt.Sprintf("https://%s%s", host, path)

	target := &models.Target{
		ID:                  uuid.New().String(),
		Name:                fmt.Sprintf("k8s-%s", name),
		URL:                 url,
		Type:                models.TargetHTTPS,
		Interval:            time.Duration(intervalSec) * time.Second,
		Timeout:             10 * time.Second,
		ExpectedStatusCodes: []int{200, 204},
		FailureThreshold:    3,
		SuccessThreshold:    2,
		Active:              true,
		CreatedAt:           time.Now().UTC(),
		UpdatedAt:           time.Now().UTC(),
	}

	return target, true
}
