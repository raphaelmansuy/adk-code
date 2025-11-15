# ADR 0008: Vision/Image Analysis Tool Implementation

**Status:** Proposed  
**Date:** 2025-11-15  
**Decision Makers:** Development Team  
**Technical Story:** Adding multimodal vision capabilities to adk-code agent for analyzing images, screenshots, and visual content

## Context and Problem Statement

The adk-code agent currently lacks the ability to analyze and understand images, screenshots, and visual content. While modern LLM providers (Gemini 2.0 Vision, GPT-4 Vision, Claude 3 Vision) all support vision/multimodal capabilities, the agent cannot leverage these powerful features. This creates significant limitations in workflows that require visual understanding:

**Current Gaps:**
- **Debug screenshots**: Cannot analyze error messages shown in screenshots or UI issues
- **Diagram analysis**: Cannot interpret architecture diagrams, flowcharts, or design mockups
- **Visual code review**: Cannot look at design assets or rendered output
- **Accessibility checks**: Cannot verify UI color contrast, spacing, or layout visually
- **Chart/graph analysis**: Cannot read data from visual charts or graphs
- **Document analysis**: Cannot extract text or understand structure from visual documents

**User Scenarios:**
- "Analyze this screenshot of an error and help me fix it"
- "Look at this design mockup and generate the HTML/CSS for it"
- "Examine this architecture diagram and explain the system"
- "Check if this UI meets accessibility standards"
- "Extract the data from this chart"
- "Parse the structure of this PDF screenshot"

This tool fills a critical capability gap, enabling workflows that competing agents (Codex, Cline, OpenHands) already support.

## Decision Drivers

* **User Need**: Modern development requires understanding visual information
* **Model Capability**: All major LLM providers now support vision APIs
* **Competitive Parity**: Codex, Cline, OpenHands all have image/vision support
* **Workflow Enablement**: Opens new use cases (design review, screenshot debugging)
* **Architecture Consistency**: Must follow existing tool patterns in adk-code
* **Model Agnostic**: Should work across all LLM providers (Gemini, OpenAI, Claude, Ollama)
* **Security**: File validation, size limits, format restrictions

## Considered Options

### Option 1: Native Vision Tool Implementation (Chosen)
Implement a dedicated image analysis tool integrated with adk-code's existing tool framework, delegating actual vision work to the LLM provider's vision API.

**Pros:**
- Model-agnostic (works with any provider supporting vision)
- Full control over image preparation and validation
- Can preprocess images (resize, format conversion)
- Integrates seamlessly with adk-code tool patterns
- Can add caching for frequently analyzed images
- Secure with file validation and size limits

**Cons:**
- Requires integration with each model provider's vision API
- Need to handle different vision API formats (base64 vs URL vs local file)
- Image preprocessing adds complexity
- Different models may have different image format support
- Token counting for images varies by provider

### Option 2: Shell-Based Vision (ImageMagick + OCR)
Wrap system tools like ImageMagick, Tesseract OCR, and basic analysis.

**Pros:**
- No external dependencies on LLM provider vision APIs
- Works locally, no API calls needed
- Free and open source

**Cons:**
- Limited to OCR and basic image analysis
- Cannot understand complex visual content semantically
- Heavy reliance on system tools
- Less portable across platforms
- Poor for diagrams, charts, complex layouts

### Option 3: Delegate to LLM Completely (No Tool)
Just pass image data directly in messages without a dedicated tool.

**Pros:**
- Simplest implementation
- No tool abstraction needed
- Direct model capability usage

**Cons:**
- Less controlled image handling
- No validation or preprocessing
- No error handling for unsupported formats
- Hard to track image usage and billing
- Difficult to provide consistent UX across models

### Option 4: Use Third-Party Vision Service (AWS Rekognition, Azure Vision)
Integrate with specialized vision services.

**Pros:**
- Specialized vision capabilities
- High accuracy for specific tasks

**Cons:**
- Adds external dependencies and costs
- Reduces model autonomy
- More complex error handling
- Tightly coupled to specific provider
- Not true multimodal integration

## Decision Outcome

