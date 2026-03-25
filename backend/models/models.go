package models

// CodeReview represents a structured code review output.
type CodeReview struct {
	Comments []Comment
}

// Comment represents a single review comment.
type Comment struct {
	Category string // logic, performance, security, style
	Severity string // critical, minor, suggestion
	Message  string
	Line     int
}

// LoopState represents the persistent state of the improvement loop.
type LoopState struct {
	Iteration       int
	CurrentPrompt   string
	WeaknessHistory map[string]int
	AverageScore    float64
	BestScore       int
}

// Run represents a single iteration result.
type Run struct {
	Iteration int
	Score     int
	Weakness  string
}

// EvalResult represents the evaluation score breakdown.
// Rubric (15 pts): actionability, specificity, severity (each 1–5).
// Issue-category scores (each 1–5): how well the review covers quality comments in that bucket.
type EvalResult struct {
	Actionability    int    `json:"actionability"`
	Specificity      int    `json:"specificity"`
	Severity         int    `json:"severity"`
	Total            int    `json:"total"`
	WeaknessCategory string `json:"weakness_category"`
	Logic            int    `json:"logic"`
	Performance      int    `json:"performance"`
	Security         int    `json:"security"`
	Style            int    `json:"style"`
}
