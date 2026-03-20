package llm

// Response represents the LLM API response.
type Response struct {
	Content string
	Error   string
}

// CallOpenRouter sends a request to the OpenRouter API.
func CallOpenRouter(prompt string) (*Response, error) {
	// TODO: Implement OpenRouter API integration
	return &Response{}, nil
}
