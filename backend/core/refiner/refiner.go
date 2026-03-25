package refiner

import (
	"fmt"
	"strings"

	"github.com/Shiva936/code-review-agent/backend/config"
)

// ruleMap defines refinement rules. Issue categories (spec): logic, performance, security, style.
// Rubric dimensions (legacy): actionability, specificity, severity.
var ruleMap = map[string]string{
	"logic":         "Strengthen logic analysis: control flow, edge cases, correctness, and error paths.",
	"performance":   "Strengthen performance analysis: allocations, complexity, I/O, and hot paths.",
	"security":      "Strengthen security analysis: injection, secrets, authn/z, and unsafe APIs.",
	"style":         "Strengthen style and maintainability: naming, structure, consistency, and readability.",
	"specificity":   "Always reference variable names and code lines",
	"actionability": "Each comment must include a clear fix",
	"severity":      "Assign correct severity labels",
	"structure":     "Organize output into categories",
}

// Refine updates the prompt based on identified weaknesses and manages rules.
// If the same weakness repeats, an escalating reinforcement line is added so the prompt
// does not stall across iterations.
func Refine(cfg *config.Config, prompt string, weakness string, existingRules []string, iteration int) (string, []string) {
	_ = cfg
	rule, exists := ruleMap[weakness]
	if !exists {
		return prompt, existingRules
	}

	for _, existingRule := range existingRules {
		if strings.TrimSpace(existingRule) == strings.TrimSpace(rule) {
			reinforcement := fmt.Sprintf("Reinforcement (iteration %d): apply %s feedback more strictly than the previous review.", iteration, weakness)
			updatedRules := append(existingRules, reinforcement)
			return addRuleToPrompt(prompt, reinforcement), updatedRules
		}
	}

	updatedRules := make([]string, len(existingRules)+1)
	copy(updatedRules, existingRules)
	updatedRules[len(existingRules)] = rule

	refinedPrompt := addRuleToPrompt(prompt, rule)
	return refinedPrompt, updatedRules
}

// addRuleToPrompt adds a new rule to the prompt in a clean way
func addRuleToPrompt(prompt string, rule string) string {
	// Look for existing rules section or add one
	rulesSection := "\n\nAdditional Rules:"
	if strings.Contains(prompt, rulesSection) {
		// Add to existing rules section
		return strings.Replace(prompt, rulesSection, rulesSection+"\n- "+rule, 1)
	} else {
		// Add new rules section at the end
		return prompt + rulesSection + "\n- " + rule
	}
}
