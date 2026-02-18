# XHTML Translation Service

This is a Golang application that translates XHTML pages using a local LLM (e.g., Ollama with `google/translategemma-4b-it` or compatible models).

## Features

- **XHTML Parsing**: Preserves the structure of the input XHTML document.
- **Concurrent Translation**: Translates multiple text nodes in parallel to speed up the process.
- **Local LLM Integration**: Works with any local inference server compatible with the configured API structure (defaulting to Ollama style).
- **OpenAPI Documentation**: Includes Swagger UI compatible specs.

## Usage

### Prerequisites

- Go 1.21+
- A local LLM server (e.g., [Ollama](https://ollama.com/)) running `google/translategemma-4b-it` or similar.

### Running the Server

```bash
go run cmd/server/main.go --port 8080 --llm-url http://localhost:11434/api/generate --model google/translategemma-4b-it
```

### API Endpoint

**POST** `/translate`

**Request Body:**

```json
{
  "xhtml": "<div><h1>Hello World</h1><p>This is a test.</p></div>",
  "source_lang": "en",
  "target_lang": "es"
}
```

**Response:**

```json
{
  "translated_xhtml": "<div><h1>Hola Mundo</h1><p>Esto es una prueba.</p></div>",
  "metadata": {
    "duration": 123456789,
    "model": "google/translategemma-4b-it",
    "timestamp": "2023-10-27T10:00:00Z"
  }
}
```

## Testing

Run unit tests:

```bash
go test ./...
```

## Architecture

- `cmd/server`: Main entry point.
- `internal/translator`: Core logic for traversal and concurrency.
- `internal/llm`: Client for the local model.
- `internal/api`: HTTP handlers.
- `docs`: OpenAPI specifications.
