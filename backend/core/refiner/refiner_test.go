package refiner

import (
	"strings"
	"testing"

	"github.com/Shiva936/code-review-agent/backend/config"
)

func TestRefine_RepeatingWeakness_AddsReinforcement(t *testing.T) {
	cfg := &config.Config{}
	base := "You are a reviewer."
	rules := []string{"Strengthen security analysis: injection, secrets, authn/z, and unsafe APIs."}
	rule, _ := ruleMap["security"]

	p2, r2 := Refine(cfg, base+"\n\nAdditional Rules:\n- "+rule, "security", rules, 2)
	if len(r2) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(r2))
	}
	if !strings.Contains(p2, "Reinforcement (iteration 2)") {
		t.Fatalf("expected reinforcement in prompt, got: %s", p2)
	}
}

func TestRefine_NewWeakness_AppendsRule(t *testing.T) {
	cfg := &config.Config{}
	p, rules := Refine(cfg, "base", "logic", nil, 1)
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if !strings.Contains(p, "Additional Rules") {
		t.Fatalf("expected rules section: %s", p)
	}
}