**Chosen Option:** Option 1 - Native Vision Tool Implementation

This option provides the best balance of:
- **Model compatibility** (works with all vision-capable models)
- **User control** (explicit tool for image analysis)
- **Architecture consistency** (follows adk-code tool patterns)
- **Flexibility** (can preprocess and validate images)
- **Security** (size limits, format validation)
- **Simplicity** (delegates vision work to LLM, not building our own vision system)

### Implementation Strategy

Follow the established tool architecture pattern in adk-code, similar to Fetch Web Tool (ADR 0007):

```
adk-code/tools/
├── vision/               # New package for vision tools
│   ├── init.go          # Auto-registration
│   ├── analyze_image.go # Analyze image tool implementation
│   ├── analyze_image_test.go
│   ├── image_utils.go   # Image processing utilities
│   ├── image_utils_test.go
│   └── providers.go     # Provider-specific vision API adapters
```

## Technical Details

### Architecture Integration

The Vision/Image Analysis tool follows the standard tool pattern:

1. **Input/Output Structs**: Define image parameters and analysis results
2. **Tool Constructor**: `NewVisionAnalyzeTool()` creates and registers the tool
3. **Auto-registration**: `init()` function ensures automatic discovery
4. **Category**: `CategorySearchDiscovery` (alongside Fetch Web, Google Search)
5. **Provider Adapters**: Abstractions for Gemini Vision, OpenAI Vision, Claude Vision
6. **Metadata**: Priority and usage hints for LLM guidance

### Core Components

#### 1. Input Type

```go
package vision

// VisionAnalyzeInput defines parameters for image analysis
type VisionAnalyzeInput struct {
    // ImageSource specifies where the image comes from (required)
    // Can be:
    // - "file:/path/to/image.png" - local file
    // - "url:https://example.com/image.png" - remote URL
    // - "data:image/png;base64,..." - embedded base64 data
    ImageSource string `json:"image_source" jsonschema:"Image source: file path, URL, or base64 data URI"`

    // Question is what to analyze about the image (required)
    Question string `json:"question" jsonschema:"Question or task related to the image analysis"`

    // Detail level of analysis (optional, default: "standard")
    // "low" - quick summary
    // "standard" - balanced analysis
    // "high" - detailed comprehensive analysis
    // Note: Higher detail may use more tokens
    DetailLevel *string `json:"detail_level,omitempty" jsonschema:"Analysis detail level: low, standard, high (default: standard)"`

    // MaxTokens limits the response length in tokens (optional, default: 1024)
    MaxTokens *int `json:"max_tokens,omitempty" jsonschema:"Maximum tokens in response (default: 1024)"`

    // IncludeRawOutput indicates whether to include raw image metadata (optional, default: false)
    IncludeRawOutput *bool `json:"include_raw_output,omitempty" jsonschema:"Include raw image metadata in response (default: false)"`

    // AnalysisType specifies the type of analysis (optional, default: "general")
    // "general" - general understanding
    // "ocr" - text extraction
    // "layout" - structural analysis
    // "code" - code analysis from screenshots
    // "accessibility" - a11y compliance check
    // "chart" - data extraction from charts/graphs
    // "diagram" - architecture/flowchart interpretation
    AnalysisType *string `json:"analysis_type,omitempty" jsonschema:"Analysis type: general, ocr, layout, code, accessibility, chart, diagram (default: general)"`

    // AllowRemoteImages enables fetching images from URLs (optional, default: true)
    AllowRemoteImages *bool `json:"allow_remote_images,omitempty" jsonschema:"Allow fetching images from URLs (default: true)"`

    // TimeoutSeconds for image fetch/analysis (optional, default: 30)
    TimeoutSeconds *int `json:"timeout_seconds,omitempty" jsonschema:"Timeout in seconds (default: 30)"`
}
```

#### 2. Output Type

