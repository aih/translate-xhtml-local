package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client implements the translator.LLMClient interface.
type Client struct {
	endpoint string
	model    string
	client   *http.Client
}

// NewClient creates a new LLM client.
func NewClient(endpoint, model string) *Client {
	return &Client{
		endpoint: endpoint,
		model:    model,
		client: &http.Client{
			Timeout: 60 * time.Second, // Adjust timeout as needed
		},
	}
}

// GetModelName returns the model name.
func (c *Client) GetModelName() string {
	return c.model
}

// TranslateText sends a translation request to the local LLM.
// This example assumes an Ollama-compatible API or similar simple JSON interface.
// Adjust the request/response structure based on the actual local server.
func (c *Client) TranslateText(ctx context.Context, text, sourceLang, targetLang string) (string, error) {
	// Refined prompt: use a "completion" style rather than "chat" to avoid conversational filler.
	// We wrap it in a strict pattern.
	prompt := fmt.Sprintf(`Translate the english text "%s" to %s. return only the translated string.`, text, targetLang)

	reqBody := map[string]interface{}{
		"model":  c.model,
		"prompt": prompt,
		"stream": false,
		"options": map[string]interface{}{
			"temperature": 0.0, // Zero temperature for maximum determinism
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("LLM server returned status %d: %s", resp.StatusCode, string(body))
	}

	var respBody struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return respBody.Response, nil
}
