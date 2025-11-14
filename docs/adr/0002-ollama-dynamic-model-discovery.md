# ADR 0002: Add Dynamic Model Discovery and Capability Detection for Ollama

## Status

Proposed

## Context

- ADR 0001 established basic Ollama provider support allowing users to specify any model name
  (`--model ollama/<model>`), but without visibility into:
  - What models are actually available on the Ollama server
  - Capabilities of each model (vision support, tool calling, streaming, etc.)
  - Model parameters and characteristics (size, quantization level, model family)
- Users currently must know model names in advance and cannot query the server for available options
- Model discovery can be expensive (hitting the `/api/tags` endpoint), but the list changes infrequently
- The Ollama API (`github.com/ollama/ollama/api`) exposes:
  - `Client.List()` → `/api/tags` - lists all available models with metadata
  - `Client.Show()` → `/api/show` - retrieves detailed information about a specific model
  - Model details include format, family, parameter size, quantization level

## Decision

1. **Add model discovery capability with caching**:
   - Create a `ModelRegistry` package that wraps `ollama.Client.List()`
   - Implement an in-memory cache with TTL (configurable, default 5 minutes) to avoid repeated API calls
   - Support cache invalidation on-demand via CLI flag (`--refresh-models`)
   - Store cached model list with timestamp for expiration checking

2. **Detect and expose model capabilities**:
   - Infer capabilities from model metadata returned by `/api/tags`:
     - **Vision**: detect from model family (e.g., `llava`, `llama-vision`, `minichat-3b`)
     - **Tool Calling**: For v1, assume tool calling is supported by default (since Ollama models generally support function calling); fall back to heuristics only if needed
     - **Streaming**: supported by all Ollama models via `/api/generate` and `/api/chat`
     - **Context Window**: estimate from parameter size or retrieve from model details
   - Add a `Capabilities` struct to represent model features
   - For v1, default tool calling support to `true` for all models unless explicit API metadata indicates otherwise
   - Future versions can extend with runtime probing (test a small request to verify capability support)

3. **Implement generic model introspection tools** (provider-agnostic):
   - **`list-models`**: Generic tool that lists available models across all providers
     - For Ollama: queries `/api/tags`, includes size, family, quantization
     - For Gemini: lists available models via `genai.ListModels()`
     - For OpenAI: lists available models via OpenAI API
     - Returns formatted table with columns: name, provider, size, family, capabilities
     - Filter options: `--provider <name>`, `--family <name>`, `--has-vision`, `--has-tools`
   - **`model-info`**: Gets detailed information about a specific model
     - Works across providers: `--model ollama/<name>`, `--model gemini/<name>`, etc.
     - Shows metadata, capabilities, parameter details (when available)
     - For Ollama: displays quantization level, estimated memory requirements
     - For others: displays rate limits, context window, pricing (if available)
   - **`verify-capability`**: Tests whether a model supports a specific capability (optional)
     - Works across all providers
     - Tests: vision, tool calling, streaming, JSON mode, etc.

4. **Document configuration and usage**:
   - Update README with examples of discovering models across all providers
   - Show how to filter by capability, provider, and model characteristics
   - Document provider-specific behavior (Ollama cache, Gemini pricing, OpenAI rate limits)
   - Document cache behavior for Ollama and `--refresh-models` flag
   - Provide guidance on capability inference accuracy and limitations
   - Point to provider-specific API references:
     - Ollama: `https://docs.ollama.com/api/tags`
     - Gemini: `https://ai.google.dev/docs/models/gemini`
     - OpenAI: `https://platform.openai.com/docs/models`

## Consequences

**Benefits**:

- Users can discover available models across all providers without external tools
- Generic tools work uniformly across Ollama, Gemini, and OpenAI backends
- Filtering by provider and capability helps users select appropriate models
- Ollama caching minimizes overhead after initial discovery
- Seamless integration with existing `--model <provider>/<model>` syntax

**Tradeoffs**:

- First model discovery call adds latency (varies by provider: Ollama 100-500ms, Gemini/OpenAI 200-1000ms)
- V1 assumes all Ollama models support tool calling by default (may not be accurate for smaller/older models)
- Vision capability detection relies on heuristics and model naming (not 100% accurate)
- Ollama cache TTL means changes to model list won't be reflected immediately (unless `--refresh-models` used)
- Provider-specific features (pricing, rate limits) may not be available for all backends

**Future Improvements**:

- V2: Runtime capability probing via test requests to verify tool calling and vision support
- Integration with provider metadata registries for more accurate capability detection
- Support for capability requirements in agent tool selection
- CLI autocomplete for model names based on discovered list
- Alert mechanism when new compatible models become available
- Model recommendation engine based on task requirements

## Implementation Notes

1. Create `internal/ollama/registry.go` for Ollama-specific model discovery and caching logic
2. Extend `internal/llm/provider.go` to add discovery methods to the provider interface
3. Create `pkg/tools/list_models.go` for the generic `list-models` tool
4. Create `pkg/tools/model_info.go` for the generic `model-info` tool
5. Create provider-specific discovery implementations:
   - `internal/ollama/discovery.go` - Ollama model listing with caching
   - `internal/gemini/discovery.go` - Gemini model listing
   - `internal/openai/discovery.go` - OpenAI model listing
6. Wire tools into the enhanced prompt system for agent awareness
7. Add tests for:
   - Ollama cache behavior (TTL, invalidation)
   - Capability detection heuristics across providers
   - Tool output formatting for different providers
   - Cross-provider filtering and sorting
8. Consider performance impact on startup time and add optional lazy loading
