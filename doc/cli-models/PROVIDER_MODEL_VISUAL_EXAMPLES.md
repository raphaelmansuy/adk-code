# Provider/Model Selection - Visual Design & Examples

## CLI Syntax Comparison

### Old Approach (Current)
```bash
# Using Gemini API with specific model
code-agent --backend gemini --model gemini-2.5-flash

# Using Vertex AI - verbose and confusing
code-agent \
  --backend vertexai \
  --project my-gcp-project \
  --location us-central1 \
  --model gemini-1.5-pro-vertex

# Switching providers requires understanding model ID suffixes
code-agent --backend gemini --model gemini-1.5-pro        # Remove -vertex
code-agent --backend vertexai --model gemini-1.5-pro-vertex  # Add -vertex
```

**Issues:**
- ❌ Model ID changed based on backend (confusing)
- ❌ Flag order not intuitive
- ❌ Must remember `-vertex` suffix convention
- ❌ No visual hierarchy in `/models` output

---

### New Approach (Proposed)

#### Quick Syntax
```bash
# Fast and clear - provider/model format
code-agent --model gemini/2.5-flash
code-agent --model vertexai/2.5-flash

# Shorthand - remove version numbers
code-agent --model gemini/flash    # Latest flash model for Gemini
code-agent --model gemini/pro      # Latest pro model for Gemini

# Ultra-short - model name only (uses default provider)
code-agent --model flash           # Same as: gemini/flash
code-agent --model pro             # Same as: gemini/pro

# Vertex AI with all required flags
code-agent \
  --model vertexai/2.5-flash \
  --project my-gcp-project \
  --location us-central1
```

**Benefits:**
- ✅ Provider/model clearly separated by `/`
- ✅ No model ID duplication or suffixes
- ✅ Shorthand aliases for frequently used models
- ✅ Intuitive even for first-time users

---

## Interactive Display Changes

### Current `/models` Output

```
════════════════════════════════════════════════════════════════
                    Available AI Models
════════════════════════════════════════════════════════════════

🔷 Gemini API Models:
   ✓ 💵 Gemini 2.5 Flash - Fast, affordable multimodal model...
      Context: 1000000 tokens | Tools: true | Vision: true
   ○ 💵 Gemini 2.0 Flash - Previous generation fast model...
      Context: 1000000 tokens | Tools: true | Vision: true
   ○ 💵 Gemini 1.5 Flash - Earlier flash model...
      Context: 1000000 tokens | Tools: true | Vision: true
   ○ 💎 Gemini 1.5 Pro - Advanced reasoning model...
      Context: 2000000 tokens | Tools: true | Vision: true

🔶 Vertex AI Models:
   ○ 💵 Gemini 2.5 Flash - Fast, affordable model via Vertex AI...
      Context: 1000000 tokens | Tools: true | Vision: true
   ○ 💎 Gemini 1.5 Pro - Advanced model via Vertex AI...
      Context: 2000000 tokens | Tools: true | Vision: true

Use --model flag to select a model (e.g., --model gemini-1.5-pro)
```

**Problems:**
- ❌ Duplication: Same model listed twice
- ❌ IDs confusing: `-vertex` suffix only shown implicitly
- ❌ No shorthand suggestions
- ❌ Unclear which is the "same model with different backend"

---

### Proposed `/providers` Output

```
════════════════════════════════════════════════════════════════
                    Available Providers & Models
════════════════════════════════════════════════════════════════

🔷 Gemini API
   Type: REST API (requires GOOGLE_API_KEY)
   Status: ✓ Configured
   Default Model: gemini/2.5-flash

   📋 Models:
      ✓ gemini/2.5-flash
        ├─ ID: gemini-2.5-flash
        ├─ Cost: 💵 Economy
        ├─ Context: 1,000,000 tokens
        ├─ Tools: ✓ Vision: ✓
        ├─ Latest generation, fastest
        └─ Aliases: gemini/flash, flash (with default provider)

      ○ gemini/2.0-flash
        ├─ ID: gemini-2.0-flash
        ├─ Cost: 💵 Economy
        ├─ Context: 1,000,000 tokens
        ├─ Aliases: gemini/old-flash

      ○ gemini/1.5-flash
        ├─ ID: gemini-1.5-flash
        ├─ Cost: 💵 Economy
        ├─ Context: 1,000,000 tokens

      ○ gemini/1.5-pro
        ├─ ID: gemini-1.5-pro
        ├─ Cost: 💎 Premium
        ├─ Context: 2,000,000 tokens
        └─ Aliases: gemini/pro

───────────────────────────────────────────────────────────────

🔶 Vertex AI
   Type: GCP Native (requires GOOGLE_CLOUD_PROJECT, GOOGLE_CLOUD_LOCATION)
   Status: ⚠️  Not fully configured
   Missing: GOOGLE_CLOUD_PROJECT, GOOGLE_CLOUD_LOCATION
   Default Model: vertexai/2.5-flash

   📋 Models:
      ○ vertexai/2.5-flash
        ├─ Base: gemini-2.5-flash (Vertex AI backend)
        ├─ Cost: 💵 Economy
        ├─ Context: 1,000,000 tokens
        ├─ Tools: ✓ Vision: ✓
        └─ Aliases: vertexai/flash

      ○ vertexai/1.5-pro
        ├─ Base: gemini-1.5-pro (Vertex AI backend)
        ├─ Cost: 💎 Premium
        ├─ Context: 2,000,000 tokens
        └─ Aliases: vertexai/pro

   ℹ️  To use Vertex AI:
       export GOOGLE_CLOUD_PROJECT=your-project-id
       export GOOGLE_CLOUD_LOCATION=us-central1

════════════════════════════════════════════════════════════════

Usage Examples:
   code-agent --model gemini/2.5-flash
   code-agent --model vertexai/1.5-pro --project my-proj --location us-central1
   code-agent --model pro              (uses default provider)

View all available aliases:     /providers --aliases
View provider requirements:     /providers --requirements
```