```go
// VisionAnalyzeOutput contains the image analysis results
type VisionAnalyzeOutput struct {
    // Success indicates whether the analysis was successful
    Success bool `json:"success"`

    // Analysis is the detailed analysis result from the vision model
    Analysis string `json:"analysis"`

    // ImageMetadata contains information about the processed image
    ImageMetadata ImageMetadata `json:"image_metadata"`

    // AnalysisType indicates what type of analysis was performed
    AnalysisType string `json:"analysis_type"`

    // TokensUsed provides token usage information (if available from provider)
    TokensUsed *TokenUsage `json:"tokens_used,omitempty"`

    // ProcessingTimeMS is the time taken for analysis in milliseconds
    ProcessingTimeMS int `json:"processing_time_ms"`

    // Error contains error message if analysis failed
    Error string `json:"error,omitempty"`

    // ErrorCode provides machine-readable error classification
    // "unsupported_format", "file_too_large", "invalid_source", "timeout", 
    // "vision_api_error", "invalid_image", etc.
    ErrorCode string `json:"error_code,omitempty"`

    // ExtractedText contains OCR results if AnalysisType was "ocr" (optional)
    ExtractedText string `json:"extracted_text,omitempty"`

    // StructuredData contains parsed data for specific analysis types (optional)
    // For "chart" type: chart data points
    // For "layout" type: element positions and sizes
    // For "diagram" type: component relationships
    StructuredData map[string]interface{} `json:"structured_data,omitempty"`

    // Warnings contains non-fatal issues during analysis
    Warnings []string `json:"warnings,omitempty"`
}

type ImageMetadata struct {
    // OriginalFilePath is the original file path (if from file)
    OriginalFilePath string `json:"original_file_path,omitempty"`

    // Format is the detected image format (jpeg, png, webp, gif, etc.)
    Format string `json:"format"`

    // Dimensions are the original image dimensions
    Dimensions ImageDimensions `json:"dimensions"`

    // FileSizeBytes is the size of the original image in bytes
    FileSizeBytes int64 `json:"file_size_bytes"`

    // ProcessedFileSizeBytes is the size after any resizing (may differ from original)
    ProcessedFileSizeBytes int64 `json:"processed_file_size_bytes"`

    // HasTransparency indicates if the image has an alpha channel
    HasTransparency bool `json:"has_transparency"`

    // ColorSpace describes the color space (RGB, RGBA, CMYK, etc.)
    ColorSpace string `json:"color_space,omitempty"`
}

type ImageDimensions struct {
    Width  int `json:"width"`
    Height int `json:"height"`
}

type TokenUsage struct {
    InputTokens  int `json:"input_tokens,omitempty"`
    OutputTokens int `json:"output_tokens,omitempty"`
    TotalTokens  int `json:"total_tokens,omitempty"`
}
```

#### 3. Handler Function

```go
// VisionAnalyzeHandler implements the image analysis logic
func VisionAnalyzeHandler(ctx tool.Context, input VisionAnalyzeInput) VisionAnalyzeOutput {
    startTime := time.Now()
    output := VisionAnalyzeOutput{
        Success:      false,
        AnalysisType: getAnalysisType(input.AnalysisType),
    }

    // 1. Validate and load image
    imageData, metadata, err := loadImage(ctx, input.ImageSource, input.AllowRemoteImages)
    if err != nil {
        output.Error = err.Error()
        output.ErrorCode = classifyImageError(err)
        return output
    }
    output.ImageMetadata = metadata

    // 2. Validate image constraints
    if err := validateImage(imageData, metadata); err != nil {
        output.Error = err.Error()
        output.ErrorCode = "invalid_image"
        return output
    }

    // 3. Prepare image for vision API
    // Some APIs require resizing, format conversion, or base64 encoding
    processedImage, err := prepareImageForAPI(ctx, imageData, metadata)
    if err != nil {
        output.Error = fmt.Sprintf("Failed to prepare image: %v", err)
        output.ErrorCode = "image_preparation_error"
        return output
    }

    // 4. Get vision provider from current model context
    provider, err := getVisionProviderForModel(ctx)
    if err != nil {
        output.Error = fmt.Sprintf("Vision not supported by current model: %v", err)
        output.ErrorCode = "vision_api_error"
        return output
    }

    // 5. Call vision API with appropriate prompt engineering
    request := VisionRequest{
        ImageData:     processedImage,
        Question:      input.Question,
        DetailLevel:   getDetailLevel(input.DetailLevel),
        AnalysisType:  output.AnalysisType,
        MaxTokens:     getMaxTokens(input.MaxTokens),
    }

    response, err := provider.AnalyzeImage(ctx, request)
    if err != nil {
        output.Error = fmt.Sprintf("Vision API error: %v", err)
        output.ErrorCode = "vision_api_error"
        return output
    }

    // 6. Process analysis response
    output.Success = true
    output.Analysis = response.Analysis
    output.TokensUsed = response.TokenUsage
    output.ProcessingTimeMS = int(time.Since(startTime).Milliseconds())

    // 7. Extract structured data if applicable
    if response.StructuredData != nil {
        output.StructuredData = response.StructuredData
    }

    if response.ExtractedText != "" {
        output.ExtractedText = response.ExtractedText
    }

    // 8. Collect warnings
    if response.Warnings != nil {
        output.Warnings = response.Warnings
    }

    return output
}
```

