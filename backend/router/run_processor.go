package router

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/Shiva936/code-review-agent/backend/config"
	"github.com/Shiva936/code-review-agent/backend/core/evaluator"
	"github.com/Shiva936/code-review-agent/backend/core/generator"
	"github.com/Shiva936/code-review-agent/backend/core/refiner"
	"github.com/Shiva936/code-review-agent/backend/models"
	"github.com/Shiva936/code-review-agent/backend/storage"
)

// allow tests to stub the processor
var processRunGroupAsync = processRunGroup

// samplesForRun uses the user-submitted snippet for review/eval (stored on run_groups for the UI).
// Falls back to hardcoded demos only if code is empty.
func samplesForRun(code string) []string {
	c := strings.TrimSpace(code)
	if c == "" {
		return hardcodedCodeSamples
	}
	return []string{c}
}

func processRunGroup(cfg *config.Config, runGroupID int, code string, extraPrompt string) {
	const iterations = 5
	samples := samplesForRun(code)
	if len(samples) == 0 {
		log.Printf("run_group %d: no hardcoded samples", runGroupID)
		_ = storage.UpdateRunGroupStatus(runGroupID, "failed")
		return
	}

	if err := storage.UpdateRunGroupStatus(runGroupID, "running"); err != nil {
		log.Printf("run_group %d: failed to set status=running: %v", runGroupID, err)
		return
	}

	basePrompt := `You are a strict code reviewer... Provide categorized feedback (logic, performance, security, style) with clear severities (critical, minor, suggestion) and actionable fixes. Avoid vague advice.`
	prompt := basePrompt
	if extraPrompt != "" {
		prompt = prompt + "\n\nAdditional instructions:\n" + extraPrompt
	}
	basePromptWithContext := prompt
	existingRules := []string{}
	prevReviewBySample := make([]string, len(samples))
	var prevAggRubric int

	for iter := 1; iter <= iterations; iter++ {
		_ = storage.UpdateRunGroupRunStatus(runGroupID, iter, "running")
		prompt = refiner.BuildPrompt(basePromptWithContext, existingRules)

		rulesJSON, _ := json.Marshal(existingRules)
		if err := storage.SaveRunGroupPromptVersion(&storage.RunGroupPromptVersion{
			RunGroupID: runGroupID,
			Iteration:  iter,
			PromptText: prompt,
			RulesJSON:  string(rulesJSON),
			Source:     "active",
			Reason:     "prompt used for generation/evaluation in this iteration",
		}); err != nil {
			log.Printf("run_group %d iter %d: warning: failed to save prompt version: %v", runGroupID, iter, err)
		}

		results := make([]*models.EvalResult, 0, len(samples))
		reviewsThisIter := make([]string, len(samples))
		for si, sampleCode := range samples {
			genCtx := &models.IterationContext{
				Iteration:           iter,
				TotalIterations:     iterations,
				PreviousReview:      prevReviewBySample[si],
				PreviousRubricTotal: prevAggRubric,
				IterSampleSeed:      int64(iter*100000 + si*1000 + runGroupID),
			}
			review, genErr := generator.Generate(cfg, prompt, sampleCode, genCtx)
			if genErr != nil {
				log.Printf("run_group %d iter %d sample %d: generation error: %v", runGroupID, iter, si+1, genErr)
				review = "No review generated due to an error."
			}
			reviewsThisIter[si] = review

			evalResult, evalErr := evaluator.Evaluate(cfg, review, genCtx)
			if evalResult == nil {
				_ = storage.UpdateRunGroupRunStatus(runGroupID, iter, "failed")
				_ = storage.UpdateRunGroupStatus(runGroupID, "failed")
				log.Printf("run_group %d iter %d sample %d: evaluation returned nil (err=%v)", runGroupID, iter, si+1, evalErr)
				return
			}
			if evalErr != nil {
				log.Printf("run_group %d iter %d sample %d: evaluation warning: %v", runGroupID, iter, si+1, evalErr)
			}

			log.Printf(
				"run_group %d iter %d sample %d/%d: rubric total=%d (A=%d S=%d Sev=%d) weakest_rubric=%s | categories logic=%d perf=%d sec=%d style=%d",
				runGroupID, iter, si+1, len(samples),
				evalResult.Total, evalResult.Actionability, evalResult.Specificity, evalResult.Severity, evalResult.WeaknessCategory,
				evalResult.Logic, evalResult.Performance, evalResult.Security, evalResult.Style,
			)

			results = append(results, evalResult)
		}

		avgTotal, weakestIssue := aggregateAcrossSamples(results)
		log.Printf(
			"run_group %d iter %d aggregate: avg_total=%d weakest_issue_category=%s (refine target)",
			runGroupID, iter, avgTotal, weakestIssue,
		)

		score := avgTotal
		weakness := weakestIssue
		prevAggRubric = score
		copy(prevReviewBySample, reviewsThisIter)

		detailJSON, mErr := models.BuildIterationMetrics(results, weakestIssue)
		if mErr != nil {
			log.Printf("run_group %d iter %d: failed to build iteration metrics: %v", runGroupID, iter, mErr)
			detailJSON = []byte("{}")
		}

		if err := storage.UpdateRunGroupRun(runGroupID, iter, score, weakness, "completed", string(detailJSON)); err != nil {
			_ = storage.UpdateRunGroupRunStatus(runGroupID, iter, "failed")
			_ = storage.UpdateRunGroupStatus(runGroupID, "failed")
			log.Printf("run_group %d iter %d: failed to save iteration: %v", runGroupID, iter, err)
			return
		}
		if err := storage.SaveRun(&storage.Run{Iteration: iter, Score: score, Weakness: weakness}); err != nil {
			log.Printf("run_group %d iter %d: warning: failed to save legacy run: %v", runGroupID, iter, err)
		}

		if err := storage.TouchRunGroup(runGroupID); err != nil {
			log.Printf("run_group %d iter %d: warning: failed to touch updated_at: %v", runGroupID, iter, err)
		}

		refineDecision := refiner.RefineWithGuardrails(cfg, basePromptWithContext, weakness, existingRules, iter, string(detailJSON))
		prompt = refineDecision.Prompt
		existingRules = refineDecision.Rules

		if err := storage.SaveRunGroupPromptDelta(&storage.RunGroupPromptDelta{
			RunGroupID:       runGroupID,
			Iteration:        iter,
			WeakestIssue:     weakness,
			InputJSON:        refineDecision.InputJSON,
			RawOutput:        refineDecision.RawOutput,
			DeltaJSON:        refineDecision.DeltaJSON,
			ValidationStatus: refineDecision.ValidationStatus,
			Applied:          refineDecision.Applied,
			Source:           refineDecision.Source,
			Reason:           refineDecision.Reason,
		}); err != nil {
			log.Printf("run_group %d iter %d: warning: failed to save prompt delta: %v", runGroupID, iter, err)
		}
	}

	if err := storage.UpdateRunGroupStatus(runGroupID, "completed"); err != nil {
		log.Printf("run_group %d: failed to set status=completed: %v", runGroupID, err)
		return
	}

	log.Printf("run_group %d: completed %d iterations over %d code sample(s)", runGroupID, iterations, len(samples))
}

func progressPercent(iteration int, totalIterations int) int {
	if totalIterations <= 0 || iteration <= 0 {
		return 0
	}
	if iteration >= totalIterations {
		return 100
	}
	return int((float64(iteration) / float64(totalIterations)) * 100.0)
}

func validateStatus(status string) error {
	switch status {
	case "pending", "running", "completed", "failed":
		return nil
	default:
		return fmt.Errorf("invalid status: %s", status)
	}
}
