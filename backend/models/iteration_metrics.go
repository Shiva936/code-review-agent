package models

import (
	"encoding/json"
	"math"
)

// SampleEvalMetrics is one hardcoded sample's evaluation (for API/UI breakdown).
type SampleEvalMetrics struct {
	Index            int    `json:"index"`
	Total            int    `json:"total"`
	Actionability    int    `json:"actionability"`
	Specificity      int    `json:"specificity"`
	Severity         int    `json:"severity"`
	WeaknessCategory string `json:"weakness_category"`
	Logic            int    `json:"logic"`
	Performance      int    `json:"performance"`
	Security         int    `json:"security"`
	Style            int    `json:"style"`
}

// IterationMetrics is stored in run_group_runs.detail_json (means across 3 samples per iteration).
type IterationMetrics struct {
	Actionability        int `json:"actionability"`
	Specificity          int `json:"specificity"`
	Severity             int `json:"severity"` // rubric: correct severity labels
	Structure            int `json:"structure"` // mean style score (UI “Structure” column)
	Logic                int `json:"logic"`
	Performance          int `json:"performance"`
	Security             int `json:"security"`
	WeakestIssueCategory string `json:"weakest_issue_category"`
	Samples              []SampleEvalMetrics `json:"samples,omitempty"`
}

// BuildIterationMetrics computes rounded means across samples and JSON-marshals for DB storage.
func BuildIterationMetrics(results []*EvalResult, weakestIssueCategory string) ([]byte, error) {
	n := len(results)
	if n == 0 {
		return []byte("{}"), nil
	}

	var sumA, sumSp, sumSev, sumL, sumP, sumS, sumY int
	samples := make([]SampleEvalMetrics, 0, n)
	for i, r := range results {
		sumA += r.Actionability
		sumSp += r.Specificity
		sumSev += r.Severity
		sumL += r.Logic
		sumP += r.Performance
		sumS += r.Security
		sumY += r.Style
		samples = append(samples, SampleEvalMetrics{
			Index:            i + 1,
			Total:            r.Total,
			Actionability:    r.Actionability,
			Specificity:      r.Specificity,
			Severity:         r.Severity,
			WeaknessCategory: r.WeaknessCategory,
			Logic:            r.Logic,
			Performance:      r.Performance,
			Security:         r.Security,
			Style:            r.Style,
		})
	}

	fn := func(sum int) int {
		return int(math.Round(float64(sum) / float64(n)))
	}

	m := IterationMetrics{
		Actionability:        fn(sumA),
		Specificity:          fn(sumSp),
		Severity:             fn(sumSev),
		Structure:            fn(sumY), // mean style → “Structure” column in UI
		Logic:                fn(sumL),
		Performance:          fn(sumP),
		Security:             fn(sumS),
		WeakestIssueCategory: weakestIssueCategory,
		Samples:              samples,
	}
	return json.Marshal(m)
}
