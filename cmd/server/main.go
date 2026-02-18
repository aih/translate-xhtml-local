package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/arihershowitz/translate-xhtml-local/internal/api"
	"github.com/arihershowitz/translate-xhtml-local/internal/llm"
	"github.com/arihershowitz/translate-xhtml-local/internal/translator"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/arihershowitz/translate-xhtml-local/docs" // Import generated docs
)

func main() {
	defaultPort := os.Getenv("PORT")
	if defaultPort == "" {
		defaultPort = "8090"
	}

	var (
		port        = flag.String("port", defaultPort, "Server port")
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

	// Serve Swagger UI
	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	log.Printf("Starting server on port %s", *port)
	log.Printf("Using LLM at %s with model %s", *llmEndpoint, *llmModel)
	log.Printf("Swagger UI available at http://localhost:%s/swagger/index.html", *port)

	if err := http.ListenAndServe(":"+*port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
