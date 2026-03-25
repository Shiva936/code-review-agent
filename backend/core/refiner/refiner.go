package refiner

import (
	"strings"

	"github.com/Shiva936/code-review-agent/backend/config"
)

// ruleMap defines the specific rules for each weakness category
var ruleMap = map[string]string{
	"specificity":   "Always reference variable names and code lines",
	"actionability": "Each comment must include a clear fix",
	"severity":      "Assign correct severity labels",
	"structure":     "Organize output into categories",
}

// Refine updates the prompt based on identified weaknesses and manages rules.
func Refine(cfg *config.Config, prompt string, weakness string, existingRules []string) (string, []string) {
	// Get the rule for this weakness category
	rule, exists := ruleMap[weakness]
	if !exists {
		// If weakness is not recognized, return unchanged
		return prompt, existingRules
	}

	// Check if this rule already exists
	for _, existingRule := range existingRules {
		if strings.TrimSpace(existingRule) == strings.TrimSpace(rule) {
			// Rule already exists, return unchanged
			return prompt, existingRules
		}
	}

	// Add the new rule to the rules list
	updatedRules := make([]string, len(existingRules)+1)
	copy(updatedRules, existingRules)
	updatedRules[len(existingRules)] = rule

	// Update the prompt with the new rule
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
