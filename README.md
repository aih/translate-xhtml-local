# XHTML Translation Service

> [!WARNING]
> **CURRENT STATUS: POOR TRANSLATION QUALITY**
> 
> This application is currently **not producing reliable translations** with the tested local LLMs (e.g., `gemma:2b`).
> 
> **Known Issues:**
> - **Extraneous Text**: The model often includes conversational filler (e.g., "Sure, here is the translation:", "The translated text is:", "In Portuguese this means...").
> - **Incomplete/Incorrect Translations**: Short phrases or headers are sometimes skipped or hallucinated.
> 
> Please refer to the `output_samples/` directory to see examples of these issues (e.g., look for "Sure, here is..." inside the HTML tags).
> 
> **Work in Progress**: We are actively working on prompt engineering and model selection to enforce strict output formats.

This is a Golang application that translates XHTML pages using a local LLM (e.g., Ollama with `google/translategemma-4b-it` or compatible models).

## Features

- **XHTML Parsing**: Preserves the structure of the input XHTML document.
- **Concurrent Translation**: Translates multiple text nodes in parallel to speed up the process.
- **Local LLM Integration**: Works with any local inference server compatible with the configured API structure (defaulting to Ollama style).
- **OpenAPI Documentation**: Includes Swagger UI compatible specs.

## Usage

### Prerequisites

- Go 1.21+
- [Task](https://taskfile.dev/) (optional, for running tasks)
- Docker (optional, for containerization)
- A local LLM server (e.g., [Ollama](https://ollama.com/)) running `google/translategemma-4b-it` or similar.

### Running with Task

```bash
task run
```

### Running with Docker

```bash
task docker-build
task docker-run
```

The server will be available at `http://localhost:8090`.
Swagger UI is available at `http://localhost:8090/swagger/index.html`.

### Manual Execution

```bash
go run cmd/server/main.go --port 8090 --llm-url http://localhost:11434/api/generate --model google/translategemma-4b-it
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
task test
```

### Large Scale Integration Testing

The integration tests verify the system's ability to handle high concurrency and preserve XHTML structure (including images).

**Default Mode (Mock LLM)**:
By default, `task test-integration` uses a **Mock LLM**.
- **Behavior**: It "translates" text by **reversing it** (e.g., `Hello` -> `[es] olleH`).
- **Purpose**: Validates concurrency, file handling, and structure preservation without the latency or GPU cost of a real model.
- **Speed**: Very fast (~2 seconds for 200 files).

```bash
task test-integration
```

**Real LLM Mode**:
To verify actual translation quality with your local model:
- **Behavior**: Sends real requests to the configured LLM endpoint.
- **Purpose**: End-to-end verification of translation logic.
- **Speed**: Depends on your hardware (can take minutes).

```bash
TEST_REAL_LLM=true task test-integration
```

The output files are stored in `test/integration/output`. Sample real translations can be found in `output_samples/`.

## Architecture

- `cmd/server`: Main entry point.
- `internal/translator`: Core logic for traversal and concurrency.
- `internal/llm`: Client for the local model.
- `internal/api`: HTTP handlers.
- `docs`: OpenAPI specifications.
