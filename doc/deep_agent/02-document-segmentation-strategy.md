# Document Segmentation Strategy: Handling Large Files

## Introduction

**The Problem**: LLMs have fixed context windows (e.g., Gemini 2.5 Flash: 1M tokens). A single 100K-line codebase can exceed this. Traditional chunking breaks files blindly, splitting algorithms across chunks.

**DeepCode Solution**: Intelligent semantic segmentation that understands content structure and preserves algorithmic integrity.

**Example**: 
```
❌ Naive approach:
Document: 200K characters
Split into 50K-char chunks at arbitrary position
  Chunk 1: Lines 1-500 (complete but arbitrary boundary)
  Chunk 2: Lines 501-1000 (might split algorithm in half)
  Problem: Related code is now in different chunks

✅ Semantic approach:
Analyze document structure: sections, algorithms, formulas
Identify natural semantic boundaries (end of algorithm, section break)
Chunk intelligently:
  Chunk 1: Intro + Algorithm A (semantic unit)
  Chunk 2: Algorithm B + Implementation (semantic unit)
  Problem solved: Related code stays together
```

---

## Core Concepts

## Core Concepts

### 1. Document Type Detection

The actual implementation in `DocumentAnalyzer` class (lines 52-230 in `document_segmentation_server.py`) uses semantic content analysis instead of just structural markers.

**Semantic Indicators** (from code, lines 63-88):
```python
ALGORITHM_INDICATORS = {
    "high": ["algorithm", "procedure", "method", "approach", "technique", "framework"],
    "medium": ["step", "process", "implementation", "computation", "calculation"],
    "low": ["example", "illustration", "demonstration"],
}

TECHNICAL_CONCEPT_INDICATORS = {
    "high": ["formula", "equation", "theorem", "lemma", "proof", "definition"],
    "medium": ["parameter", "variable", "function", "model", "architecture"],
    "low": ["notation", "symbol", "term"],
}

IMPLEMENTATION_INDICATORS = {
    "high": ["code", "implementation", "programming", "software", "system"],
    "medium": ["design", "structure", "module", "component", "interface"],
    "low": ["tool", "library", "package"],
}
```

**Document Type Detection** (method `analyze_document_type()`, lines 186-223):
- Returns: `Tuple[str, float]` = (document_type, confidence_score)
- Document types: `"research_paper"`, `"technical_doc"`, `"algorithm_focused"`, `"implementation_guide"`, `"general_document"`
- Confidence range: 0.5-0.95 based on pattern matching strength

**Pattern Matching** (lines 90-105):
```python
RESEARCH_PAPER_PATTERNS = [
    r"(?i)\babstract\b.*?\n.*?(introduction|motivation|background)",
    r"(?i)(methodology|method).*?(experiment|evaluation|result)",
    r"(?i)(conclusion|future work|limitation).*?(reference|bibliography)",
    r"(?i)(related work|literature review|prior art)",
]

TECHNICAL_DOC_PATTERNS = [
    r"(?i)(getting started|installation|setup).*?(usage|example)",
    r"(?i)(api|interface|specification).*?(parameter|endpoint)",
    r"(?i)(tutorial|guide|walkthrough).*?(step|instruction)",
    r"(?i)(troubleshooting|faq|common issues)",
]
```

### 2. Semantic Boundary Detection

The actual implementation in `DocumentSegmenter` class (document_segmentation_server.py) identifies semantic boundaries through specialized detection methods.

**Algorithm Block Identification** (method `_identify_algorithm_blocks()`, lines 661-730):
```python
# Identifies complete algorithm blocks with context
# Returns: List of algorithm blocks with preserved structure
# Preserves: Complete pseudocode, step numbering, end markers
```

**Concept Group Extraction** (method `_identify_concept_groups()`, lines 733-785):
```python
# Identifies concept definitions and groupings
# Pattern examples:
# - "theorem|lemma|proposition|corollary"
# - "notation|symbol|parameter"
# Result: Related concepts kept together in same chunk
```

**Formula Chain Recognition** (method `_identify_formula_chains()`, lines 788-854):
```python
# Identifies mathematical formula sequences
# Preserves: $$...$$  block formulas and $...$ inline formulas
# Merges: Nearby formulas within 500 characters
# Output: Formula chains with context (200 chars before/after)
```