---

## Model Selection State Diagram

```
User Input (--model flag)
    ↓
    ├─ Contains "/" ?
    │  ├─ YES: Parse as "provider/model"
    │  │  └─ provider/model/alias_level
    │  │     ├─ Level 3: Full ID (gemini-2.5-flash)
    │  │     ├─ Level 2: Short ID (2.5-flash)
    │  │     └─ Level 1: Alias (flash)
    │  │
    │  └─ NO: Treat as shorthand for default provider
    │     └─ Lookup in default provider's aliases
    │
    ├─ Provider Known?
    │  ├─ YES: Resolve model in that provider
    │  └─ NO: Error with suggestions
    │
    ├─ Model Found?
    │  ├─ YES: Check requirements
    │  │  ├─ Gemini API: GOOGLE_API_KEY set?
    │  │  ├─ Vertex AI: GOOGLE_CLOUD_PROJECT & GOOGLE_CLOUD_LOCATION set?
    │  │  └─ Create LLM instance
    │  │
    │  └─ NO: Suggest similar models
    │     └─ Error with alternatives
    │
    └─ Ready to Start Agent ✓
```

---

## Implementation Examples

### Example 1: Simple Model Selection

```bash
$ code-agent --model gemini/2.5-flash

════════════════════════════════════════════════════════════════
                  Code Agent v1.0.0
════════════════════════════════════════════════════════════════

📦 Provider: Gemini API
   Model: Gemini 2.5 Flash
   Context: 1,000,000 tokens
   Cost: 💵 Economy
   Status: ✓ Ready

Working Directory: /Users/raphael/projects/my-app
Session: default

════════════════════════════════════════════════════════════════
```

### Example 2: Vertex AI with Shorthand

```bash
$ code-agent --model vertexai/pro --project my-gcp-proj --location us-central1

════════════════════════════════════════════════════════════════
                  Code Agent v1.0.0
════════════════════════════════════════════════════════════════

📦 Provider: Vertex AI
   Model: Gemini 1.5 Pro
   Project: my-gcp-proj
   Location: us-central1
   Context: 2,000,000 tokens
   Cost: 💎 Premium
   Status: ✓ Ready

Working Directory: /Users/raphael/projects/my-app
Session: default

════════════════════════════════════════════════════════════════
```

### Example 3: Invalid Model Selection with Helpful Error

```bash
$ code-agent --model vertexai/unknown

❌ Error: Model "unknown" not found in provider "vertexai"

Available models in vertexai:
   • vertexai/2.5-flash (gemini-2.5-flash)
   • vertexai/1.5-pro   (gemini-1.5-pro)

Did you mean?
   • --model vertexai/2.5-flash
   • --model vertexai/1.5-pro

Or switch providers:
   • --model gemini/2.5-flash
   • --model gemini/1.5-pro

For more options:  /providers
```

### Example 4: Vertex AI Not Configured

```bash
$ code-agent --model vertexai/2.5-flash

❌ Error: Vertex AI backend requires additional configuration

Required environment variables not set:
   ❌ GOOGLE_CLOUD_PROJECT
   ❌ GOOGLE_CLOUD_LOCATION

Set these with:
   export GOOGLE_CLOUD_PROJECT=my-project-id
   export GOOGLE_CLOUD_LOCATION=us-central1

Or pass as flags:
   code-agent --model vertexai/2.5-flash \
     --project my-project-id \
     --location us-central1

For more help:  /providers --requirements
```

---

## Backward Compatibility Examples

All current syntax continues to work:

```bash
# Still works - old explicit backend style
code-agent --backend gemini --model gemini-2.5-flash ✓

# Still works - old Vertex AI style
code-agent --backend vertexai --model gemini-1.5-pro-vertex ✓
  (auto-detects and strips -vertex suffix)

# Still works - env var detection
export GOOGLE_GENAI_USE_VERTEXAI=true
code-agent ✓
  (uses Vertex AI with default model)

# Still works - API key only
export GOOGLE_API_KEY=sk-...
code-agent ✓
  (uses Gemini API with default model)
```

