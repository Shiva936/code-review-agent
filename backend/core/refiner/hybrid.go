package refiner

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Shiva936/code-review-agent/backend/config"
	"github.com/Shiva936/code-review-agent/backend/llm"
)

type RuleModification struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type PromptDelta struct {
	AddRules    []string           `json:"add_rules"`
	RemoveRules []string           `json:"remove_rules"`
	ModifyRules []RuleModification `json:"modify_rules"`
	Reason      string             `json:"reason"`
}

type RefineDecision struct {
	Prompt           string
	Rules            []string
	Source           string
	Reason           string
	InputJSON        string
	RawOutput        string
	DeltaJSON        string
	ValidationStatus string
	Applied          bool
}

type refinerInput struct {
	WeakestCategory      string   `json:"weakest_category"`
	Iteration            int      `json:"iteration"`
	MaxRules             int      `json:"max_rules"`
	MaxRuleChars         int      `json:"max_rule_chars"`
	ActiveRules          []string `json:"active_rules"`
	RecentEvaluatorNotes string   `json:"recent_evaluator_feedback"`
}

func RefineWithGuardrails(cfg *config.Config, basePrompt string, weakness string, existingRules []string, iteration int, recentEvalFeedback string) RefineDecision {
	mode := strings.ToLower(strings.TrimSpace(cfg.Refiner.Mode))
	if mode == "" || mode == "rule_based" {
		return fallbackRuleBased(cfg, basePrompt, weakness, existingRules, iteration, "rule_based_mode")
	}

	input := buildRefinerInput(cfg, weakness, existingRules, iteration, recentEvalFeedback)
	inputBytes, _ := json.Marshal(input)

	raw, err := callRefinerLLM(cfg, input)
	if err != nil {
		d := fallbackRuleBased(cfg, basePrompt, weakness, existingRules, iteration, "llm_call_failed")
		d.InputJSON = string(inputBytes)
		d.RawOutput = err.Error()
		d.ValidationStatus = "fallback"
		return d
	}

	delta, deltaJSON, err := parsePromptDelta(raw)
	if err != nil {
		d := fallbackRuleBased(cfg, basePrompt, weakness, existingRules, iteration, "delta_parse_failed")
		d.InputJSON = string(inputBytes)
		d.RawOutput = raw
		d.ValidationStatus = "fallback"
		return d
	}

	newRules, validationStatus, validationErr := applyDeltaWithValidation(cfg, existingRules, delta)
	if validationErr != nil {
		d := fallbackRuleBased(cfg, basePrompt, weakness, existingRules, iteration, validationErr.Error())
		d.InputJSON = string(inputBytes)
		d.RawOutput = raw
		d.DeltaJSON = deltaJSON
		d.ValidationStatus = "fallback"
		return d
	}

	return RefineDecision{
		Prompt:           BuildPrompt(basePrompt, newRules),
		Rules:            newRules,
		Source:           "llm_delta",
		Reason:           strings.TrimSpace(delta.Reason),
		InputJSON:        string(inputBytes),
		RawOutput:        raw,
		DeltaJSON:        deltaJSON,
		ValidationStatus: validationStatus,
		Applied:          true,
	}
}

func BuildPrompt(basePrompt string, rules []string) string {
	out := strings.TrimSpace(basePrompt)
	normalized := normalizeRules(rules)
	if len(normalized) == 0 {
		return out
	}
	var b strings.Builder
	b.WriteString(out)
	b.WriteString("\n\nAdditional Rules:")
	for _, r := range normalized {
		b.WriteString("\n- ")
		b.WriteString(r)
	}
	return b.String()
}

func buildRefinerInput(cfg *config.Config, weakness string, rules []string, iteration int, recent string) refinerInput {
	maxRules := cfg.Refiner.MaxRules
	if maxRules <= 0 {
		maxRules = 8
	}
	maxRuleChars := cfg.Refiner.MaxRuleChars
	if maxRuleChars <= 0 {
		maxRuleChars = 200
	}
	limitedRules := normalizeRules(rules)
	if len(limitedRules) > maxRules {
		limitedRules = limitedRules[len(limitedRules)-maxRules:]
	}
	recent = strings.TrimSpace(recent)
	if len(recent) > 1200 {
		recent = recent[:1200]
	}
	return refinerInput{
		WeakestCategory:      weakness,
		Iteration:            iteration,
		MaxRules:             maxRules,
		MaxRuleChars:         maxRuleChars,
		ActiveRules:          limitedRules,
		RecentEvaluatorNotes: recent,
	}
}

func callRefinerLLM(cfg *config.Config, input refinerInput) (string, error) {
	model := strings.TrimSpace(cfg.Refiner.Model)
	if model == "" {
		model = cfg.EvaluatorModel
	}
	temp := cfg.Refiner.Temperature
	if temp <= 0 {
		temp = 0.2
	}
	payload, _ := json.Marshal(input)
	prompt := fmt.Sprintf(`You improve code-review policy rules using guarded deltas only.

Return JSON ONLY with this exact schema:
{
  "add_rules": ["string"],
  "remove_rules": ["string"],
  "modify_rules": [{"from":"string","to":"string"}],
  "reason": "string"
}

Hard constraints:
- Never return a full prompt.
- Keep changes high-signal and specific to weakest category.
- Keep rules concise and non-contradictory.
- If no safe improvement exists, return empty arrays and explain why in "reason".

Input:
%s`, string(payload))

	return llm.CallLLMWithOpts(cfg, "refine", prompt, model, &llm.CallOpts{
		Temperature: temp,
		TopP:        0.9,
	})
}

