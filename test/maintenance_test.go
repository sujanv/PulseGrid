package test

import (
	"testing"
	"time"

	"github.com/sujanv/PulseGrid/internal/maintenance"
)

func TestMaintenanceWindowManager(t *testing.T) {
	mgr := maintenance.NewManager()

	start := time.Now().Add(-1 * time.Hour)
	end := time.Now().Add(1 * time.Hour)

	win := mgr.AddWindow("Database Upgrade", "Scheduled patching", []string{"tgt-db"}, start, end)
	if win == nil {
		t.Fatalf("expected created window")
	}

	// Should be active now for tgt-db
	if !mgr.IsUnderMaintenance("tgt-db", time.Now()) {
		t.Errorf("expected tgt-db to be under maintenance")
	}

	// Should not apply to tgt-api
	if mgr.IsUnderMaintenance("tgt-api", time.Now()) {
		t.Errorf("tgt-api should not be under maintenance")
	}

	// Remove window
	if !mgr.RemoveWindow(win.ID) {
		t.Errorf("expected window removal to succeed")
	}

	if mgr.IsUnderMaintenance("tgt-db", time.Now()) {
		t.Errorf("expected tgt-db maintenance to be inactive after removal")
	}
}
