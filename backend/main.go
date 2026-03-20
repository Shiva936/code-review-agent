package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"sync"

	"github.com/Shiva936/code-review-agent/backend/core/evaluator"
	"github.com/Shiva936/code-review-agent/backend/core/generator"
	"github.com/Shiva936/code-review-agent/backend/core/refiner"
	"github.com/Shiva936/code-review-agent/backend/storage"
)

var (
	runMu     sync.Mutex
	isRunning bool
)

func setCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := storage.InitDB(""); err != nil {
		log.Fatalf("failed to init db: %v", err)
	}

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/run", runHandler)
	http.HandleFunc("/runs", runsHandler)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Server running on port %s\n", port)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

type LoopSummary struct {
	Iterations  int      `json:"iterations"`
	SampleCount int      `json:"sample_count"`
	AvgScores   []int    `json:"avg_scores"`
	Weaknesses  []string `json:"weaknesses"`
}

func runHandler(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

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

	summary, err := runLoop()
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

type runResponse struct {
	Iteration int    `json:"iteration"`
	Score     int    `json:"score"`
	Weakness  string `json:"weakness"`
}

func runsHandler(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	runs, err := storage.GetRuns()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("failed to load runs: %v", err)})
		return
	}

	resp := make([]runResponse, 0, len(runs))
	for _, run := range runs {
		resp = append(resp, runResponse{
			Iteration: run.Iteration,
			Score:     run.Score,
			Weakness:  run.Weakness,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string][]runResponse{"runs": resp})
}

func runLoop() (*LoopSummary, error) {
	const iterations = 5

	// Base prompt used as the initial "additional instructions" for the generator.
	basePrompt := `You are a strict code reviewer... Provide categorized feedback (logic, performance, security, style) with clear severities (critical, minor, suggestion) and actionable fixes. Avoid vague advice.`

	// Three hardcoded code samples for deterministic loop behavior.
	codeSamples := []string{
		`// Sample 1: SQL injection risk (security)
// NOTE: This is intentionally unsafe for the demo.
func FindUserByEmail(db *sql.DB, email string) (string, error) {
	// Vulnerability: string concatenation in query construction
	query := "SELECT id, name FROM users WHERE email = '" + email + "'"
	row := db.QueryRow(query)
	var name string
	if err := row.Scan(&name); err != nil {
		return "", err
	}
	return name, nil
}`,
		`// Sample 2: Panic and weak error handling (logic + severity)
func ParseConfig(path string) map[string]string {
	data, _ := os.ReadFile(path) // Ignoring error is a bug
	lines := strings.Split(string(data), "\n")

	cfg := make(map[string]string)
	for _, line := range lines {
		parts := strings.Split(line, "=")
		// Vulnerability: parts[1] panics when "=" is missing
		cfg[parts[0]] = parts[1]
	}
	return cfg
}`,
		`// Sample 3: Inefficient performance + insecure randomness (performance + security)
func GenerateToken(n int) string {
	// Insecure: math/rand is not suitable for tokens.
	// Inefficient: repeated string concatenation in a loop.
	token := ""
	for i := 0; i < n; i++ {
		token += string(rune('a' + rand.Intn(26)))
	}
	return token
}`,
	}

	summary := &LoopSummary{
		Iterations:  iterations,
		SampleCount: len(codeSamples),
		AvgScores:   make([]int, 0, iterations),
		Weaknesses:  make([]string, 0, iterations),
	}

	// Prompt refinement state (used to avoid duplicate rules).
	prompt := basePrompt
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

	for iter := 1; iter <= iterations; iter++ {
		log.Printf("=== Iteration %d/%d ===", iter, iterations)

		totals := make([]int, 0, len(codeSamples))
		weaknesses := make([]string, 0, len(codeSamples))

		for i, code := range codeSamples {
			log.Printf("Sample %d/%d: generating review...", i+1, len(codeSamples))

			review, genErr := generator.Generate(prompt, code)
			if genErr != nil {
				// Keep the loop autonomous: proceed with an error placeholder.
				log.Printf("Sample %d/%d: generation error: %v", i+1, len(codeSamples), genErr)
				review = "No review generated due to an error."
			}

			log.Printf("Sample %d/%d: evaluating review...", i+1, len(codeSamples))
			evalResult, evalErr := evaluator.Evaluate(review)
			if evalResult == nil {
				return nil, fmt.Errorf("evaluation returned nil result")
			}
			if evalErr != nil {
				log.Printf("Sample %d/%d: evaluation warning: %v", i+1, len(codeSamples), evalErr)
			}

			log.Printf(
				"Sample %d/%d: total=%d (actionability=%d specificity=%d severity=%d) weakness=%s",
				i+1,
				len(codeSamples),
				evalResult.Total,
				evalResult.Actionability,
				evalResult.Specificity,
				evalResult.Severity,
				evalResult.WeaknessCategory,
			)

			totals = append(totals, evalResult.Total)
			weaknesses = append(weaknesses, evalResult.WeaknessCategory)

			// Refine prompt for the next sample within this iteration.
			log.Printf("Sample %d/%d: refining prompt for weakness=%s...", i+1, len(codeSamples), evalResult.WeaknessCategory)
			prompt, existingRules = refiner.Refine(prompt, evalResult.WeaknessCategory, existingRules)
		}

		var sum int
		for _, t := range totals {
			sum += t
		}
		avg := float64(sum) / float64(len(totals))
		avgInt := int(math.Round(avg))

		// Choose the worst weakness category for the iteration based on the lowest-scoring sample.
		worstIdx := 0
		worstScore := totals[0]
		for idx, s := range totals {
			if s < worstScore {
				worstIdx = idx
				worstScore = s
			}
		}
		iterationWeakness := weaknesses[worstIdx]

		log.Printf(
			"Iteration %d summary: scores=%v avg=%.2f saved_score=%d weakness=%s",
			iter,
			totals,
			avg,
			avgInt,
			iterationWeakness,
		)

		// Persist this iteration's aggregated result.
		if err := storage.SaveRun(&storage.Run{
			Iteration: iter,
			Score:     avgInt,
			Weakness:  iterationWeakness,
		}); err != nil {
			log.Printf("warning: failed to save run: %v", err)
		}

		// Persist the refined prompt state for this iteration.
		if err := storage.SavePromptVersion(promptVersionStart+iter, prompt, fmt.Sprintf("refined after weakest category: %s", iterationWeakness)); err != nil {
			log.Printf("warning: failed to save prompt version: %v", err)
		}

		summary.AvgScores = append(summary.AvgScores, avgInt)
		summary.Weaknesses = append(summary.Weaknesses, iterationWeakness)
	}

	return summary, nil
}
