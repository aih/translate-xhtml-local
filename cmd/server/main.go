package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/arihershowitz/translate-xhtml-local/internal/api"
	"github.com/arihershowitz/translate-xhtml-local/internal/llm"
	"github.com/arihershowitz/translate-xhtml-local/internal/translator"

	_ "github.com/arihershowitz/translate-xhtml-local/docs" // Import generated docs
)

func main() {
	var (
		port        = flag.String("port", "8080", "Server port")
		llmEndpoint = flag.String("llm-url", "http://localhost:11434/api/generate", "Local LLM endpoint")
		llmModel    = flag.String("model", "google/translategemma-4b-it", "Model name to use")
	)
	flag.Parse()

	// Initialize LLM client
	llmClient := llm.NewClient(*llmEndpoint, *llmModel)

	// Initialize Translator Service
	translationService := translator.NewService(llmClient)

	// Initialize API Handler
	handler := api.NewHandler(translationService)

	// Setup Routes
	mux := http.NewServeMux()
	mux.HandleFunc("/translate", handler.Translate)

	// Serve Swagger UI (placeholder, assumes valid swagger.json generated)
	// For simplicity, we just log where to find docs in this version
	// In a real app, use http-swagger or similar to serve the UI

	log.Printf("Starting server on port %s", *port)
	log.Printf("Using LLM at %s with model %s", *llmEndpoint, *llmModel)

	if err := http.ListenAndServe(":"+*port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
