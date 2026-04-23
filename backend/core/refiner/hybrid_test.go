package refiner

import (
	"testing"

	"github.com/Shiva936/code-review-agent/backend/config"
)

func TestApplyDeltaWithValidation_AcceptsBoundedDelta(t *testing.T) {
	cfg := &config.Config{}
	cfg.Refiner.MaxRules = 3
	cfg.Refiner.MaxRuleChars = 120
	cfg.Refiner.MaxDeltaOps = 4

	oldRules := []string{"Strengthen error handling in critical paths", "Prefer explicit variable names for readability"}
	delta := &PromptDelta{
		AddRules: []string{"Require concrete fix steps with affected symbol names"},
		Reason:   "add focused guidance",
	}

	newRules, status, err := applyDeltaWithValidation(cfg, oldRules, delta)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if status != "accepted" {
		t.Fatalf("expected accepted status, got %s", status)
	}
	if len(newRules) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(newRules))
	}
}

func TestApplyDeltaWithValidation_RejectsNoop(t *testing.T) {
	cfg := &config.Config{}
	cfg.Refiner.MaxRules = 3
	cfg.Refiner.MaxRuleChars = 120
	cfg.Refiner.MaxDeltaOps = 4

	oldRules := []string{"Require concrete fix steps with affected symbol names"}
	delta := &PromptDelta{
		AddRules: []string{"Require concrete fix steps with affected symbol names"},
		Reason:   "duplicate",
	}

	_, status, err := applyDeltaWithValidation(cfg, oldRules, delta)
	if err == nil {
		t.Fatalf("expected error for noop delta")
	}
	if status != "rejected_noop" {
		t.Fatalf("expected rejected_noop, got %s", status)
	}
}
