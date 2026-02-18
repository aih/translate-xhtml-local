package translator

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

// TranslationService defines the interface for translating XHTML content.
type TranslationService interface {
	Translate(ctx context.Context, r *strings.Reader, sourceLang, targetLang string) (string, Metadata, error)
}

// Metadata contains information about the translation process.
type Metadata struct {
	Duration  time.Duration `json:"duration" swaggertype:"primitive,integer"`
	Model     string        `json:"model"`
	Timestamp time.Time     `json:"timestamp"`
}

// LLMClient defines the interface for the language model client.
type LLMClient interface {
	TranslateText(ctx context.Context, text, sourceLang, targetLang string) (string, error)
	GetModelName() string
}

// Service implements TranslationService.
type Service struct {
	llm LLMClient
}

// NewService creates a new TranslationService.
func NewService(llm LLMClient) *Service {
	return &Service{llm: llm}
}

// Translate parses the XHTML, translates text nodes, and returns the result.
func (s *Service) Translate(ctx context.Context, r *strings.Reader, sourceLang, targetLang string) (string, Metadata, error) {
	start := time.Now()

	doc, err := html.Parse(r)
	if err != nil {
		return "", Metadata{}, fmt.Errorf("failed to parse XHTML: %w", err)
	}

	// Channel to collect text nodes that need translation
	type textNode struct {
		n    *html.Node
		text string
	}
	nodesToTranslate := make([]*textNode, 0)

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			trimmed := strings.TrimSpace(n.Data)
			if trimmed != "" && n.Parent != nil && n.Parent.Data != "script" && n.Parent.Data != "style" {
				nodesToTranslate = append(nodesToTranslate, &textNode{n: n, text: n.Data})
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	// Process translations concurrently
	// Limit concurrency to avoid overwhelming the local LLM
	sem := make(chan struct{}, 5) // Adjust concurrency limit as needed
	var wg sync.WaitGroup
	errChan := make(chan error, len(nodesToTranslate))

	for _, tn := range nodesToTranslate {
		wg.Add(1)
		go func(tn *textNode) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			translated, err := s.llm.TranslateText(ctx, tn.text, sourceLang, targetLang)
			if err != nil {
				errChan <- err
				return
			}
			tn.n.Data = translated
		}(tn)
	}

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		return "", Metadata{}, <-errChan
	}

	var buf strings.Builder
	if err := html.Render(&buf, doc); err != nil {
		return "", Metadata{}, fmt.Errorf("failed to render translated XHTML: %w", err)
	}

	return buf.String(), Metadata{
		Duration:  time.Since(start),
		Model:     s.llm.GetModelName(),
		Timestamp: time.Now(),
	}, nil
}