#### 4. Provider Adapters

```go
// VisionProvider abstracts different LLM vision APIs
type VisionProvider interface {
    // Name returns the provider name
    Name() string

    // AnalyzeImage calls the provider's vision API
    AnalyzeImage(ctx context.Context, req VisionRequest) (*VisionResponse, error)

    // SupportsFormat checks if the provider supports an image format
    SupportsFormat(format string) bool

    // MaxImageSize returns the maximum image size in bytes
    MaxImageSize() int64
}

// GeminiVisionProvider wraps Gemini 2.0 Vision API
type GeminiVisionProvider struct {
    client *generativelanguage.Client
}

// OpenAIVisionProvider wraps GPT-4 Vision API
type OpenAIVisionProvider struct {
    client *openai.Client
}

// ClaudeVisionProvider wraps Claude 3 Vision API
type ClaudeVisionProvider struct {
    client *anthropic.Client
}

// Local image analysis using stdlib for accessibility checks
type LocalVisionProvider struct {
    // Can perform basic analysis without API call
}
```

### Security Considerations

**Image Validation:**
- Only allow supported formats (JPEG, PNG, WebP, GIF, PDF)
- Validate image headers (magic bytes)
- Enforce size limits (default 20 MB, configurable up to 100 MB)
- Check image dimensions (reasonable bounds)

**File Operations:**
- Only accept local file paths within workspace root
- Prevent path traversal attacks
- Validate remote URLs (https only, domain whitelist optional)

**API Security:**
- Never log or store actual image data
- Use model context for API key retrieval (not in input)
- Rate limiting per image type and user
- Timeout enforcement (default 30s, max 5 minutes)

**Data Privacy:**
- Images sent to LLM providers are subject to their privacy policies
- Document this clearly to users
- Option to use local analysis for sensitive images

### Model Compatibility

The Vision Image Analysis tool supports **all major vision-capable models**:

✅ **Gemini** - Full support (Gemini 2.0 Vision, 1.5 Pro Vision)  
✅ **OpenAI** - Full support (GPT-4 Vision, GPT-4 Turbo with Vision)  
✅ **Claude** - Full support (Claude 3 Opus, Sonnet, Haiku)  
✅ **Ollama** - Support via compatible models (llava, bakllava)  
⚠️ **Local models** - Limited support (basic image analysis via stdlib)

Unlike text-only models, this tool will gracefully degrade with clear error messages for unsupported models.

### Complementary to Existing Tools

| Tool | Purpose | Relationship |
|------|---------|--------------|
| **Fetch Web** | Get content from URLs (text/JSON) | Vision for visual web content |
| **Read File** | Read file contents (text) | Vision for visual file content |
| **Execute Program** | Run commands | Vision for screenshot output analysis |
| **Google Search** | Find web resources | Vision for analyzing search results visually |
| **Vision** (NEW) | Analyze images visually | Complementary to all text-based tools |

**Combined Workflow Example:**
```
User: "Fetch the design from this URL and implement it"
1. Fetch Web Tool → Retrieves HTML/screenshots
2. Vision Tool → Analyzes design visually
3. Edit Tools → Implements based on analysis
4. Execute Program → Tests the implementation
5. Vision Tool → Analyzes test results/screenshots
```