---

## Shorthand Alias Examples

```
Provider: gemini
═════════════════════════════════════════════════════════════════
Model ID                  │ Shorthand (Level 2) │ Alias (Level 1)
─────────────────────────────────────────────────────────────────
gemini-2.5-flash          │ gemini/2.5-flash    │ gemini/flash
gemini-2.0-flash          │ gemini/2.0-flash    │ gemini/old-flash
gemini-1.5-flash          │ gemini/1.5-flash    │ (none)
gemini-1.5-pro            │ gemini/1.5-pro      │ gemini/pro
─────────────────────────────────────────────────────────────────

Provider: vertexai
═════════════════════════════════════════════════════════════════
Model ID              │ Shorthand (Level 2)   │ Alias (Level 1)
─────────────────────────────────────────────────────────────────
gemini-2.5-flash      │ vertexai/2.5-flash    │ vertexai/flash
gemini-1.5-pro        │ vertexai/1.5-pro      │ vertexai/pro
─────────────────────────────────────────────────────────────────

With default provider = gemini:
═════════════════════════════════════════════════════════════════
Input                 │ Resolves To
─────────────────────────────────────────────────────────────────
--model flash         │ gemini/2.5-flash
--model pro           │ gemini/1.5-pro
--model 2.5-flash     │ gemini/2.5-flash
--model 1.5-pro       │ gemini/1.5-pro
─────────────────────────────────────────────────────────────────
```

---

## Key Design Decisions

### Decision 1: Provider/Model Separator

**Options:**
- `provider/model` ← **CHOSEN** (intuitive, filesystem-like)
- `provider:model` (too SQL-like)
- `provider-model` (ambiguous with dashes in model names)
- `provider.model` (filesystem-like but less common in CLI)

**Rationale:** Most familiar to developers (DNS, file paths, package managers)

---

### Decision 2: Shorthand Alias Levels

**Level 3 (Full):** `gemini-2.5-flash`
- Registry key, internal use

**Level 2 (Explicit):** `gemini/2.5-flash`
- Provider explicitly specified
- Clearest intent

**Level 1 (Shorthand):** `gemini/flash`
- Latest of its type for provider
- Human-friendly

**Level 0 (Ultra-short):** `flash`
- Uses default provider
- Fastest for experienced users

---

### Decision 3: Default Provider Fallback

```
User Input: --model pro

Lookup Order:
1. Try "gemini/pro" (default provider)
   └─ Found! Use gemini/1.5-pro ✓
2. If not found, try "vertexai/pro"
3. If still not found, error with alternatives
```

**Rationale:** Principle of least surprise - users expect default provider

---

### Decision 4: Model Display Order

In `/providers` output, show models by:
1. Default model first (marked with ✓)
2. By version (newest to oldest)
3. By cost tier (economy before premium)

**Rationale:** Most users start with defaults; version matters for updates

---

## Future Extensions

### Add New Provider (e.g., OpenAI)

```bash
# Syntax automatically supports it
code-agent --model openai/gpt-4
code-agent --model openai/gpt-4o

# /providers shows it naturally
🔵 OpenAI API
   Default Model: openai/gpt-4o
   
   📋 Models:
      ○ openai/gpt-4o
      ○ openai/gpt-4-turbo
      ○ openai/gpt-3.5-turbo
```

### Add New Gemini Model Variant

```bash
# When Google releases gemini-3.0-ultra:
code-agent --model gemini/3.0-ultra
code-agent --model gemini/ultra  # Shorthand

# Backward compatible - existing commands unchanged
```

### Configuration File Support

```yaml
# ~/.code_agent/config.yaml
default_provider: gemini
default_model: 2.5-flash

providers:
  gemini:
    api_key: ${GOOGLE_API_KEY}
  
  vertexai:
    project: my-gcp-project
    location: us-central1
  
  openai:
    api_key: ${OPENAI_API_KEY}

# Usage:
code-agent --model pro     # Uses from config
code-agent --project other # Override config
```

---

## Summary Table

| Feature | Current | Proposed |
|---------|---------|----------|
| **Primary Syntax** | `--backend X --model Y` | `--model provider/model` |
| **Shorthand** | None | `provider/alias` |
| **Ultra-shorthand** | None | Model name only |
| **Model Duplication** | Yes (-vertex suffix) | No (single definition) |
| **Provider Visibility** | Implicit in flags | Explicit in syntax |
| **Discoverability** | `/models` (flat) | `/providers` (hierarchical) |
| **Error Messages** | Generic | Provider/model aware |
| **Backward Compat** | N/A | 100% preserved |
| **Extensibility** | Hard (per-backend flags) | Easy (just add provider) |

