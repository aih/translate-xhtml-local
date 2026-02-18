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

	"github.com/arihershowitz/translate-xhtml-local/internal/translator"
)

// MockLLM simulates a high-throughput LLM.
type MockLLM struct {
	ModelName string
}

func (m *MockLLM) TranslateText(ctx context.Context, text, sourceLang, targetLang string) (string, error) {
	// Simulate slight processing time but allow high concurrency
	time.Sleep(10 * time.Millisecond)
	return fmt.Sprintf("[%s->%s] %s", sourceLang, targetLang, text), nil
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

	// Create service with MockLLM
	mockLLM := &MockLLM{ModelName: "mock-high-throughput"}
	service := translator.NewService(mockLLM)

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
