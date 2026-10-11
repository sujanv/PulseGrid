package main

import (
	"context"
	"fmt"
	"time"

	"github.com/sujanv/PulseGrid/internal/config"
	"github.com/sujanv/PulseGrid/internal/storage/postgres"
)

func runTUI() {
	cfg := config.Load()
	db, err := postgres.Connect(cfg.Database)
	if err != nil {
		fmt.Printf("Database connection failed: %v\n", err)
		return
	}
	defer db.Close()

	targetRepo := postgres.NewTargetRepository(db)
	resultRepo := postgres.NewResultRepository(db)
	ctx := context.Background()

	// Clear screen
	fmt.Print("\033[2J\033[H")

	for {
		fmt.Print("\033[H") // move cursor home
		fmt.Println("================================================================================")
		fmt.Println(" ⚡ PULSEGRID REAL-TIME MONITORING TUI                                            ")
		fmt.Printf(" Time: %s | Refresh: 3s                                  \n", time.Now().Format("15:04:05"))
		fmt.Println("================================================================================")
		fmt.Printf("%-24s %-8s %-12s %-14s %-12s\n", "TARGET", "TYPE", "STATUS", "LATENCY", "UPTIME (24h)")
		fmt.Println("────────────────────────────────────────────────────────────────────────────────")

		targets, _ := targetRepo.ListAll(ctx)
		for _, t := range targets {
			uptime, _ := resultRepo.CalculateUptime(ctx, t.ID, 24*time.Hour)
			stats, _ := resultRepo.GetLatencyStats(ctx, t.ID, 24*time.Hour)
			avgLat := 0.0
			if stats != nil {
				avgLat = stats.AvgMs
			}

			statusStr := "\033[32mUP\033[0m"
			if !t.Active {
				statusStr = "\033[90mPAUSED\033[0m"
			}

			name := t.Name
			if len(name) > 23 {
				name = name[:20] + "..."
			}

			fmt.Printf("%-24s %-8s %-20s %-14s %-12s\n",
				name, t.Type, statusStr, fmt.Sprintf("%.2f ms", avgLat), fmt.Sprintf("%.2f%%", uptime))
		}
		fmt.Println("================================================================================")
		fmt.Println("Press Ctrl+C to exit interactive mode.")

		time.Sleep(3 * time.Second)
	}
}
