package router

import (
	"math"

	"github.com/Shiva936/code-review-agent/backend/models"
)

// aggregateAcrossSamples computes mean total score and the issue category with lowest
// mean score across snippets (logic, performance, security, style). Tie-break: lexicographic name.
func aggregateAcrossSamples(results []*models.EvalResult) (avgTotal int, weakestIssueCategory string) {
	if len(results) == 0 {
		return 0, "logic"
	}

	sumTotal := 0
	for _, r := range results {
		sumTotal += r.Total
	}
	avgTotal = int(math.Round(float64(sumTotal) / float64(len(results))))

	n := float64(len(results))
	sumL, sumP, sumS, sumY := 0, 0, 0, 0
	for _, r := range results {
		sumL += r.Logic
		sumP += r.Performance
		sumS += r.Security
		sumY += r.Style
	}
	avgs := map[string]float64{
		"logic":       float64(sumL) / n,
		"performance": float64(sumP) / n,
		"security":    float64(sumS) / n,
		"style":       float64(sumY) / n,
	}

	order := []string{"logic", "performance", "security", "style"}
	minV := 999.0
	weakestIssueCategory = "logic"
	for _, name := range order {
		v := avgs[name]
		if v < minV || (v == minV && name < weakestIssueCategory) {
			minV = v
			weakestIssueCategory = name
		}
	}

	return avgTotal, weakestIssueCategory
}
