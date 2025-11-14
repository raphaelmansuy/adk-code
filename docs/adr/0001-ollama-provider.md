# ADR 0001: Add Ollama model provider

## Status
Accepted

## Context

- The current CLI ships with Gemini/Vertex and OpenAI backends only.
  Issue #1 requests a new provider for Ollama that can work against the
  local server and the remote `ollama.com` endpoint using the official
  Go support library (`github.com/ollama/ollama/api`).
- Ollama exposes a streaming REST API (`/api/generate`, `/api/chat`)
  with dynamic model names (`model:tag`) and environment-aware
  configuration variables (`OLLAMA_HOST`, `OLLAMA_AUTH`, etc.) described in
  `docs/api.md` and `envconfig/config.go`.
- The new provider must support the same ADK `model.LLM` interface and be
  configurable via CLI flags/environment variables similar to existing
  backends.

## Decision

1. Integrate the official Ollama Go package as a first-class provider.
   - Create a new `models` factory that constructs an `ollama.Client`
     using `envconfig.Host()` and optional `OLLAMA_HOST` overrides.
   - Map the client’s streaming responses to `genai.Content` so the
     ADK agent can continue to use the same tooling layer.
2. Extend provider metadata and registry plumbing to include `ollama`.
   - Add `ProviderOllama` to `pkg/models/provider.go` and wire it into the
     backend registry (`internal/llm/provider.go`).
   - Register an `ollama` catalog entry so users can resolve the backend via
     `--model ollama/<model>` but allow unknown model names at runtime by
     falling back to the literal identifier when no alias exists.
3. Document new configuration options.
   - Surface `OLLAMA_HOST`, `OLLAMA_API_KEY` (for cloud models), streaming
     behavior, and how to choose which server/model to run in docs/README.
   - Ensure ADR and future README entries point to the official API
     documentation (`https://docs.ollama.com/api`).

## Consequences

- We gain a dynamic Ollama provider that can use any locally installed model
  without pre-registrations.
- Additional conversion code is required to translate between HTTP/NDJSON
  responses and `genai.Content` while preserving tool call metadata.
- The CLI must continue to work when `OLLAMA_HOST` defaults to
  `http://127.0.0.1:11434` but also allow overriding for remote endpoints.
- Future work may include optional model discovery, caching, and tighter CLI
  prompts for local vs. cloud modes, but those are outside this ADR.
