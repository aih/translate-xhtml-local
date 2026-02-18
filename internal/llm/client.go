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
	// Refined prompt to guide the model better
	prompt := fmt.Sprintf(`Translate the following text from %s to %s. 
Rules:
1. Output ONLY the translated text.
2. Do NOT add notes, explanations, or enclosing quotes.
3. Preserve the original meaning and tone.
4. If the text is a number or proper noun that shouldn't change, keep it as is.
5. If the translation is unclear, provide the most direct literal translation.

Text to translate:
"%s"`, sourceLang, targetLang, text)

	reqBody := map[string]interface{}{
		"model":  c.model,
		"prompt": prompt,
		"stream": false,
		"options": map[string]interface{}{
			"temperature": 0.1, // Lower temperature for more deterministic/focused output
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
