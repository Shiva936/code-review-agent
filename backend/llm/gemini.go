package llm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Shiva936/code-review-agent/backend/config"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

func callGemini(cfg *config.Config, prompt string, model string, opts *CallOpts) (string, error) {
	if strings.TrimSpace(cfg.GeminiAPIKey) == "" {
		return "", fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}
	if strings.TrimSpace(model) == "" {
		model = "gemini-1.5-flash"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.GeminiAPIKey))
	if err != nil {
		return "", fmt.Errorf("failed to create gemini client: %w", err)
	}
	defer client.Close()

	gModel := client.GenerativeModel(model)
	if opts != nil {
		if opts.Temperature > 0 {
			t := float32(opts.Temperature)
			gModel.Temperature = &t
		}
		if opts.TopP > 0 && opts.TopP <= 1 {
			p := float32(opts.TopP)
			gModel.TopP = &p
		}
	}

	resp, err := gModel.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("gemini generate failed: %w", err)
	}
	if resp == nil || len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return "", fmt.Errorf("gemini returned no candidates")
	}

	var sb strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		switch v := part.(type) {
		case genai.Text:
			sb.WriteString(string(v))
		default:
			// Best-effort stringify for non-text parts.
			sb.WriteString(fmt.Sprint(v))
		}
	}
	out := strings.TrimSpace(sb.String())
	if out == "" {
		return "", fmt.Errorf("gemini returned empty content")
	}
	return out, nil
}