## Code Architecture

### Directory Structure

```
adk-code/
├── adk-code/
│   └── tools/
│       ├── vision/              # Image analysis tools (NEW - this ADR)
│       │   ├── analyze_image.go # Vision analysis tool
│       │   ├── analyze_image_test.go
│       │   ├── image_utils.go   # Image loading/validation
│       │   ├── image_utils_test.go
│       │   ├── providers.go     # Vision provider adapters
│       │   ├── providers_test.go
│       │   └── init.go          # Auto-registration
│       ├── websearch/           # Google Search (ADR 0005)
│       ├── web/                 # Fetch Web (ADR 0007)
│       ├── file/                # File operations
│       ├── edit/                # Code editing
│       ├── exec/                # Command execution
│       └── base/                # Base registry & types
└── docs/
    └── adr/
        ├── 0005-google-search-tool-integration.md
        ├── 0007-fetch-web-tool.md
        └── 0008-vision-image-analysis-tool.md  # This ADR
```

### Integration Points

1. **Tool Registry** (`tools/base/registry.go`)
   - Register in `CategorySearchDiscovery` (alongside vision-related tools)
   - Priority 2 (after Google Search and Fetch Web)
   - Usage hints for vision capabilities

2. **Exports** (`tools/tools.go`)
   - Export `VisionAnalyzeInput`, `VisionAnalyzeOutput` types
   - Export `NewVisionAnalyzeTool()` constructor

3. **Model Context** (`pkg/models/*`)
   - Check model capabilities for vision support
   - Route vision requests appropriately based on model

4. **Agent Loop** (`pkg/agents/agent.go`)
   - No changes needed; agent auto-discovers tool
   - Tool available in all agent contexts with vision models

5. **Display/REPL** (`internal/display/tools/`)
   - Format image analysis results with visual metrics
   - Show image metadata alongside analysis
   - Display processing time and token usage

## Usage Examples

### Example 1: Analyze Error Screenshot

```
User: "I got this error when running the build. Fix it."
[User pastes screenshot of error]

Agent calls:
{
    "tool": "builtin_vision_analyze",
    "input": {
        "image_source": "data:image/png;base64,...",
        "question": "What is the error message and what might be causing it?",
        "analysis_type": "ocr"
    }
}

Response:
{
    "success": true,
    "analysis": "The error shows 'TypeError: Cannot read property map of undefined' in main.js at line 42...",
    "extracted_text": "[Full error text extracted]",
    "image_metadata": {
        "format": "png",
        "dimensions": {"width": 1920, "height": 1080}
    }
}
```

### Example 2: Design to Code

```
User: "Generate React component code for this design"
[User provides screenshot of design mockup]

Agent calls:
{
    "tool": "builtin_vision_analyze",
    "input": {
        "image_source": "file:/workspace/designs/header.png",
        "question": "Describe the layout, colors, typography, and UI elements in this design",
        "analysis_type": "layout",
        "detail_level": "high"
    }
}

Response:
{
    "success": true,
    "analysis": "A responsive header with: dark blue background (#1a1a2e), white text, logo on left, navigation links in center, user menu on right. Font: Inter. Spacing: 16px padding. Button colors: primary blue (#3366ff), secondary gray (#666).",
    "structured_data": {
        "colors": {
            "background": "#1a1a2e",
            "text": "#ffffff",
            "primary": "#3366ff"
        },
        "layout": "flex",
        "components": ["logo", "nav", "user-menu"]
    }
}
```

### Example 3: Architecture Diagram Interpretation

```
User: "Explain the architecture shown in this diagram"

Agent calls:
{
    "tool": "builtin_vision_analyze",
    "input": {
        "image_source": "file:/workspace/docs/architecture.png",
        "question": "What are the main components and how do they interact?",
        "analysis_type": "diagram"
    }
}

Response:
{
    "success": true,
    "analysis": "The system consists of: Frontend (React SPA), API Gateway, Microservices (Auth, User, Orders, Payments), Databases (PostgreSQL, Redis), Message Queue (RabbitMQ). Data flows from Frontend → Gateway → Services → Databases.",
    "structured_data": {
        "components": ["Frontend", "API Gateway", "Auth Service", "User Service", "Orders Service", "Payments Service", "PostgreSQL", "Redis", "RabbitMQ"],
        "connections": [
            {"from": "Frontend", "to": "API Gateway"},
            {"from": "API Gateway", "to": "Auth Service"}
        ]
    }
}
```