**Content Block Merging** (method `_merge_related_content_blocks()`, lines 857-930):
```python
# Intelligently merges nearby related blocks
# Criteria:
# - If blocks within 300 characters
# - If blocks are semantically related (algorithm+formula, etc.)
# Result: Cohesive chunks, no algorithm splitting
```

**Boundary Type Priorities** (from implementation):
- **Hard Boundaries** (always split): Section headers (###, ==), algorithm starts
- **Smart Merge** (consider joining): Related content within 300 chars, related content types
- **Never Split**: Algorithm blocks, proofs, formula chains

### 3. Query-Aware Retrieval

The actual implementation uses MCP tool `read_document_segments()` (lines 1603-1720 in `document_segmentation_server.py`) with intelligent segment selection.

**Query Types and Scoring** (from code, lines 1648-1670):
```python
def read_document_segments(
    paper_dir: str,
    query_type: str,  # "concept_analysis", "algorithm_extraction", or "code_planning"
    keywords: List[str] = None,
    max_segments: int = 3,
    max_total_chars: int = None,
) -> str:
    """Intelligently retrieve relevant document segments based on query type"""
```

**Relevance Scoring Algorithm** (lines 1671-1695):
```python
scored_segments = []
for segment in document_index.segments:
    # Base relevance score from pre-computed relevance_scores
    relevance_score = segment.relevance_scores.get(query_type, 0.5)
    
    # Enhanced keyword matching with position weighting
    if keywords:
        keyword_score = _calculate_enhanced_keyword_score(segment, keywords)
        relevance_score += keyword_score
    
    # Content completeness bonus (prefer comprehensive segments)
    completeness_bonus = _calculate_completeness_bonus(segment, document_index)
    relevance_score += completeness_bonus
    
    scored_segments.append((segment, relevance_score))

# Sort by enhanced relevance score (highest first)
scored_segments.sort(key=lambda x: x[1], reverse=True)
```

**Adaptive Character Limits** (function `_calculate_adaptive_char_limit()`):
- Concept analysis: Lower char limit (2K-4K), emphasis on intro/overview
- Algorithm extraction: Higher char limit (6K-12K), includes proofs and complexity analysis
- Code planning: Medium-high (5K-10K), combines algorithm + implementation notes

**Segment Selection with Integrity** (function `_select_segments_with_integrity()`):
- Intelligently combines top-ranked segments
- Ensures algorithm blocks are never split across selections
- Maintains narrative flow by including transitions
- Respects max_total_chars budget

---

## Implementation Architecture

### Integration Point in code_agent

```go
// New tool for document-aware processing
type SegmentDocumentInput struct {
    FilePath           string   `json:"file_path"`
    DocumentType       string   `json:"document_type,omitempty"` // Let agent detect if not specified
    PreferredChunkSize int      `json:"preferred_chunk_size,omitempty"` // bytes
    OutputPath         string   `json:"output_path"`
}

type DocumentSegment struct {
    ID            string
    Title         string
    Content       string
    ContentType   string    // "introduction", "algorithm", "proof", "code", "reference"
    StartLine     int
    EndLine       int
    CharCount     int
    Keywords      []string
    SectionPath   string    // "3.2.1" for nested sections
}

type SegmentDocumentOutput struct {
    Success        bool
    SegmentsCount  int
    DocumentType   string
    Strategy       string    // "semantic_research_focused", "algorithm_preserve", etc.
    Segments       []DocumentSegment
    StorePath      string    // Where segments index is saved
    Error          string
}
```

**Usage Flow**:

```go
// Agent receives task: analyze a 100K-line research paper

// Step 1: Segment the document
segmentResult := agent.ExecuteTool("segment_document", SegmentDocumentInput{
    FilePath: "paper.md",
    // DocumentType will be auto-detected
    // OutputPath: ".segments/paper/"
})

// Step 2: Agent determines what it needs
conceptQuery := "What problem does this paper address?"

// Step 3: Retrieve relevant segments
relevantSegments := agent.ExecuteTool("read_document_segments", ReadDocSegmentsInput{
    SegmentIndexPath: segmentResult.StorePath,
    QueryType: "concept_analysis",
    MaxSegments: 3,
})

// Step 4: Agent receives concise, relevant content
// Instead of 100K chars, gets 3-5K chars that directly answer the question
```

### Segmentation Strategies

**Strategy 1: Semantic Research-Focused**
```
For: Academic papers with clear structure
Detects: Sections (1, 2, 3...), subsections (2.1, 2.2)
Preserves: Algorithm blocks, proofs, formula chains
Splits at: Section boundaries (hard breaks)
```

**Strategy 2: Algorithm Preserve Integrity**
```
For: Documents focused on algorithms and pseudocode
Detects: Algorithm keywords, pseudocode blocks, proofs
Preserves: Complete algorithm blocks, proof logic
Splits at: Algorithm boundaries, major proof sections
Never splits: Within pseudocode, within proofs
```

**Strategy 3: Concept-Implementation Hybrid**
```
For: Tutorials, guides combining concepts with code
Detects: Concept explanations, code examples, steps
Preserves: Concept groups, code examples (with explanation)
Splits at: Concept topic boundaries, code example boundaries
```

**Strategy 4: Semantic Chunking Enhanced**
```
For: Long documents with unclear structure
Uses: Advanced boundary detection (LLM-powered)
Detects: Semantic topics regardless of explicit markers
Preserves: Topic coherence, related concepts
Dynamically adjusts: Chunk size based on content density
```

---

## DeepCode Segmentation Features

### Implementation Details

**Document Analyzer** (class `DocumentAnalyzer`, lines 52-230):
- Semantic content analysis using weighted indicator scoring
- Pattern matching for document type detection (confidence 0.5-0.95)
- Configurable detection thresholds

**Document Segmenter** (class `DocumentSegmenter`, lines 233-430):
- Segment strategy selection based on document type and content density
- Specialized detection methods for algorithms, concepts, formulas
- Automatic merging of related content blocks

**Main MCP Tools**:

1. **analyze_and_segment_document()** (lines 1432-1596)
   - Input: `paper_dir`, `force_refresh` (boolean)
   - Process: Load/analyze/segment document, save index
   - Output: JSON with strategy, segment count, store path

2. **read_document_segments()** (lines 1603-1720)
   - Input: `paper_dir`, `query_type`, `keywords`, `max_segments`, `max_total_chars`
   - Process: Rank segments by relevance, select intelligently
   - Output: Curated segments for specific query

3. **get_document_overview()** (lines 1723-1780)
   - Input: `paper_dir`
   - Output: Complete document metadata with all segments listed

**Segmentation Strategies**:

Based on content analysis, the system selects from:
- `semantic_research_focused`: For papers with algorithmic content
- `algorithm_preserve_integrity`: For algorithm-heavy documents
- `concept_implementation_hybrid`: For tutorial/guide documents
- `semantic_chunking_enhanced`: For long unstructured documents
- `content_aware_segmentation`: Default fallback strategy

---

## Integration with Workspace Tools

### Read Segmented Document

```go
type ReadDocumentSegmentsInput struct {
    SegmentIndexPath string   `json:"segment_index_path"`
    QueryType        string   `json:"query_type"` // "concept", "algorithm", "code_planning"
    MaxSegments      int      `json:"max_segments,omitempty"`
    MaxTotalChars    int      `json:"max_total_chars,omitempty"`
    KeywordFilter    []string `json:"keyword_filter,omitempty"`
}

type ReadDocumentSegmentsOutput struct {
    Success        bool
    SegmentsCount  int
    TotalChars     int
    Content        string
    SourceSegments []string // Which segments were combined
    Error          string
}
```

**Agent Behavior**:

```
Traditional approach:
├─ Agent: read_file("huge_paper.md", offset=0, limit=2000)
├─ Gets: Lines 1-2000 (might be just introduction)
├─ Problem: Needs multiple reads, might miss important parts
└─ Cost: High token usage for irrelevant content

Segmentation approach:
├─ Agent: read_document_segments(index_path, query_type="algorithm_extraction")
├─ Gets: Algorithm section + proof + implementation notes
├─ Benefit: All related content in one call, no noise
└─ Cost: 70% fewer tokens for same understanding
```

---

## Practical Guidance

### When Segmentation Helps Most

✅ **High benefit** when:
- File > 50KB (or > 5000 lines)
- File contains multiple logical sections
- Agent needs specific information (not everything)
- Same file will be queried multiple times

❌ **Low benefit** when:
- File < 5KB
- Single-purpose file (monolithic function)
- Agent needs the entire context
- One-off query

### Chunk Size Tuning

**Default: 8K-15K characters per chunk** (roughly 2K-4K tokens)

```
Too small chunks (<3K chars):
  ✗ Too many chunks per document
  ✗ Query system becomes noisy
  ✓ Exact retrieval
  ✗ Loses narrative flow

Optimal (8K-15K chars):
  ✓ Balanced between coherence and queryability
  ✓ Typically 1-3 chunks per semantic topic
  ✓ Good signal-to-noise in retrieval

Too large chunks (>30K chars):
  ✓ Fewer chunks to manage
  ✗ Less precise retrieval
  ✗ Agent might need only 3K chars but gets 30K
  ✗ Wastes tokens
```

### Handling Failing Cases

**Case: Mixed languages in file** (English + pseudocode + math formulas)
```
Solution: Segment with multi-modal awareness
├─ Recognize code blocks, preserve them
├─ Recognize math blocks, preserve them
├─ Segment around English prose
├─ Each chunk is coherent for its content type
```

**Case: Very deep nesting** (multiple levels of hierarchy)
```
Solution: Section path tracking
├─ Store "3.2.1.4" for deeply nested section
├─ Query system can use breadcrumb navigation
├─ Agent knows context: "this is in Theory → Subsection B → Algorithm 1"
```

**Case: References and dependencies**
```
Solution: Smart footnote handling
├─ Keep footnote references in place
├─ Include referenced content in same chunk
├─ Mark cross-references
├─ Agent knows where to find related parts
```

---

## Performance Characteristics

### Segmentation Cost

```
One-time cost per document:
├─ Small file (5KB): ~0.05 seconds, $0.001
├─ Medium file (50KB): ~0.5 seconds, $0.005
├─ Large file (500KB): ~5 seconds, $0.05
└─ Huge file (5MB): ~50 seconds, $0.50

Cost amortization:
├─ 1 query: Bad ROI
├─ 10 queries: Breaks even
├─ 100+ queries: Excellent ROI (1000x savings vs. re-reading)
```

### Query Cost

```
Traditional full read:
├─ Query: "What's the algorithm?"
├─ System: Reads entire file (500K chars = 125K tokens)
├─ Cost: ~$0.015-0.03 per query

With segmentation:
├─ Query: "What's the algorithm?"
├─ System: Retrieves algorithm segments (15K chars = 3.75K tokens)
├─ Cost: ~$0.0004-0.001 per query
└─ Savings: 30-75x reduction per query
```

---

## Next Steps

1. **[01-advanced-context-engineering.md](01-advanced-context-engineering.md)** - CodeRAG for repositories
2. **[04-memory-hierarchy.md](04-memory-hierarchy.md)** - Combine with memory management
3. **[07-implementation-roadmap.md](07-implementation-roadmap.md)** - Implementation timeline

---

## References

**DeepCode Implementation**: `/research/DeepCode/tools/document_segmentation_server.py` (1,938 lines)

| Component | Location | Purpose |
|-----------|----------|---------|
| **DocumentAnalyzer class** | lines 52-230 | Type detection + strategy selection |
| **analyze_document_type()** | lines 186-223 | Semantic document classification |
| **detect_segmentation_strategy()** | lines 245-265 | Chooses optimal segmentation approach |
| **DocumentSegmenter class** | lines 233-430 | Core segmentation logic |
| **_identify_algorithm_blocks()** | lines 661-730 | Algorithm preservation |
| **_identify_concept_groups()** | lines 733-785 | Concept grouping |
| **_identify_formula_chains()** | lines 788-854 | Formula sequence detection |
| **_merge_related_content_blocks()** | lines 857-930 | Content block fusion |
| **analyze_and_segment_document()** MCP tool | lines 1432-1596 | Main indexing endpoint |
| **read_document_segments()** MCP tool | lines 1603-1720 | Intelligent retrieval endpoint |
| **get_document_overview()** MCP tool | lines 1723-1780 | Document metadata endpoint |
| **DocumentSegment** dataclass | lines 35-51 | Segment metadata structure |
| **DocumentIndex** dataclass | lines 54-62 | Complete index structure |