func parsePromptDelta(raw string) (*PromptDelta, string, error) {
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start == -1 || end == -1 || end <= start {
		return nil, "", fmt.Errorf("no_json")
	}
	body := raw[start : end+1]
	var d PromptDelta
	if err := json.Unmarshal([]byte(body), &d); err != nil {
		return nil, "", err
	}
	return &d, body, nil
}

func applyDeltaWithValidation(cfg *config.Config, existingRules []string, delta *PromptDelta) ([]string, string, error) {
	maxRules := cfg.Refiner.MaxRules
	if maxRules <= 0 {
		maxRules = 8
	}
	maxRuleChars := cfg.Refiner.MaxRuleChars
	if maxRuleChars <= 0 {
		maxRuleChars = 200
	}
	maxOps := cfg.Refiner.MaxDeltaOps
	if maxOps <= 0 {
		maxOps = 4
	}

	ops := len(delta.AddRules) + len(delta.RemoveRules) + len(delta.ModifyRules)
	if ops == 0 {
		return existingRules, "rejected_noop", fmt.Errorf("noop_delta")
	}
	if ops > maxOps {
		return existingRules, "rejected_budget", fmt.Errorf("too_many_delta_ops")
	}

	rules := append([]string{}, normalizeRules(existingRules)...)
	for _, r := range delta.RemoveRules {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}
		filtered := make([]string, 0, len(rules))
		for _, ex := range rules {
			if !strings.EqualFold(strings.TrimSpace(ex), r) {
				filtered = append(filtered, ex)
			}
		}
		rules = filtered
	}

	for _, m := range delta.ModifyRules {
		from := strings.TrimSpace(m.From)
		to := strings.TrimSpace(m.To)
		if from == "" || to == "" || len(to) > maxRuleChars {
			return existingRules, "rejected_invalid_modify", fmt.Errorf("invalid_modify_rule")
		}
		found := false
		for i, ex := range rules {
			if strings.EqualFold(strings.TrimSpace(ex), from) {
				rules[i] = to
				found = true
				break
			}
		}
		if !found {
			return existingRules, "rejected_missing_modify_target", fmt.Errorf("modify_target_not_found")
		}
	}

	for _, add := range delta.AddRules {
		add = strings.TrimSpace(add)
		if add == "" || len(add) > maxRuleChars || looksLowSignal(add) {
			return existingRules, "rejected_low_signal_add", fmt.Errorf("invalid_add_rule")
		}
		dup := false
		for _, ex := range rules {
			if strings.EqualFold(strings.TrimSpace(ex), add) {
				dup = true
				break
			}
		}
		if !dup {
			rules = append(rules, add)
		}
	}

	rules = normalizeRules(rules)
	if hasConflictingRules(rules) {
		return existingRules, "rejected_conflict", fmt.Errorf("conflicting_rules")
	}

	if len(rules) > maxRules {
		// Keep newest rules to enforce budget.
		rules = rules[len(rules)-maxRules:]
	}

	oldNorm := strings.Join(normalizeRules(existingRules), "\n")
	newNorm := strings.Join(rules, "\n")
	if oldNorm == newNorm {
		return existingRules, "rejected_noop", fmt.Errorf("no_effect_after_apply")
	}
	return rules, "accepted", nil
}

func normalizeRules(rules []string) []string {
	out := make([]string, 0, len(rules))
	seen := map[string]bool{}
	for _, r := range rules {
		s := strings.TrimSpace(r)
		if s == "" {
			continue
		}
		key := strings.ToLower(s)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, s)
	}
	return out
}

func looksLowSignal(s string) bool {
	low := strings.ToLower(strings.TrimSpace(s))
	return low == "improve quality" || low == "be specific" || len(low) < 12
}

func hasConflictingRules(rules []string) bool {
	hasAlways := false
	hasNever := false
	for _, r := range rules {
		l := strings.ToLower(r)
		if strings.Contains(l, "always") {
			hasAlways = true
		}
		if strings.Contains(l, "never") {
			hasNever = true
		}
	}
	return hasAlways && hasNever
}

func fallbackRuleBased(cfg *config.Config, basePrompt string, weakness string, existingRules []string, iteration int, reason string) RefineDecision {
	startPrompt := BuildPrompt(basePrompt, existingRules)
	prompt, rules := Refine(cfg, startPrompt, weakness, existingRules, iteration)
	return RefineDecision{
		Prompt:           prompt,
		Rules:            rules,
		Source:           "rule_based",
		Reason:           reason,
		ValidationStatus: "fallback",
		Applied:          true,
	}
}