### Example 4: Handle Vision Errors

```
User: "Analyze this unsupported image format"

Agent calls:
{
    "tool": "builtin_vision_analyze",
    "input": {
        "image_source": "file:/workspace/image.bmp"
    }
}

Response:
{
    "success": false,
    "error": "Image format BMP not supported",
    "error_code": "unsupported_format",
    "image_metadata": {
        "format": "bmp"
    },
    "warnings": ["BMP format is not supported. Please use PNG, JPEG, WebP, or GIF."]
}
```

## Consequences

### Positive Impacts

✅ **Visual Understanding**: Agent can now interpret screenshots, diagrams, mockups  
✅ **Error Debugging**: Analyze error messages in screenshots  
✅ **Design Implementation**: Generate code from visual designs  
✅ **Accessibility**: Check UI compliance visually  
✅ **Model Parity**: Match competitors (Codex, Cline, OpenHands)  
✅ **New Workflows**: Enable screenshot-based debugging, visual analysis  
✅ **User Flexibility**: Users can share visual context directly  
✅ **Comprehensive Context**: Combine visual + text understanding  

### Potential Challenges

⚠️ **Provider Limitations**: Different models have different image support  
⚠️ **Token Cost**: Vision analysis uses more tokens than text  
⚠️ **Processing Time**: Image analysis takes longer than text (network latency)  
⚠️ **Format Restrictions**: Not all image formats supported by all providers  
⚠️ **Interpretation Bias**: AI may misinterpret complex or ambiguous visuals  
⚠️ **Privacy**: Images sent to LLM providers (user must be aware)  

**Mitigation:**
- Document model vision capabilities clearly
- Provide token cost estimates before analysis
- Include provider fallback logic
- Clear error messages for unsupported formats
- Privacy notice in tool documentation

### Resource Impact

| Resource | Impact | Notes |
|----------|--------|-------|
| Memory | ~50-100 MB per large image | Image preprocessing and temp storage |
| Network | Higher latency | Vision APIs typically slower than text |
| API Cost | 2-5x text costs | Vision tokens more expensive |
| CPU | Medium impact | Image format conversion/resizing |
| Storage | Minimal | No persistent caching by default |

## Implementation Checklist

### Phase 1: Image Loading & Validation
- [ ] Create `tools/vision/` directory structure
- [ ] Implement image loading from files, URLs, and base64 data
- [ ] Implement image format validation (magic bytes)
- [ ] Implement size and dimension checking
- [ ] Handle EXIF data and image metadata extraction
- [ ] Write unit tests for image validation

### Phase 2: Provider Adapters
- [ ] Implement Gemini Vision provider adapter
- [ ] Implement OpenAI Vision provider adapter
- [ ] Implement Claude Vision provider adapter
- [ ] Implement Ollama vision model support (if available)
- [ ] Implement local image analysis provider (basic)
- [ ] Provider feature detection and fallback logic
- [ ] Write provider unit tests

### Phase 3: Vision Tool Implementation
- [ ] Implement `VisionAnalyzeInput` and `VisionAnalyzeOutput` types
- [ ] Implement `VisionAnalyzeHandler` with all analysis types
- [ ] Implement prompt engineering for different analysis types
- [ ] Add image preprocessing (resizing, format conversion)
- [ ] Handle provider-specific image encoding
- [ ] Write comprehensive unit tests

### Phase 4: Tool Registration & Integration
- [ ] Create `tools/vision/init.go` with `NewVisionAnalyzeTool()`
- [ ] Register tool with `CategorySearchDiscovery`
- [ ] Add exports to `tools/tools.go`
- [ ] Verify auto-registration in tool discovery
- [ ] Test tool appears in `/tools` REPL command
- [ ] Test tool help in `/help` command

