package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/sujanv/PulseGrid/internal/checker"
	"github.com/sujanv/PulseGrid/internal/config"
	"github.com/sujanv/PulseGrid/internal/models"
	"github.com/sujanv/PulseGrid/internal/storage/postgres"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]
	switch command {
	case "check":
		if len(os.Args) < 3 {
			fmt.Println("Usage: pulsegrid-cli check <url>")
			os.Exit(1)
		}
		runCheck(os.Args[2])
	case "seed":
		runSeed()
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("PulseGrid CLI - Uptime Monitoring & Diagnostics")
	fmt.Println("Usage:")
	fmt.Println("  pulsegrid-cli check <url>   - Perform one-off check with latency breakdown")
	fmt.Println("  pulsegrid-cli seed          - Populate PostgreSQL with sample targets")
}

func runCheck(targetURL string) {
	fmt.Printf("🔍 Probing endpoint: %s\n", targetURL)
	c := checker.NewHTTPChecker(10 * time.Second)
	target := &models.Target{
		ID:                  uuid.New().String(),
		Name:                "CLI-Check",
		URL:                 targetURL,
		Type:                models.TargetHTTPS,
		Timeout:             10 * time.Second,
		ExpectedStatusCodes: []int{200, 201, 204, 301, 302},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := c.Check(ctx, target)
	if err != nil {
		fmt.Printf("❌ Check failed with error: %v\n", err)
		return
	}

	fmt.Println("──────────────────────────────────────────────────")
	fmt.Printf("Status:       %s\n", res.Status)
	fmt.Printf("HTTP Status:  %d\n", res.StatusCode)
	fmt.Printf("Total Latency: %v\n", res.Latency)
	fmt.Printf("DNS Lookup:    %v\n", res.Breakdown.DNSLookup)
	fmt.Printf("TCP Connect:   %v\n", res.Breakdown.TCPConnect)
	fmt.Printf("TLS Handshake: %v\n", res.Breakdown.TLSHandshake)
	fmt.Printf("TTFB:          %v\n", res.Breakdown.TTFB)
	fmt.Printf("Transfer:      %v\n", res.Breakdown.ContentTransfer)
	fmt.Printf("Payload Size:  %d bytes\n", res.ResponseBytes)
	fmt.Println("──────────────────────────────────────────────────")
}

func runSeed() {
	cfg := config.Load()
	db, err := postgres.Connect(cfg.Database)
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		return
	}
	defer db.Close()

	repo := postgres.NewTargetRepository(db)
	ctx := context.Background()

	samples := []*models.Target{
		{
			ID: uuid.New().String(), Name: "Google DNS Primary", URL: "8.8.8.8", Host: "8.8.8.8", Port: 53,
			Type: models.TargetDNS, Interval: 30 * time.Second, Timeout: 5 * time.Second,
			FailureThreshold: 3, SuccessThreshold: 2, Active: true, CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
		{
			ID: uuid.New().String(), Name: "Cloudflare 1.1.1.1", URL: "https://1.1.1.1",
			Type: models.TargetHTTPS, Interval: 15 * time.Second, Timeout: 5 * time.Second,
			ExpectedStatusCodes: []int{200}, FailureThreshold: 3, SuccessThreshold: 2, Active: true, CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
		{
			ID: uuid.New().String(), Name: "GitHub API Status", URL: "https://api.github.com/zen",
			Type: models.TargetHTTPS, Interval: 30 * time.Second, Timeout: 10 * time.Second,
			ExpectedStatusCodes: []int{200}, FailureThreshold: 3, SuccessThreshold: 2, Active: true, CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
		{
			ID: uuid.New().String(), Name: "Wikipedia TLS Cert", URL: "https://www.wikipedia.org", Host: "www.wikipedia.org", Port: 443,
			Type: models.TargetTLS, Interval: 60 * time.Second, Timeout: 10 * time.Second,
			FailureThreshold: 2, SuccessThreshold: 1, Active: true, CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
	}

	for _, s := range samples {
		if err := repo.Create(ctx, s); err != nil {
			fmt.Printf("Failed to seed %s: %v\n", s.Name, err)
		} else {
			fmt.Printf("✅ Seeded monitor: %s (%s)\n", s.Name, s.Type)
		}
	}
	fmt.Println("🎉 Database seeded successfully!")
}
