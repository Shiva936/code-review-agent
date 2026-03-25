package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Shiva936/code-review-agent/backend/config"
)

// CallOpts configures sampling. Non-zero Temperature reduces mode-collapse (same scores/text every call).
// Seed (when non-zero) is sent to OpenRouter to reduce duplicate completions across identical-looking prompts.
type CallOpts struct {
	Temperature float64
	TopP        float64 // 0 = omit; e.g. 0.95
	Seed        int64   // 0 = omit
}

// CallLLM sends a prompt to the specified model via OpenRouter API.
func CallLLM(cfg *config.Config, requestType string, prompt string, model string) (string, error) {
	return callLLM(cfg, requestType, prompt, model, nil)
}

// CallLLMWithOpts is like CallLLM but sets temperature (and future options) on the completion request.
func CallLLMWithOpts(cfg *config.Config, requestType string, prompt string, model string, opts *CallOpts) (string, error) {
	return callLLM(cfg, requestType, prompt, model, opts)
}

func callLLM(cfg *config.Config, requestType string, prompt string, model string, opts *CallOpts) (string, error) {
	if cfg.OpenRouterAPIKey == "" {
		return "", fmt.Errorf("OPENROUTER_API_KEY environment variable not set")
	}

	requestBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}
	if opts != nil {
		if opts.Temperature > 0 {
			requestBody["temperature"] = opts.Temperature
		}
		if opts.TopP > 0 && opts.TopP <= 1 {
			requestBody["top_p"] = opts.TopP
		}
		if opts.Seed != 0 {
			requestBody["seed"] = opts.Seed
		}
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	_ = requestType // reserved for tracing / future logging
	req.Header.Set("Authorization", "Bearer "+cfg.OpenRouterAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	log.Printf("OpenRouter API response: %s", string(body))

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Error.Message != "" {
		return "", fmt.Errorf("API error: %s", response.Error.Message)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return response.Choices[0].Message.Content, nil
}
