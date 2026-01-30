# Copilot OpenAI-Compatible Server

An HTTP server that exposes OpenAI-compatible API endpoints (`/v1/chat/completions`, `/v1/models`), powered by the [GitHub Copilot SDK](https://github.com/github/copilot-sdk).

This allows you to use GitHub Copilot's agentic capabilities with any client that supports the OpenAI API (e.g., Open WebUI, custom scripts, etc.).

## Prerequisites

1.  **GitHub Copilot CLI** must be installed and authenticated.
    *   Install via GitHub CLI: `gh extension install github/gh-copilot`
    *   Or ensure the `copilot` binary is in your PATH.
    *   Verify installation: `copilot --version`
2.  **Go 1.21+** installed.

## Docker

You can run the server using Docker, which includes the GitHub Copilot CLI.

### Build the Docker Image

```bash
docker build -t copilot-openai-server .
```

### Run the Container

Set your GitHub personal access token with Copilot permissions:

```bash
export GH_TOKEN="your_personal_access_token"
docker run -e GH_TOKEN=$GH_TOKEN -p 8080:8080 copilot-openai-server
```

**Optional:** Enable API key validation:

```bash
export GH_TOKEN="your_personal_access_token"
export API_KEY="your-secret-key"
docker run -e GH_TOKEN=$GH_TOKEN -e API_KEY=$API_KEY -p 8080:8080 copilot-openai-server
```

**GitHub personal access token (PAT)** must be fine-grained with the `Copilot requests` permission enabled; see https://github.com/github/copilot-cli/issues/91 for details.

The server will be available at `http://localhost:8080`.

### Versioning note

The Dockerfile pulls the latest Copilot SDK and Copilot CLI at build time instead of pinning specific releases ([Dockerfile](Dockerfile)). If a future SDK or CLI update introduces an incompatible change, the image may break until you pin or update both components together.

## Quick Start

### 1. Build the Server

```bash
git clone https://github.com/RajatGarga/copilot-openai-server.git
cd copilot-openai-server
go build -o copilot-server .
```

### 2. Run the Server

```bash
# Default port is 8080
./copilot-server

# Specify a custom port
./copilot-server -port 9000
```

## API Key Authentication (Optional)

You can optionally secure your server with an API key by setting the `API_KEY` environment variable:

```bash
# Run with API key validation enabled
export API_KEY="your-secret-key-here"
./copilot-server
```

**Behavior:**
- **If `API_KEY` is set:** The server validates the `Authorization` header in all requests. Requests must include `Authorization: Bearer your-secret-key-here` or the API key directly.
- **If `API_KEY` is not set:** The server accepts any API key (default behavior, no validation).

**Example with API key:**

```bash
# Start server with API key
export API_KEY="my-secure-key-123"
./copilot-server

# Make requests with the API key
curl http://localhost:8080/v1/models \
  -H "Authorization: Bearer my-secure-key-123"
```

## API Endpoints

### List Models (`GET /v1/models`)

Lists available models from the Copilot SDK.

```bash
curl http://localhost:8080/v1/models
```

### Chat Completions (`POST /v1/chat/completions`)

Supports standard OpenAI chat completion parameters, including streaming and tool calling.

**Basic Request:**

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [
      {"role": "system", "content": "You are a helpful assistant."},
      {"role": "user", "content": "Hello! How does this server work?"}
    ]
  }'
```

**Streaming Request:**

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "Write a short poem about coding."}],
    "stream": true
  }'
```

**Tool Calling:**

The server supports defining tools in the OpenAI format. These are passed to the Copilot SDK, which allows the model to "call" them. Note that the **client** is responsible for executing the tool and sending the result back if needed (standard OpenAI flow).

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "What'\''s the weather in London?"}],
    "tools": [{
      "type": "function",
      "function": {
        "name": "get_weather",
        "description": "Get current weather",
        "parameters": {
          "type": "object",
          "properties": {
            "location": {"type": "string"}
          },
          "required": ["location"]
        }
      }
    }]
  }'
```

## Open WebUI Integration

You can easily use this with [Open WebUI](https://docs.openwebui.com/):

1.  Run this server (`./copilot-server`).
2.  In Open WebUI, go to **Settings > Connections > OpenAI**.
3.  Add a new connection:
    *   **API Base URL:** `http://localhost:8080`
    *   **API Key:** 
        - If you set the `API_KEY` environment variable, use that key value.
        - If `API_KEY` is not set, you can use any string (not validated).
4.  Save and select a Copilot model to start chatting.

**Note:** Authentication with GitHub Copilot happens via the CLI on the host machine, separate from the optional API_KEY validation.

## Development

To add this as a dependency in your own Go project:

```bash
go get github.com/github/copilot-sdk/go
```

## License

GPL-3.0 License