### Phase 5: Testing & Validation
- [ ] Integration test with Gemini Vision model
- [ ] Integration test with OpenAI GPT-4 Vision
- [ ] Integration test with Claude 3 Opus
- [ ] Test all analysis types (ocr, layout, code, etc.)
- [ ] Test error handling (unsupported formats, too large, etc.)
- [ ] Test image preprocessing workflows
- [ ] Security validation (path traversal, format validation)
- [ ] Performance testing with various image sizes
- [ ] Run `make check` successfully
- [ ] Run full test suite

### Phase 6: Documentation & Deployment
- [ ] Document tool usage in README.md
- [ ] Add examples to TOOL_DEVELOPMENT.md
- [ ] Document vision API provider support matrix
- [ ] Document token cost considerations
- [ ] Document privacy and data handling
- [ ] Create troubleshooting guide for common issues
- [ ] Update ARCHITECTURE.md with vision tool section
- [ ] Update CHANGELOG.md with new feature
- [ ] Create release notes with migration guide

## Testing Strategy

### Unit Tests

```go
func TestLoadImage_FromFile(t *testing.T)
func TestLoadImage_FromURL(t *testing.T)
func TestLoadImage_FromBase64(t *testing.T)
func TestValidateImage_SupportedFormats(t *testing.T)
func TestValidateImage_UnsupportedFormats(t *testing.T)
func TestValidateImage_SizeLimit(t *testing.T)
func TestValidateImage_DimensionBounds(t *testing.T)
func TestPrepareImage_Resizing(t *testing.T)
func TestPrepareImage_FormatConversion(t *testing.T)
func TestVisionAnalyze_OCR(t *testing.T)
func TestVisionAnalyze_Layout(t *testing.T)
func TestVisionAnalyze_Code(t *testing.T)
func TestVisionAnalyze_Diagram(t *testing.T)
func TestGeminiVisionProvider_SupportsFormat(t *testing.T)
func TestOpenAIVisionProvider_SupportsFormat(t *testing.T)
```

### Integration Tests

```go
func TestVisionAnalyze_Integration_Gemini(t *testing.T)
func TestVisionAnalyze_Integration_OpenAI(t *testing.T)
func TestVisionAnalyze_Integration_Claude(t *testing.T)
func TestVisionAnalyze_Integration_MultiFormat(t *testing.T)
func TestVisionAnalyze_Integration_LargeImage(t *testing.T)
func TestVisionAnalyze_Combined_WithFetchWeb(t *testing.T)
```

### Manual Testing

```bash
# Build and test
cd adk-code
make build
make test

# Test vision tool in REPL with screenshot
./bin/adk-code

# In REPL:
> /tools  # Verify vision_analyze appears
> Analyze this screenshot: [screenshot]
> /help vision_analyze  # Show tool help

# Test with different models:
./bin/adk-code --model gemini/2.0-flash
./bin/adk-code --model openai/gpt-4-vision
./bin/adk-code --model claude/3-opus
```

### Test Data

- Various image formats (PNG, JPEG, WebP, GIF)
- Different image sizes (small, medium, large, oversized)
- Different image types (screenshots, diagrams, charts, designs)
- EXIF metadata in images
- Corrupted image files
- Unsupported formats
- Very large dimensions
- Various color spaces and transparency

## Future Enhancements

1. **Batch Image Analysis**: Analyze multiple images in one request
2. **Image Modification**: Generate modified versions of images (crop, highlight, annotate)
3. **Comparison Analysis**: Compare two images side-by-side
4. **Optical Character Recognition (OCR)**: Enhanced text extraction with language support
5. **Document Processing**: Multi-page document analysis (PDFs)
6. **Video Frame Analysis**: Extract and analyze key frames from videos
7. **Image Caching**: Cache frequently analyzed images
8. **Chunk Analysis**: Analyze specific regions of large images
9. **Accessibility Scoring**: Automated accessibility compliance scoring
10. **Performance Profiling**: Visual memory/CPU usage analysis from screenshots

## Alternative Approaches Rejected

