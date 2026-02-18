package translator

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"
)

// MockLLM is a mock implementation of LLMClient.
type MockLLM struct {
	TranslateFunc func(ctx context.Context, text, sourceLang, targetLang string) (string, error)
	ModelName     string
}

func (m *MockLLM) TranslateText(ctx context.Context, text, sourceLang, targetLang string) (string, error) {
	if m.TranslateFunc != nil {
		return m.TranslateFunc(ctx, text, sourceLang, targetLang)
	}
	return "TRANSLATED_" + text, nil
}

func (m *MockLLM) GetModelName() string {
	return m.ModelName
}

func TestTranslate_Structure(t *testing.T) {
	mockLLM := &MockLLM{
		ModelName: "test-model",
		TranslateFunc: func(ctx context.Context, text, sourceLang, targetLang string) (string, error) {
			return "TR:" + text, nil
		},
	}
	service := NewService(mockLLM)

	input := `<div><h1>Hello</h1><p>World</p></div>`
	// net/html adds html, head, and body tags if missing
	expected := `<html><head></head><body><div><h1>TR:Hello</h1><p>TR:World</p></div></body></html>`

	translated, _, err := service.Translate(context.Background(), strings.NewReader(input), "en", "es")
	if err != nil {
		t.Fatalf("Translate failed: %v", err)
	}

	if translated != expected {
		t.Errorf("Expected %q, got %q", expected, translated)
	}
}

func TestTranslate_Concurrency(t *testing.T) {
	var mu sync.Mutex
	activeRequests := 0
	maxConcurrent := 0

	mockLLM := &MockLLM{
		ModelName: "test-model",
		TranslateFunc: func(ctx context.Context, text, sourceLang, targetLang string) (string, error) {
			mu.Lock()
			activeRequests++
			if activeRequests > maxConcurrent {
				maxConcurrent = activeRequests
			}
			mu.Unlock()

			time.Sleep(50 * time.Millisecond) // Simulate processing time

			mu.Lock()
			activeRequests--
			mu.Unlock()

			return "TR:" + text, nil
		},
	}
	service := NewService(mockLLM)

	// optimized input with multiple text nodes to trigger concurrency
	input := `<div>
		<p>Line 1</p>
		<p>Line 2</p>
		<p>Line 3</p>
		<p>Line 4</p>
		<p>Line 5</p>
		<p>Line 6</p>
	</div>`

	start := time.Now()
	_, _, err := service.Translate(context.Background(), strings.NewReader(input), "en", "es")
	if err != nil {
		t.Fatalf("Translate failed: %v", err)
	}
	duration := time.Since(start)

	// If 6 items take 50ms each serially, it would take ~300ms.
	// With concurrency, it should be significantly faster (closer to 50ms + overhead).
	if duration > 200*time.Millisecond {
		t.Errorf("Translation took too long (%v), indicating lack of concurrency", duration)
	}

	if maxConcurrent <= 1 {
		t.Errorf("Max concurrent requests was %d, expected > 1", maxConcurrent)
	}
}
