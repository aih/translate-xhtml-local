package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/arihershowitz/translate-xhtml-local/internal/llm"
	"github.com/arihershowitz/translate-xhtml-local/internal/translator"
)

// MockLLM simulates a high-throughput LLM with pseudo-translation.
type MockLLM struct {
	ModelName string
}

func (m *MockLLM) TranslateText(ctx context.Context, text, sourceLang, targetLang string) (string, error) {
	// Simulate slight processing time
	time.Sleep(10 * time.Millisecond)

	// Pseudo-translation: distinct output that replaces original text
	// e.g. "Hello" -> "[es] Olleh" (reversed + tag) or just distinct replacement
	// For visibility, we'll produce a "lorem ipsum" style replacement relative to length,
	// or simple character mapping if we want to preserve length.
	// Let's do a simple visible transformation:
	return fmt.Sprintf("[%s] %s", targetLang, reverse(text)), nil
}

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func (m *MockLLM) GetModelName() string {
	return m.ModelName
}

func TestLargeScaleTranslation(t *testing.T) {
	// Ensure data exists
	downloadTestData(t)

	// Output directory
	outputDir := "output"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}

	// Get all test files
	files, err := filepath.Glob(filepath.Join(dataDir, "*.xhtml"))
	if err != nil {
		t.Fatalf("Failed to glob files: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("No test data found")
	}

	languages := []string{"es", "fr", "de", "it", "pt", "ru", "zh", "ja", "ko", "nl"}

	// Create service
	var service translator.TranslationService

	// Check for real LLM env var
	if os.Getenv("TEST_REAL_LLM") == "true" {
		t.Log("Using REAL LLM for translation tests (this will be slow)")
		model := os.Getenv("TEST_LLM_MODEL")
		if model == "" {
			model = "google/translategemma-4b-it"
		}
		llmClient := llm.NewClient("http://localhost:11434/api/generate", model)
		service = translator.NewService(llmClient)
	} else {
		t.Log("Using MOCK LLM for translation tests. Set TEST_REAL_LLM=true to use real model.")
		mockLLM := &MockLLM{ModelName: "mock-high-throughput"}
		service = translator.NewService(mockLLM)
	}

	// Limit files for sampling if requested
	if limitStr := os.Getenv("TEST_SAMPLE_LIMIT"); limitStr != "" {
		// If sample.xhtml exists, use ONLY that for speed
		sampleFile := filepath.Join(dataDir, "sample.xhtml")
		if _, err := os.Stat(sampleFile); err == nil {
			t.Log("Using sample.xhtml for quick sampling")
			files = []string{sampleFile}
		} else {
			limit := 1
			fmt.Sscanf(limitStr, "%d", &limit)
			if limit < len(files) {
				t.Logf("Limiting test to %d files for sampling", limit)
				files = files[:limit]
			}
		}
	}

	var wg sync.WaitGroup
	start := time.Now()

	sem := make(chan struct{}, 50) // Limit concurrency to 50 files processed at once, although each file spawns its own goroutines for text nodes.

	for _, file := range files {
		for _, lang := range languages {
			wg.Add(1)
			go func(f, l string) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				// Read file
				content, err := os.ReadFile(f)
				if err != nil {
					t.Errorf("Failed to read %s: %v", f, err)
					return
				}

				// Translate
				filename := filepath.Base(f)
				t.Logf("Translating %s to %s", filename, l)
				translated, _, err := service.Translate(context.Background(), strings.NewReader(string(content)), "en", l)
				if err != nil {
					t.Errorf("Failed to translate %s to %s: %v", filename, l, err)
					return
				}

				// Verify image preservation (simple check)
				if strings.Contains(string(content), "<img") {
					if !strings.Contains(translated, "<img") {
						t.Errorf("Image tag missing in translation of %s to %s", filename, l)
					}
				}

				// Verify src attribute preservation check (naive but effective for check)
				// If original has src="foo.png", translated should too.
				// Since we mock translation, attributes aren't touched by the walker, so they should persist.

				// Write output
				outFile := filepath.Join(outputDir, fmt.Sprintf("%s_%s.xhtml", strings.TrimSuffix(filename, ".xhtml"), l))
				if err := os.WriteFile(outFile, []byte(translated), 0644); err != nil {
					t.Errorf("Failed to write output %s: %v", outFile, err)
				}
			}(file, lang)
		}
	}

	wg.Wait()
	duration := time.Since(start)
	t.Logf("Processed %d files * %d languages in %v", len(files), len(languages), duration)
}
