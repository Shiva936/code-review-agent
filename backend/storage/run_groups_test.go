package storage

import (
	"path/filepath"
	"testing"

	"github.com/Shiva936/code-review-agent/backend/config"
)

func TestRunGroups_CRUD(t *testing.T) {
	cfg := &config.Config{
		DatabasePath: filepath.Join(t.TempDir(), "test.db"),
	}
	if err := InitDB(cfg); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	t.Cleanup(func() {
		_ = Close()
	})

	groupID, err := CreateRunGroup("code", "prompt", 5)
	if err != nil {
		t.Fatalf("CreateRunGroup failed: %v", err)
	}
	if groupID <= 0 {
		t.Fatalf("expected groupID > 0, got %d", groupID)
	}

	if err := SaveRunGroupRun(groupID, 1, 9, "actionability"); err != nil {
		t.Fatalf("SaveRunGroupRun failed: %v", err)
	}

	count, err := GetRunGroupsCount()
	if err != nil {
		t.Fatalf("GetRunGroupsCount failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected count=1, got %d", count)
	}

	groups, err := GetRunGroups(10, 0)
	if err != nil {
		t.Fatalf("GetRunGroups failed: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}

	runs, err := GetRunGroupRuns(groupID)
	if err != nil {
		t.Fatalf("GetRunGroupRuns failed: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
}
