package router

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"sync"

	"github.com/Shiva936/code-review-agent/backend/config"
	"github.com/Shiva936/code-review-agent/backend/core/evaluator"
	"github.com/Shiva936/code-review-agent/backend/core/generator"
	"github.com/Shiva936/code-review-agent/backend/core/refiner"
	"github.com/Shiva936/code-review-agent/backend/storage"
)

var (
	runMu     sync.Mutex
	isRunning bool
)

type LoopSummary struct {
	Iterations  int      `json:"iterations"`
	SampleCount int      `json:"sample_count"`
	AvgScores   []int    `json:"avg_scores"`
	Weaknesses  []string `json:"weaknesses"`
	GroupID     int      `json:"group_id"`
}

func runHandler(cfg *config.Config, w http.ResponseWriter, r *http.Request) {
	runMu.Lock()
	if isRunning {
		runMu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooEarly)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "loop already running"})
		return
	}
	isRunning = true
	runMu.Unlock()

	defer func() {
		runMu.Lock()
		isRunning = false
		runMu.Unlock()
	}()

	var req struct {
		Code   string `json:"code"`
		Prompt string `json:"prompt"`
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON body"})
		return
	}
	if req.Code == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "`code` is required"})
		return
	}

	summary, err := runLoop(cfg, req.Code, req.Prompt)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("loop failed: %v", err)})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(summary)
}

func runLoop(cfg *config.Config, code string, extraPrompt string) (*LoopSummary, error) {
	const iterations = 5

	// Base prompt used as the initial "additional instructions" for the generator.
	basePrompt := `You are a strict code reviewer... Provide categorized feedback (logic, performance, security, style) with clear severities (critical, minor, suggestion) and actionable fixes. Avoid vague advice.`

	summary := &LoopSummary{
		Iterations:  iterations,
		SampleCount: 1,
		AvgScores:   make([]int, 0, iterations),
		Weaknesses:  make([]string, 0, iterations),
	}

	// Prompt refinement state (used to avoid duplicate rules).
	prompt := basePrompt
	if extraPrompt != "" {
		prompt = prompt + "\n\nAdditional instructions:\n" + extraPrompt
	}
	existingRules := []string{}

	// Persist the initial prompt once (version 0).
	promptVersionStart := 0
	if versions, err := storage.GetPromptVersions(); err != nil {
		log.Printf("warning: failed to load prompt versions: %v", err)
	} else if len(versions) > 0 {
		promptVersionStart = versions[len(versions)-1].Version + 1
	}

	if err := storage.SavePromptVersion(promptVersionStart, prompt, "initial prompt"); err != nil {
		log.Printf("warning: failed to save initial prompt: %v", err)
	}

	groupID, err := storage.CreateRunGroup(code, basePrompt, iterations)
	if err != nil {
		return nil, err
	}
	summary.GroupID = groupID

	for iter := 1; iter <= iterations; iter++ {
		log.Printf("=== Iteration %d/%d ===", iter, iterations)

		log.Printf("Generating review...")
		review, genErr := generator.Generate(cfg, prompt, code)
		if genErr != nil {
			log.Printf("Generation error: %v", genErr)
			review = "No review generated due to an error."
		}

		log.Printf("Evaluating review...")
		evalResult, evalErr := evaluator.Evaluate(cfg, review)
		if evalResult == nil {
			return nil, fmt.Errorf("evaluation returned nil result")
		}
		if evalErr != nil {
			log.Printf("Evaluation warning: %v", evalErr)
		}

		score := evalResult.Total
		iterationWeakness := evalResult.WeaknessCategory
		scoreInt := int(math.Round(float64(score)))

		log.Printf(
			"Iteration %d summary: score=%d weakness=%s",
			iter,
			score,
			iterationWeakness,
		)

		// Persist this iteration's aggregated result.
		if err := storage.SaveRun(&storage.Run{
			Iteration: iter,
			Score:     scoreInt,
			Weakness:  iterationWeakness,
		}); err != nil {
			log.Printf("warning: failed to save run: %v", err)
		}

		if err := storage.SaveRunGroupRun(groupID, iter, scoreInt, iterationWeakness); err != nil {
			log.Printf("warning: failed to save run group run: %v", err)
		}

		// Persist the refined prompt state for this iteration.
		if err := storage.SavePromptVersion(promptVersionStart+iter, prompt, fmt.Sprintf("refined after weakest category: %s", iterationWeakness)); err != nil {
			log.Printf("warning: failed to save prompt version: %v", err)
		}

		// Refine prompt for the next iteration.
		prompt, existingRules = refiner.Refine(cfg, prompt, iterationWeakness, existingRules)

		summary.AvgScores = append(summary.AvgScores, scoreInt)
		summary.Weaknesses = append(summary.Weaknesses, iterationWeakness)
	}

	return summary, nil
}
