# Examples

This directory contains example programs demonstrating various features of AINative Code.

## Available Examples

### 1. Embeddings Example
Location: `examples/embeddings/`

Demonstrates how to work with embeddings and vector operations.

```bash
go run ./examples/embeddings
```

### 2. Gemini Provider Example
Location: `examples/gemini/`

Shows how to use the Google Gemini LLM provider.

```bash
go run ./examples/gemini
```

### 3. OpenAI Provider Example
Location: `examples/openai_provider/`

Demonstrates integration with OpenAI's API.

```bash
go run ./examples/openai_provider
```

## Building Examples

Each example is in its own subdirectory and can be built independently:

```bash
# Build a specific example
go build ./examples/embeddings

# Build all examples
go build ./examples/embeddings && \
go build ./examples/gemini && \
go build ./examples/openai_provider
```

## Configuration

Most examples require configuration files. See the example config files in this directory:

- `config.yaml` - Full configuration example
- `config.minimal.yaml` - Minimal configuration
- `config_gemini.yaml` - Gemini-specific configuration
- `config-with-resolver.yaml` - Configuration with environment variable resolution

## Running Examples

1. Copy an appropriate config file to `~/.ainative-code.yaml`
2. Set any required environment variables (API keys, etc.)
3. Run the example using `go run`

## Notes

- Each example is self-contained with its own `main()` function
- Examples demonstrate best practices for using AINative Code APIs
- Configuration files may need to be updated with your credentials
