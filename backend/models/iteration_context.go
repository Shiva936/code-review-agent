package models

// IterationContext carries state between loop iterations for the same code sample.
// Used by the generator (avoid repeating the same review) and evaluator (score relative to iteration).
type IterationContext struct {
	Iteration           int
	TotalIterations     int
	PreviousReview      string // truncated; empty on first iteration
	PreviousRubricTotal int    // sum actionability+specificity+severity from previous iter; 0 if none
	// IterSampleSeed is a stable nonce per (iteration, sample index) for LLM seeding / cache busting.
	IterSampleSeed int64
}