### Why not implement our own vision model?
- Adds enormous complexity and computational cost
- Modern LLMs already have excellent vision capabilities
- Would require GPU infrastructure
- Maintenance burden too high for small team

### Why not just use system commands (ImageMagick + Tesseract)?
- Limited to OCR and basic analysis
- Cannot understand semantic content
- Poor for diagrams, charts, design analysis
- Doesn't match capabilities of LLM vision APIs

### Why not delegate entirely to LLM without a tool?
- Loss of control over image handling
- No validation or security checks
- Poor error handling
- Difficult to track vision API usage

## References

### Vision API Documentation
- **Gemini Vision API**: https://ai.google.dev/gemini-2/docs
- **OpenAI GPT-4 Vision**: https://platform.openai.com/docs/guides/vision
- **Claude 3 Vision**: https://docs.anthropic.com/en/api/vision

### Image Processing in Go
- **image package**: https://pkg.go.dev/image
- **image/jpeg**: https://pkg.go.dev/image/jpeg
- **image/png**: https://pkg.go.dev/image/png
- **github.com/kolesa-team/go-webp**: WebP support

### Related ADRs
- [ADR 0005: Google Search Tool Integration](./0005-google-search-tool-integration.md)
- [ADR 0007: Fetch Web Tool](./0007-fetch-web-tool.md)
- [ARCHITECTURE.md - Tool System](../ARCHITECTURE.md#tool-system)
- [TOOL_DEVELOPMENT.md](../TOOL_DEVELOPMENT.md)

## Approval & Sign-Off

| Role | Status | Date |
|------|--------|------|
| Architecture Lead | Pending | - |
| Implementation Lead | Pending | - |
| QA Lead | Pending | - |

---

## Implementation Notes

### Starting Point

Begin with Phase 1 (Image Loading & Validation):

1. Create directory: `adk-code/tools/vision/`
2. Copy template from `tools/file/read_tool.go` for structure
3. Implement image loading from files, URLs, and base64
4. Implement format validation using magic bytes
5. Implement size and dimension checking
6. Create simple unit tests using test images
7. Get feedback before moving to provider adapters

### Key Implementation Details

- **Image Loading**: Support file://, url://, data: URIs
- **Format Detection**: Use magic bytes, not extensions
- **Preprocessing**: Resize large images to API limits
- **Error Handling**: Clear error codes for each failure mode
- **Provider Selection**: Auto-detect from model context
- **Prompt Engineering**: Different prompts for different analysis types

### Common Pitfalls to Avoid

1. **API Rate Limiting**: Vision API calls may be throttled
2. **Token Explosions**: Large images can use huge token counts
3. **Format Incompatibility**: Not all formats work with all providers
4. **Timeouts**: Vision APIs are slower than text APIs
5. **Privacy Issues**: Must warn users about image transmission
6. **Memory Bloat**: Don't keep image data in memory longer than needed
7. **Model Limitations**: Graceful handling when vision not supported

## Comparison with Competitors

### Codex (OpenAI)
- Supports image input via `-i/--image` flag
- Can paste images directly in composer
- Analyzes for errors, design, screenshots

### Cline
- Browser integration for visual testing
- Image input for design analysis
- Screenshot analysis for debugging

### OpenHands
- BrowseURLAction for visual web content
- Can capture and analyze screenshots
- Limited but functional vision support

**adk-code Vision Tool Advantages:**
- Multi-provider support (not just OpenAI)
- Structured analysis types (OCR, layout, code, etc.)
- Works with all major LLM providers
- Explicit tool with clear inputs/outputs
- Better error handling and validation

## See Also

- [ADR 0005: Google Search Tool Integration](./0005-google-search-tool-integration.md) - Complementary discovery tool
- [ADR 0007: Fetch Web Tool](./0007-fetch-web-tool.md) - Complementary content fetching
- [TOOL_DEVELOPMENT.md](../TOOL_DEVELOPMENT.md) - Tool development patterns
- [Vision API Best Practices](https://ai.google.dev/docs/vision_best_practices)
- [Claude Vision Guide](https://docs.anthropic.com/en/api/vision)
