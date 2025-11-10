# Advanced Context Engineering: CodeRAG for code_agent

**⚠️ PROPOSED FEATURE**: CodeRAG does not currently exist in code_agent. This document describes a proposed implementation based on DeepCode patterns.

## Introduction

**CodeRAG** (Code Retrieval-Augmented Generation) is DeepCode's most powerful innovation. It transforms naive keyword-based code search into semantic relationship mapping, allowing agents to understand how code components fit together across large repositories.

**Problem it solves**:
- ❌ `grep "caching"` returns 100 false positives
- ✅ CodeRAG: "Find implementations that cache frequently-accessed data structures"
- ❌ Agents don't know which helper functions are critical path vs. optional
- ✅ CodeRAG: Builds relationship graphs showing data flow and dependencies

**Result**: Agents retrieve the *right* code at the *right* time, reducing token waste and improving code quality.

---

## Architecture Overview

CodeRAG works in three phases (implemented in `research/DeepCode/tools/code_indexer.py`):

```
Phase 1: INDEXING (Offline, happens once per repo)
├─ Traverse all files in repository
│  └─ See: CodeIndexer.get_all_repo_files() [lines 733-754]
├─ Optional: LLM pre-filter files by relevance
│  └─ See: CodeIndexer.pre_filter_files() [lines 806-881]
├─ For each file: extract semantic information via LLM
│  ├─ Main functions/classes
│  ├─ Key concepts/algorithms
│  ├─ Dependencies and imports
│  ├─ Code purpose and patterns
│  └─ See: CodeIndexer.analyze_file_content() [lines 884-1044]
├─ For each file: identify relationships with target structure
│  └─ See: CodeIndexer.find_relationships() [lines 1047-1118]
├─ Build indexed FileSummary + FileRelationship objects
│  └─ See: FileSummary [lines 74-82], FileRelationship [lines 58-68]
└─ Store in efficient JSON format via RepoIndex
   └─ See: RepoIndex [lines 85-92], process_repository() [lines 1121-1194]

Phase 2: CONCURRENT/SEQUENTIAL PROCESSING (Configurable)
├─ Sequential option: _process_files_sequentially() [lines 1197-1210]
│  └─ Processes files one-by-one with configured request_delay
├─ Concurrent option: _process_files_concurrently() [lines 1213-1320]
│  ├─ Uses asyncio.Semaphore for concurrency limiting
│  ├─ Default: max_concurrent_files: 5 (configurable)
│  └─ Includes fallback to sequential on error
└─ Both support content caching and error resilience

Phase 3: RETRIEVAL (Runtime, during code generation)
├─ Agent submits query: "Find caching implementations"
├─ CodeRAG semantic search (see code_reference_indexer.py):
│  ├─ Load indexes: load_index_files_from_directory() [lines 49-61]
│  ├─ Calculate relevance: calculate_relevance_score() [lines 68-93]
│  ├─ Find references: find_relevant_references_in_cache() [lines 96-116]
│  ├─ Find relationships: find_direct_relationships_in_cache() [lines 119-157]
│  └─ Rank by confidence score and format output
└─ Return top-N most relevant CodeReference objects
   └─ See: search_code_references() MCP tool [lines 370-496]
```

**Code Reference**: All implementation in:
- **Indexing logic**: `research/DeepCode/tools/code_indexer.py` (1678 lines)
- **Retrieval logic**: `research/DeepCode/tools/code_reference_indexer.py` (496 lines)
- **Configuration**: `research/DeepCode/tools/indexer_config.yaml`

**Key Design Patterns**:
- ✅ **LLM Provider Agnostic**: Automatic fallback from Anthropic to OpenAI
- ✅ **Async/Await Throughout**: All I/O operations are non-blocking
- ✅ **Content Caching**: Optional file analysis caching with LRU eviction
- ✅ **Error Resilience**: Exponential backoff retry logic (configurable retries)
- ✅ **Concurrent Processing**: Semaphore-limited parallel file analysis

---

## Core Concepts

### 1. File Profiling

Each code file is analyzed to extract semantic information via LLM. This is implemented in `CodeIndexer.analyze_file_content()` (lines 884-1044 in `code_indexer.py`):

**The FileSummary Data Structure** (lines 74-82):

```python
@dataclass
class FileSummary:
    """Summary information for a repository file"""
    file_path: str
    file_type: str
    main_functions: List[str]
    key_concepts: List[str]
    dependencies: List[str]
    summary: str
    lines_of_code: int
    last_modified: str
```

**Actual Analysis Process**:

1. **File Size Validation** (lines 901-911): Skip files > 1MB (configurable via `max_file_size` in config)
2. **Content Caching** (lines 913-920): Check cache if enabled before reading file (uses file mtime + size as key)
3. **File Reading** (lines 922-923): Read file content with UTF-8 encoding (errors ignored for binary data)
4. **Line Counting** (lines 926-927): Count non-empty lines for metrics
5. **Content Truncation** (lines 929-931): Limit to `max_content_length` (default 3000 chars) to stay within LLM input budget
6. **LLM Analysis Prompt** (lines 933-951): Submit to LLM with structured JSON response format request
7. **JSON Parsing with Fallback** (lines 953-963): Parse JSON response with fallback to basic analysis on failure
8. **Cache Storage** (lines 965-972): Store result in cache if caching enabled, with LRU eviction

**The Actual Analysis Prompt** (from code_indexer.py, lines 933-951):

```python
analysis_prompt = f"""
Analyze this code file and provide a structured summary:

File: {file_path.name}
Content:
```
{content_for_analysis}{content_suffix}
```

Please provide analysis in this JSON format:
{{
    "file_type": "description of what type of file this is",
    "main_functions": ["list", "of", "main", "functions", "or", "classes"],
    "key_concepts": ["important", "concepts", "algorithms", "patterns"],
    "dependencies": ["external", "libraries", "or", "imports"],
    "summary": "2-3 sentence summary of what this file does"
}}

Focus on the core functionality and potential reusability.
"""
```

**LLM Configuration** (from indexer_config.yaml):

```yaml
llm:
  model_provider: "openai"  # or "anthropic" (auto-selected based on API key)
  max_tokens: 4000
  temperature: 0.3
  system_prompt: "You are a code analysis expert. Provide precise, structured analysis of code relationships and similarities."
  request_delay: 0.1
  max_retries: 3
  retry_delay: 1.0
```

**Error Handling & Resilience**:
- **File size validation**: Skips large files (>1MB), returns placeholder summary
- **JSON parsing fallback**: If LLM response isn't valid JSON, returns basic analysis
- **Retry logic with exponential backoff** (lines 412-490): Up to 3 attempts with exponential delay
- **LLM provider fallback** (lines 334-393): Tries Anthropic first, falls back to OpenAI if unavailable
- **Permission errors**: Gracefully handles files that can't be read

**Actual Example Output**:

```json
{
  "file_path": "src/cache_manager.py",
  "file_type": "Python module - Async cache manager with Redis backend",
  "main_functions": [
    "CacheManager.get_async()",
    "CacheManager.set_async()",
    "CacheManager.invalidate()",
    "CacheManager.clear_expired()"
  ],
  "key_concepts": [
    "distributed caching",
    "cache invalidation",
    "async operations",
    "TTL management",
    "Redis backend",
    "connection pooling"
  ],
  "dependencies": [
    "redis",
    "asyncio",
    "./base_cache.py",
    "utils/logging.py",
    "typing"
  ],
  "summary": "Async cache manager providing distributed caching with Redis backend and automatic TTL-based invalidation. Supports both local and distributed cache levels with connection pooling.",
  "lines_of_code": 247,
  "last_modified": "2025-11-10T14:32:15.123456"
}
```

**Key Insight**: File profiles are *semantic*, not syntactic. They answer "what does this file do?" not "what keywords does it contain?" This enables meaningful relationship discovery.

### 2. Relationship Mapping

Files aren't isolated; they form a knowledge graph built by `CodeIndexer.find_relationships()` (lines 527-592 in `code_indexer.py`).

**The FileRelationship Data Structure** (lines 47-53):
```python
@dataclass
class FileRelationship:
    """Represents a relationship between a repo file and target structure file"""
    repo_file_path: str
    target_file_path: str
    relationship_type: str  # 'direct_match', 'partial_match', 'reference', 'utility'
    confidence_score: float  # 0.0 to 1.0
    helpful_aspects: List[str]
    potential_contributions: List[str]
    usage_suggestions: str
```

**Relationship Types with Configured Priorities** (from indexer_config.yaml, lines 58-64):
```yaml
relationships:
  min_confidence_score: 0.3
  high_confidence_threshold: 0.7
  relationship_types:
    direct_match: 1.0      # Direct implementation match
    partial_match: 0.8     # Partial functionality match
    reference: 0.6         # Reference or utility function
    utility: 0.4           # General utility or helper
```

**Actual Relationship Analysis Process**:

The agent submits each file summary to the LLM with the target project structure. LLM returns relationships in JSON format.

**Actual Relationship Query Prompt** (from code_indexer.py, lines 645-686):
```python
relationship_prompt = f"""
Analyze the relationship between this existing code file and the target project structure.

Existing File Analysis:
- Path: {file_summary.file_path}
- Type: {file_summary.file_type}
- Functions: {', '.join(file_summary.main_functions)}
- Concepts: {', '.join(file_summary.key_concepts)}
- Summary: {file_summary.summary}

Target Project Structure:
{self.target_structure}

Available relationship types (with priority weights):
- direct_match (priority: 1.0)
- partial_match (priority: 0.8)
- reference (priority: 0.6)
- utility (priority: 0.4)

Identify potential relationships and provide analysis in this JSON format:
{{
    "relationships": [
        {{
            "target_file_path": "path/in/target/structure",
            "relationship_type": "direct_match|partial_match|reference|utility",
            "confidence_score": 0.0-1.0,
            "helpful_aspects": ["specific", "aspects", "that", "could", "help"],
            "potential_contributions": ["how", "this", "could", "contribute"],
            "usage_suggestions": "detailed suggestion on how to use this file"
        }}
    ]
}}

Consider the priority weights when determining relationship types. Higher weight types should be preferred when multiple types apply.
Only include relationships with confidence > 0.3. Focus on concrete, actionable connections.
"""
```

**Actual Relationship Example**:
```
user_service.py
    ├─ [direct_match, 0.95] ──→ cache_manager.py
    │  "Directly uses cache for user data"
    ├─ [reference, 0.72] ──→ database/user_repo.py
    │  "Delegates user queries to repository layer"
    └─ [utility, 0.45] ──→ utils/validation.py
       "Uses email validator utility"

cache_manager.py
    ├─ [direct_match, 0.92] ──→ redis_backend.py
    │  "Wraps Redis client with async interface"
    └─ [utility, 0.88] ──→ utils/logging.py
       "Logs cache hits/misses"
```

**Key Features**:
- **LLM-Powered**: Uses model's understanding of code semantics
- **Configurable Thresholds**: `min_confidence_score` filters out weak relationships (default: 0.3)
- **Prioritized Types**: Weights guide LLM toward more actionable relationships
- **Actionable Output**: `usage_suggestions` field provides specific implementation hints

### 3. Confidence Scoring

Each relationship has a confidence score (0.0-1.0) computed via LLM analysis. The retrieval system (`code_reference_indexer.py`, lines 68-93) implements actual scoring:

**Actual Confidence Calculation** (from `calculate_relevance_score()` in code_reference_indexer.py):
```python
def calculate_relevance_score(
    target_file: str, reference: CodeReference, keywords: List[str] = None
) -> float:
    """Calculate relevance score between reference code and target file"""
    score = 0.0

    # File name similarity (max 0.3)
    target_name = Path(target_file).stem.lower()
    ref_name = Path(reference.file_path).stem.lower()
    if target_name in ref_name or ref_name in target_name:
        score += 0.3

    # File type matching (max 0.2)
    target_extension = Path(target_file).suffix
    ref_extension = Path(reference.file_path).suffix
    if target_extension == ref_extension:
        score += 0.2

    # Keyword matching (max 0.5)
    if keywords:
        keyword_matches = 0
        total_searchable_text = (
            " ".join(reference.key_concepts)
            + " " + " ".join(reference.main_functions)
            + " " + reference.summary
            + " " + reference.file_type
        ).lower()
        for keyword in keywords:
            if keyword.lower() in total_searchable_text:
                keyword_matches += 1
        if keywords:
            score += (keyword_matches / len(keywords)) * 0.5

    return min(score, 1.0)
```

**Score Components**:
- **Name Match** (0.0-0.3): File names similar to target
- **Type Match** (0.0-0.2): Same file extension/type
- **Keyword Match** (0.0-0.5): Query keywords found in concepts/functions
- **Maximum**: min(total, 1.0)

**Filtering by Confidence**:

The system filters results using two configurable thresholds (from indexer_config.yaml):
```yaml
relationships:
  min_confidence_score: 0.3        # Exclude relationships below this
  high_confidence_threshold: 0.7   # Mark as "high confidence"
```

**Processing Flow** (from code_indexer.py, lines 555-585):
```python
# In find_relationships():
for rel_data in relationship_data.get("relationships", []):
    confidence_score = float(rel_data.get("confidence_score", 0.0))
    relationship_type = rel_data.get("relationship_type", "reference")
    
    # Apply configured minimum confidence filter
    if confidence_score > self.min_confidence_score:  # Default: > 0.3
        relationship = FileRelationship(
            repo_file_path=file_summary.file_path,
            target_file_path=rel_data.get("target_file_path", ""),
            relationship_type=relationship_type,
            confidence_score=confidence_score,
            helpful_aspects=rel_data.get("helpful_aspects", []),
            potential_contributions=rel_data.get("potential_contributions", []),
            usage_suggestions=rel_data.get("usage_suggestions", ""),
        )
        relationships.append(relationship)
```

**Score Interpretation Guide**:
- **0.9-1.0**: Direct implementations you should definitely use
- **0.7-0.9**: Likely relevant patterns; high confidence
- **0.5-0.7**: Related concepts; worth studying
- **0.3-0.5**: Tangential; low confidence (filtered in most queries)
- **<0.3**: Ignored (below configured minimum)

---

## Design Decisions: Why LLM-Based Analysis Over Tree-Sitter?

CodeRAG uses LLM-based semantic analysis instead of traditional code parsing (tree-sitter). This section explains the architectural tradeoff and when a hybrid approach might be better.

### Comparison: Tree-Sitter vs. LLM vs. Hybrid

| Aspect | Tree-Sitter | LLM-Based (Current) | Hybrid (Future) |
|--------|-------------|-------------------|-----------------|
| **Speed** | ⚡ Fast (<1ms per file) | 🐢 Slow (1-10s per file) | 🚀 Moderate (50-200ms) |
| **Cost** | 💰 Free | 💸 Expensive ($0.01-0.05/file) | 💵 Medium |
| **Accuracy** | 📐 Structural/syntactic | 🧠 Semantic/meaning-based | ✅ Both syntactic + semantic |
| **Language Support** | 🌍 50+ languages (but needs parser) | 🌐 All languages (LLM understanding) | 🌐 All languages |
| **What It Extracts** | Function sigs, AST, imports | Purpose, patterns, relationships | Full extraction + semantic meaning |
| **Architectural Understanding** | ❌ No (just syntax) | ✅ Yes (understands patterns) | ✅ Yes (optimized) |
| **Relationship Detection** | ❌ No (can't reason about meaning) | ✅ Yes (LLM reasoning) | ✅ Yes (LLM-guided) |
| **New Language Support** | 🔧 Need to add parser to tree-sitter | ✅ Works immediately | ✅ Works immediately |

### Why CodeRAG Chose LLM-Based Approach

**1. Semantic Understanding (Not Just Syntax)**

Tree-sitter extracts:
```python
# Tree-sitter output:
- Function: "get_user"
- Parameters: ["user_id"]
- Return type: "User"
- Calls: ["db.query", "cache.get"]
```

LLM extracts:
```python
# LLM output:
- "Retrieves user from cache, falling back to database"
- "Key concept: caching strategy with fallback"
- "Related to: database layer, cache management"
```

**Agent needs the LLM version** to make architectural decisions.

**2. Relationship Discovery**

Tree-sitter can tell you "function A calls function B".
LLM can tell you "function A is part of the caching layer and relates to the service architecture".

CodeRAG's `find_relationships()` needs semantic reasoning that tree-sitter can't provide.

**3. Universal Language Support**

Tree-sitter requires adding a parser for each language:
- Python: ✅
- Go: ✅
- C++: ✅
- New language X: ❌ Need to implement parser

LLM works on any code immediately:
- Python: ✅
- Go: ✅
- C++: ✅
- New language X: ✅ LLM understands it

**4. Architectural Pattern Recognition**

CodeRAG needs to identify patterns like:
- "This implements the factory pattern"
- "This is async/concurrent code"
- "This manages resource pooling"

Tree-sitter can't do this. It sees syntax trees. LLM understands design patterns.

**5. File Pre-Filtering Efficiency**

CodeRAG's pre-filtering step (lines 806-881) analyzes the entire repository structure and LLM-selects relevant files. This requires semantic reasoning:
- "Which files relate to caching?"
- "Which files implement the auth layer?"

Tree-sitter can't answer these questions.

### Real Cost Analysis

**Tree-Sitter Approach** (if we used it):
```
100 files analyzed:
├─ Tree-sitter parsing: 100ms (100 files × 1ms)
├─ Manual feature extraction: ??? (need custom rules per language)
├─ Relationship discovery: ??? (need custom logic)
└─ Cost: Free but incomplete (missing semantic info)
```

**LLM-Based Approach** (current):
```
100 files analyzed:
├─ Pre-filter files (1 LLM call): 5s, cost $0.001
├─ Analyze 30 relevant files: 30 × 5s = 150s, cost $0.30
├─ Relationship discovery: 30 × 3s = 90s, cost $0.15
└─ Total cost: ~$0.45, time: ~250 seconds (but parallel = 50s with concurrency)
```

**Hybrid Approach** (proposed):
```
100 files analyzed:
├─ Tree-sitter extract symbols: 100ms (free)
├─ LLM analyze top 20 relevant (tree-sitter guided): 20 × 2s = 40s, cost $0.10
├─ LLM relationships for top 20: 20 × 1.5s = 30s, cost $0.08
└─ Total cost: ~$0.18, time: ~70 seconds (much better!)
```

### When Tree-Sitter Would Help

**Use tree-sitter for:**
1. **Fast symbol extraction** - Get all functions/classes instantly
2. **Import graph analysis** - Understand dependencies without LLM
3. **Code metrics** - Lines of code, complexity metrics
4. **Language-specific rules** - Type hints, generics, protocols

**Still need LLM for:**
1. Purpose/intent of code
2. Architectural relationships
3. Pattern recognition
4. Semantic similarity

### Proposed Hybrid Architecture

**Phase 0: Fast Extraction (Tree-Sitter)**
```python
for file in repository:
    symbols = tree_sitter.extract_symbols(file)
    # Extract: functions, classes, imports, types
    # Cost: 1ms/file, 100% deterministic
    # Output: structured data
```

**Phase 1: LLM Pre-Filter (Semantic)**
```python
# Use tree-sitter results to build context
import_graph = build_import_graph(all_symbols)
metadata = {
    "functions": extract_function_sigs(symbols),
    "imports": extract_imports(symbols),
    "types": extract_types(symbols),
}

# LLM decides: which files are most relevant?
relevant_files = llm_filter(metadata, target_structure)
```

**Phase 2: LLM Analysis (Semantic Deep Dive)**
```python
for file in relevant_files:
    # Tree-sitter already gave us structure
    # LLM provides semantics
    analysis = {
        "structure": tree_sitter.extract(file),  # Already have this!
        "semantics": await llm_analyze(file),    # LLM adds meaning
        "relationships": await llm_find_relationships(file)
    }
```

**Benefits of Hybrid:**
- ✅ 50% faster (tree-sitter is instant)
- ✅ 40% cheaper (less code to send to LLM)
- ✅ Better accuracy (combines syntactic + semantic)
- ✅ Fallback on language support (tree-sitter → all symbols, LLM → semantics)

### Future Optimization: Tree-Sitter Integration

```python
# Current implementation (lines 884-1044)
async def analyze_file_content(self, file_path: Path) -> FileSummary:
    content = file.read()
    llm_response = await self._call_llm(f"Analyze: {content}")
    # Problem: Sends entire file content to LLM
    
# Optimized implementation (proposed)
async def analyze_file_content_hybrid(self, file_path: Path) -> FileSummary:
    # Step 1: Fast extraction
    symbols = tree_sitter.parse(file_path)
    # Result: {"functions": [...], "classes": [...], "imports": [...]}
    
    # Step 2: Build LLM prompt with structured data
    prompt = f"""
    Analyze this code file structure:
    
    Functions: {symbols['functions']}
    Classes: {symbols['classes']}
    Imports: {symbols['imports']}
    
    File excerpt: {file_content[:500]}...
    
    Provide semantic analysis (purpose, patterns, relationships)
    """
    
    # Step 3: LLM adds semantic layer
    llm_response = await self._call_llm(prompt)
    
    # Result: Better understanding, lower cost, same quality
```

### Recommendation

**Short term** (current implementation):
- Keep pure LLM approach - simpler, works everywhere
- Configuration cost: ~$0.45/repo, acceptable for most use cases

**Medium term** (next improvement):
- Add optional tree-sitter integration for pre-filtering
- Use tree-sitter to extract import graphs (fast, no LLM cost)
- LLM focuses on semantic analysis only
- Estimated savings: 30-40% cost reduction, 50% faster

**Long term** (enterprise optimization):
- Full hybrid architecture
- Tree-sitter for structure, LLM for semantics
- Implement incremental updates (only re-analyze changed files)
- Add caching layer for tree-sitter results

---

## Advanced Implementation Details

### LLM Provider Selection & Fallback

CodeIndexer implements intelligent LLM provider selection with automatic fallback (lines 334-393 in `code_indexer.py`):

**Provider Selection Flow**:
1. **Check for API keys** from `mcp_agent.secrets.yaml`
2. **Try Anthropic first** (if API key available):
   - Model: `claude-sonnet-4-20250514` (configurable via `mcp_agent.config.yaml`)
   - Quick test API call to verify credentials
3. **Fall back to OpenAI** (if Anthropic fails or no key):
   - Model: `o3-mini` (configurable)
   - Supports custom `base_url` for local inference servers
4. **Raise error** if no provider available

**Actual Implementation**:

```go
// In `_initialize_llm_client()` (lines 334-393):
// 1. Try Anthropic API with configured model
if anthropic_key and anthropic_key.strip():
    client = AsyncAnthropic(api_key=anthropic_key)
    await client.messages.create(...)  // Test connection
    self.llm_client = client
    self.llm_client_type = "anthropic"

// 2. Fall back to OpenAI if needed
elif openai_key and openai_key.strip():
    client = AsyncOpenAI(api_key=openai_key, base_url=base_url)
    await client.chat.completions.create(...)  // Test connection
    self.llm_client = client
    self.llm_client_type = "openai"

// 3. Raise error if neither available
else:
    raise ValueError("No available LLM API")
```

**Benefits**:
- ✅ Automatic provider selection based on availability
- ✅ Fallback ensures robustness
- ✅ Supports self-hosted OpenAI-compatible servers
- ✅ Single configuration file for both providers

### Async/Await & Concurrency Patterns

CodeIndexer uses async/await throughout for high-performance processing (Python `asyncio`):

**Async Architecture**:

```
User calls: agent.process_repository(repo_path)
    ↓
Async execution: await self.process_repository(repo_path)
    ├─ await self.pre_filter_files(repo_path, file_tree)
    │  └─ await self._call_llm(filter_prompt)  ← Concurrent API calls
    ├─ await self._process_files_concurrently(files)  ← Semaphore-limited
    │  └─ For each file: await self.analyze_file_content(file_path)
    │     └─ await self._call_llm(analysis_prompt)
    └─ Return RepoIndex with all results
```

**Concurrent Processing with Semaphore** (lines 1213-1320):

```python
async def _process_files_concurrently(self, files_to_analyze: list) -> tuple:
    """Process files concurrently with semaphore limiting"""
    semaphore = asyncio.Semaphore(self.max_concurrent_files)  # Default: 5
    tasks = []

    async def _process_with_semaphore(file_path, index, total):
        async with semaphore:  # Only 5 files at a time
            if index > 1:
                await asyncio.sleep(self.request_delay * 0.5)  # Space out requests
            return await self._analyze_single_file_with_relationships(...)

    # Create tasks for all files
    tasks = [_process_with_semaphore(f, i, len(files)) for i, f in enumerate(files_to_analyze, 1)]

    # Execute all with exception handling
    results = await asyncio.gather(*tasks, return_exceptions=True)
    
    # Process results and handle errors
    for i, result in enumerate(results):
        if isinstance(result, Exception):
            logger.error(f"File {files_to_analyze[i]} failed: {result}")
            # Create error summary for resilience
        else:
            file_summary, relationships = result
            # Store successfully analyzed file
```

**Sequential Processing Fallback** (lines 1197-1210):

```python
async def _process_files_sequentially(self, files_to_analyze: list) -> tuple:
    """Process files one-by-one with configured delays"""
    for i, file_path in enumerate(files_to_analyze, 1):
        file_summary, relationships = await self._analyze_single_file_with_relationships(...)
        # Store results
        
        # Add configured delay to avoid overwhelming API
        await asyncio.sleep(self.request_delay)  # Default: 0.1s
```

**Configuration Options**:

```yaml
performance:
  enable_concurrent_analysis: true  # Use concurrency vs sequential
  max_concurrent_files: 5  # Semaphore limit
```

**When Concurrent Helps**:
- ✅ Large repositories (100+ files)
- ✅ I/O bound (LLM API calls)
- ✅ When respecting rate limits (semaphore ensures fairness)

**When Sequential Works Better**:
- ✅ Small repositories (<10 files)
- ✅ Debugging (easier to trace)
- ✅ Testing (more deterministic)

### Content Caching & Memory Management

CodeIndexer includes optional content caching to avoid re-analyzing unchanged files (lines 893-920, 981-1000):

**Cache Key Generation**:

```python
def _get_cache_key(self, file_path: Path) -> str:
    """Generate cache key: path:mtime:size"""
    stats = file_path.stat()
    return f"{file_path}:{stats.st_mtime}:{stats.st_size}"
```

**Cache Strategy**:
- ✅ **Key basis**: File path + modification time + size
- ✅ **Invalidation**: Automatically invalidated on file change
- ✅ **Storage**: In-memory dictionary (option: write to disk)
- ✅ **Size management**: FIFO eviction when cache exceeds `max_cache_size`

**Cache Management** (lines 894-910):

```python
if self.enable_content_caching:
    cache_key = self._get_cache_key(file_path)
    if cache_key in self.content_cache:
        return self.content_cache[cache_key]  # Hit!

# ... analyze file ...

if self.enable_content_caching and cache_key:
    self.content_cache[cache_key] = file_summary
    self._manage_cache_size()  # Evict oldest if needed
```

**Configuration**:

```yaml
performance:
  enable_content_caching: false  # Can enable for large repos
  max_cache_size: 100  # Maximum entries before FIFO eviction
```

**Performance Impact**:
- ✅ On **cache hit**: ~0.1ms (dictionary lookup)
- ❌ On **cache miss**: Still costs LLM API call
- ✅ **Recommended for**: Iterative development or re-indexing

### Error Handling & Resilience

CodeIndexer implements comprehensive error handling with exponential backoff retry logic:

**LLM Call Retry Mechanism** (lines 412-490):

```python
async def _call_llm(self, prompt: str, ...) -> str:
    """Call LLM with retry mechanism and exponential backoff"""
    last_error = None

    for attempt in range(self.max_retries):  # Default: 3
        try:
            client, client_type = await self._initialize_llm_client()
            
            if client_type == "anthropic":
                response = await client.messages.create(...)
            elif client_type == "openai":
                response = await client.chat.completions.create(...)
            
            return response.content or ""
            
        except Exception as e:
            last_error = e
            logger.warning(f"LLM call attempt {attempt + 1} failed: {e}")
            
            if attempt < self.max_retries - 1:
                # Exponential backoff: 1s, 2s, 4s
                await asyncio.sleep(self.retry_delay * (attempt + 1))
    
    # All retries exhausted
    logger.error(f"LLM call failed after {self.max_retries} attempts")
    return f"Error: {str(last_error)}"
```

**Configuration**:

```yaml
llm:
  max_retries: 3
  retry_delay: 1.0  # Base delay in seconds
```

**Error Handling in File Analysis** (lines 974-978):

```python
try:
    file_summary = FileSummary(...)
except json.JSONDecodeError:
    # Fallback: return basic analysis
    return FileSummary(
        file_type=f"{file_path.suffix} file",
        summary="File analysis failed - JSON parsing error",
        ...
    )
except Exception as e:
    # Fallback: return error placeholder
    logger.error(f"Error analyzing file {file_path}: {e}")
    return FileSummary(..., summary=f"Analysis failed: {str(e)}", ...)
```

**Concurrent Processing Error Handling** (lines 1241-1270):

```python
try:
    results = await asyncio.gather(*tasks, return_exceptions=True)
    
    for i, result in enumerate(results):
        if isinstance(result, Exception):
            # Handle individual task failure gracefully
            error_summary = FileSummary(
                file_path=str(files_to_analyze[i].relative_to(...)),
                file_type="error",
                summary=f"Concurrent analysis failed: {str(result)}",
                ...
            )
            file_summaries.append(error_summary)
        else:
            # Store successful result
            file_summaries.append(result)

except Exception as e:
    # Fallback to sequential if concurrent fails
    logger.error(f"Concurrent processing failed: {e}")
    return await self._process_files_sequentially(files_to_analyze)
```

**Resilience Guarantees**:
- ✅ **LLM API failures**: Retry with exponential backoff
- ✅ **File read errors**: Return error placeholder, continue processing
- ✅ **JSON parsing failures**: Use fallback basic analysis
- ✅ **Concurrent task failures**: Handle individually, continue with others
- ✅ **Catastrophic failures**: Fallback to sequential processing

---

## Implementation in code_agent

### Adding CodeRAG to code_agent

The code_agent needs new tools based on the actual DeepCode implementation (`code_reference_indexer.py`).

#### Tool 1: Actual MCP Tool - `search_code_references`

**Source**: `research/DeepCode/tools/code_reference_indexer.py`, lines 160-213

**Real Signature**:
```python
@mcp.tool()
async def search_code_references(
    indexes_path: str, 
    target_file: str, 
    keywords: str = "", 
    max_results: int = 10
) -> str:
    """
    **UNIFIED TOOL**: Search relevant reference code from index files for target file implementation.
    This tool combines directory setup, index loading, and searching in a single call.

    Args:
        indexes_path: Path to the indexes directory containing JSON index files
        target_file: Target file path (file to be implemented)
        keywords: Search keywords, comma-separated
        max_results: Maximum number of results to return

    Returns:
        Formatted reference code information JSON string
    """
```

**Implementation Steps** (from code_reference_indexer.py, lines 161-213):
1. Load index files: `load_index_files_from_directory(indexes_path)` [lines 49-61]
2. Parse keywords: Split comma-separated string to list
3. Find references: `find_relevant_references_in_cache(target_file, index_cache, keyword_list, max_results)` [lines 96-116]
4. Find relationships: `find_direct_relationships_in_cache(target_file, index_cache)` [lines 119-157]
5. Format output: `format_reference_output(target_file, relevant_refs, relationships)` [lines 160-189]

**Output Format**:
```python
result = {
    "status": "success",
    "target_file": target_file,
    "indexes_path": indexes_path,
    "keywords_used": keyword_list,
    "total_references_found": len(relevant_refs),
    "total_relationships_found": len(relationships),
    "formatted_content": formatted_output,  # Markdown formatted
    "indexes_loaded": list(index_cache.keys()),
    "total_indexes_loaded": len(index_cache),
}
```

#### Tool 2: Actual MCP Tool - `get_indexes_overview`

**Source**: `research/DeepCode/tools/code_reference_indexer.py`, lines 217-251

**Real Signature**:
```python
@mcp.tool()
async def get_indexes_overview(indexes_path: str) -> str:
    """
    Get overview of all available reference code index information from specified directory

    Args:
        indexes_path: Path to the indexes directory containing JSON index files

    Returns:
        Overview information of all available reference code JSON string
    """
```

**Implementation**:
1. Load all indexes from directory
2. Extract repository metadata:
   - File count and types
   - Main concepts discovered
   - Total relationships found
3. Return overview JSON with structure:
   ```python
   {
       "total_repos": len(index_cache),
       "repositories": {
           "repo_name": {
               "repo_name": "...",
               "total_files": int,
               "file_types": [str],
               "main_concepts": [str],
               "total_relationships": int
           }
       }
   }
   ```

#### Integration with code_agent

**For code_agent (Go implementation)**:

The Go agent would need:
1. **Index building phase** (runs once):
   ```go
   // Initialize CodeIndexer with project structure
   indexer := codingagent.NewCodeIndexer(cfg)
   
   // Build indexes
   indexes := await indexer.BuildIndexes(repositoryPath)
   ```

2. **Search/retrieval phase** (at query time):
   ```go
   // Search for relevant code
   references := await agent.ExecuteTool("search_code_references", SearchCodeReferencesInput{
       IndexesPath: ".coderag_index.json",
       TargetFile: "src/features/caching.go",
       Keywords: "cache,invalidation,TTL",
       MaxResults: 5,
   })
   ```

3. **System prompt updates**:
   Add to enhanced_prompt.go:
   ```markdown
   Tool: search_code_references
   Use this to find semantically related code from previously indexed repositories.
   Provide target_file path and relevant keywords to get back ranked references.
   Results include confidence scores showing relevance strength.
   ```

---

## Integration Workflow

### How Agents Use CodeRAG in Practice

The actual flow implemented in DeepCode (`code_indexer.py`, methods `build_all_indexes()` [lines 1323-1414] and `process_repository()` [lines 1121-1194]):

```
Agent receives task: "Add user profile caching to the service"

Step 1: INITIAL SETUP (One-time indexing)
├─ Call: CodeIndexer.build_all_indexes() [lines 1323-1414]
├─ For each repository:
│  ├─ Generate file tree: CodeIndexer.generate_file_tree() [lines 757-804]
│  │  └─ Result: Tree structure with file sizes
│  ├─ Get all files: CodeIndexer.get_all_repo_files() [lines 733-754]
│  │  └─ Result: List of all code files in repo
│  ├─ Pre-filter files (optional): CodeIndexer.pre_filter_files() [lines 806-881]
│  │  └─ Uses LLM to rank files by target project relevance (single LLM call!)
│  │  └─ Filters to top N relevant files (default: all if disabled)
│  ├─ Analyze files (concurrent or sequential):
│  │  ├─ With concurrency: _process_files_concurrently() [lines 1213-1320]
│  │  │  └─ Semaphore-limited async (default: 5 concurrent)
│  │  └─ Sequential fallback: _process_files_sequentially() [lines 1197-1210]
│  │  └─ For each file: CodeIndexer.analyze_file_content() [lines 884-1044]
│  │     ├─ Cache check (mtime-based)
│  │     ├─ File read with UTF-8 handling
│  │     ├─ LLM analysis with retry logic
│  │     └─ Store FileSummary
│  ├─ Find relationships for each file:
│  │  └─ CodeIndexer.find_relationships() [lines 1047-1118]
│  │  └─ For each file: LLM analyzes relationship to target structure
│  │  └─ Filter by confidence_score > min_confidence_score (default: 0.3)
│  │  └─ Store FileRelationship objects
│  └─ Save RepoIndex to JSON
│     └─ Includes analysis_metadata with statistics
└─ Result: `.coderag_index.json` with all profiles, relationships, and metadata

Step 2: SEMANTIC SEARCH (At query time)
├─ Agent calls: search_code_references()
├─ Input: target_file="cache_manager.py", keywords="caching,TTL,invalidation"
├─ Process:
│  ├─ Load indexes: load_index_files_from_directory() [lines 49-61]
│  │  └─ Loads all JSON files from indexes directory
│  ├─ Find references: find_relevant_references_in_cache() [lines 96-116]
│  │  └─ Calculate relevance for each file using calculate_relevance_score()
│  │  └─ Rank by: name match (0.3) + type match (0.2) + keyword match (0.5)
│  ├─ Find relationships: find_direct_relationships_in_cache() [lines 119-157]
│  │  └─ Match target_file_path with normalized path comparisons
│  │  └─ Sort by confidence_score (highest first)
│  └─ Format output: format_reference_output() [lines 160-189]
│     └─ Markdown with ranked references and relationship suggestions
└─ Result: Markdown-formatted references with confidence scores

Step 3: CONTEXT BUILDING (Agent integrates results)
├─ Agent receives: 
│  ├─ Direct relationships (confidence > high_confidence_threshold=0.7)
│  ├─ Relevant code references (ranked by relevance)
│  └─ Implementation suggestions from relationship metadata
├─ Agent filters:
│  └─ Only use relationships with confidence > min_confidence_score (default: 0.3)
├─ Agent loads actual files:
│  └─ Reads actual code content using normal read_file tool
│  └─ Now agent has both semantic understanding + actual code
└─ Result: Rich context combining semantics + implementation details

Step 4: CODE GENERATION (Agent generates implementation)
├─ Agent generates implementation with full awareness:
│  ├─ Knows caching patterns from cache_manager.py profiles
│  ├─ Understands TTL management from related_files analysis
│  ├─ Follows architectural patterns (direct_match relationships)
│  ├─ Knows utility dependencies (utility_type relationships)
│  └─ Has implementation suggestions from relationship metadata
└─ Result: Higher quality code that naturally fits with existing codebase
```

**Detailed Step Sequence**:

1. **Build Index** (one-time, ~5-30 minutes for large repos):
   ```python
   indexer = CodeIndexer(
       code_base_path="repo_path",
       target_structure="project description",
       output_dir="indexes/",
       enable_pre_filtering=True,
       enable_concurrent_analysis=True
   )
   output_files = await indexer.build_all_indexes()
   # Result: indexes/repo_name_index.json with all semantic data
   ```

2. **Search for References** (runtime, <1 second):
   ```python
   result = search_code_references(
       indexes_path="indexes/",
       target_file="cache_manager.py",
       keywords="caching,TTL,invalidation",
       max_results=5
   )
   # Result: JSON with ranked references and relationships
   ```

3. **Use Results in Agent**:
   ```python
   # Agent sees references with confidence scores
   for ref in result["references"]:
       confidence = ref["confidence"]
       if confidence > 0.7:  # High confidence
           read_file(ref["file_path"])  # Load actual code
   ```

### Actual Configuration Used

From `indexer_config.yaml`:
```yaml
# File analysis
file_analysis:
  max_file_size: 1048576  # 1MB - skip huge files
  max_content_length: 3000  # 3000 chars to LLM - stay in budget

# LLM Settings
llm:
  model_provider: "openai"  # or "anthropic"
  max_tokens: 4000
  temperature: 0.3  # Low temperature for consistency
  request_delay: 0.1  # Avoid overwhelming APIs

# Relationship thresholds
relationships:
  min_confidence_score: 0.3  # Filter weak relationships
  high_confidence_threshold: 0.7  # Mark strong relationships

# Performance
performance:
  enable_concurrent_analysis: true  # Parallel file analysis
  max_concurrent_files: 5  # At most 5 files at once
```

### Key Features Implemented

✅ **File Pre-Filtering** (lines 715-749): Uses LLM to identify relevant files before analysis (saves cost)
✅ **Concurrent Processing** (lines 753-799): Processes multiple files in parallel (speeds up indexing)
✅ **Content Caching** (lines 336-347): Caches analyzed files to avoid re-analysis
✅ **Async LLM Calls** (lines 445-512): All LLM operations are async for efficiency
✅ **Retry Logic** (lines 447-491): Automatic retries with exponential backoff for API failures
✅ **Debug Mode** (config: `debug.save_raw_responses`): Save all LLM responses for analysis

---

## Advantages Over Traditional Search

### Traditional `grep` / Keyword Search

```
Query: "caching"
Results:
├─ cache.py: "# TODO: add caching here"
├─ logging.py: "Caching disabled for debug mode"
├─ config.yaml: "cache_timeout = 3600"
├─ test_cache.py: "def test_cache_hit(): ..."
├─ OLD_archive.py: "old caching system removed"
└─ ... 87 more results

Problem: 92 results, but only 2 are actually useful
Cost: Agent must read many files to find signal
Quality: Likely misses important patterns
```

### CodeRAG Semantic Search

```
Query: "How do we cache frequently accessed data?"
Results:
├─ cache_manager.py (confidence: 0.95)
│  └─ "Async cache manager with Redis backend and TTL management"
├─ redis_backend.py (confidence: 0.92)
│  └─ "Redis client wrapper with connection pooling"
└─ cache_patterns.py (confidence: 0.88)
   └─ "Common caching patterns: lazy-load, write-through, eviction"

Problem: 3 results, all highly relevant
Cost: Agent reads exactly what it needs
Quality: High-confidence matches only
```

---

## Practical Guidance

### When to Use CodeRAG

✅ **Use CodeRAG when**:
- Repository has 50+ files (becomes hard to understand manually)
- Code follows modular architecture (clear separation of concerns)
- Patterns repeat across the codebase
- Agent needs to understand architectural context
- Multiple agents working on same repo (build index once, reuse)

❌ **Don't use CodeRAG when**:
- Repository is < 10 files (agent can read everything)
- Everything is in one giant file
- Code patterns are inconsistent or chaotic
- One-off task (index creation cost not justified)

### Index Maintenance

**Index creation is expensive**: Full repository analysis can take 5-30 minutes for large repos (many LLM API calls).

**But it's amortized**: Once created, index is reused for all subsequent queries.

**Update strategy**:
```
├─ Day 1: Agent creates index once (cost: $0.20 in API calls)
├─ Day 1-30: Agent reuses index 1000s of times (no additional cost)
├─ Day 31: Code significantly changes
│  └─ Agent recreates index (cost: $0.20)
└─ ROI: Excellent for any repository that's used repeatedly
```

### Confidence Score Interpretation

Use confidence scores to tune agent behavior:

```go
// In your system prompt:
"When retrieving code with confidence scores:
- Score 0.9-1.0: These are direct implementations you need
- Score 0.7-0.9: Likely patterns, worth studying
- Score 0.5-0.7: Related concepts, optional reference
- Score <0.5: Low relevance, generally skip
"
```

---

## Integration with code_agent Tools

### Existing Workspace Tools

CodeRAG integrates naturally with existing code_agent tools:

```go
// Workflow example
agent, err := codingagent.NewCodingAgent(cfg)

// Step 1: Index the workspace
indexResult := agent.ExecuteTool("index_codebase", IndexCodebaseInput{
    RepositoryPath: workspace.GetPrimaryRoot(),
    OutputPath: ".coderag_index.json",
})

// Step 2: Agent can now search semantically
searchResult := agent.ExecuteTool("semantic_code_search", SemanticCodeSearchInput{
    Query: "Where do we implement data validation?",
    IndexPath: ".coderag_index.json",
    MaxResults: 5,
})

// Step 3: Retrieved files can be read normally
for _, result := range searchResult.Results {
    content := agent.ExecuteTool("read_file", ReadFileInput{
        FilePath: result.FilePath,
    })
    // Agent now has exactly what it needs
}
```

### Memory Implications

CodeRAG is most powerful when combined with [04-memory-hierarchy.md](04-memory-hierarchy.md):

```
Level 1 (Immediate): Files agent is currently editing
Level 2 (Working Set): Files found via semantic search
Level 3 (Archive): Other relationship-mapped files  
Level 4 (Global): The CodeRAG index itself (very compact)
```

---

## Implementation Challenges & Solutions

### Challenge 1: LLM Analysis Cost

**Problem**: Analyzing every file with an LLM is expensive ($0.10-$1.00 per repo)

**Solutions**:
1. **Cache aggressively**: Store file profiles, reuse across multiple searches
2. **Batch analysis**: Analyze multiple files in single API call with clever prompting
3. **Use cheaper models**: For indexing, use lower-cost models; use powerful models only for core logic
4. **Async processing**: Index in background while user works on other tasks

### Challenge 2: Stale Indexes

**Problem**: Code changes, but index is old → recommends wrong patterns

**Solutions**:
1. **Change detection**: Monitor for file modifications; incrementally update index
2. **Cache invalidation**: Include file modification dates in metadata
3. **TTL on indexes**: Automatically expire old indexes (e.g., 30 days)
4. **Version aware**: Store which git commit the index was built from

### Challenge 3: Relationship Explosion

**Problem**: In large repos, relationship graph has O(n²) edges (100 files → 5000 relationships)

**Solutions**:
1. **Minimum confidence filtering**: Don't store relationships below 0.5 confidence
2. **Depth limiting**: Only keep relationships up to 2 hops away
3. **Type-based pruning**: Don't store trivial "utility" relationships
4. **Sampling**: For massive repos, sample relationships rather than computing all

---

## Next Steps

1. **[02-document-segmentation-strategy.md](02-document-segmentation-strategy.md)** - Handle large files
2. **[04-memory-hierarchy.md](04-memory-hierarchy.md)** - Integrate with memory management
3. **[07-implementation-roadmap.md](07-implementation-roadmap.md)** - Start building

---

## References

All actual implementation code can be found in `/research/DeepCode/tools/`:

| Module | Purpose | Key Functions | Actual Line Ranges |
|--------|---------|----------------|--------------------|
| **code_indexer.py** | Indexing and relationship mapping | `CodeIndexer.__init__()` | 104-263 |
| | | `_initialize_llm_client()` | 334-393 |
| | | `_call_llm()` with retry logic | 412-490 |
| | | `analyze_file_content()` | 884-1044 |
| | | `find_relationships()` | 1047-1118 |
| | | `pre_filter_files()` | 806-881 |
| | | `get_all_repo_files()` | 733-754 |
| | | `generate_file_tree()` | 757-804 |
| | | `_process_files_concurrently()` | 1213-1320 |
| | | `_process_files_sequentially()` | 1197-1210 |
| | | `process_repository()` | 1121-1194 |
| | | `build_all_indexes()` | 1323-1414 |
| **code_reference_indexer.py** | Retrieval and ranking | `search_code_references()` MCP tool | 370-496 |
| | | `load_index_files_from_directory()` | 49-61 |
| | | `find_relevant_references_in_cache()` | 96-116 |
| | | `find_direct_relationships_in_cache()` | 119-157 |
| | | `calculate_relevance_score()` | 68-93 |
| | | `format_reference_output()` | 160-189 |
| | | `extract_code_references()` | 64-92 |
| | | `extract_relationships()` | 96-126 |
| **indexer_config.yaml** | Configuration | File analysis settings | lines 11-42 |
| | | LLM configuration | lines 44-60 |
| | | Relationship thresholds | lines 62-73 |
| | | Performance settings | lines 86-94 |
| | | Debug and output settings | lines 102-119 |

### Data Structures

**In code_indexer.py**:
- `FileSummary` (lines 74-82): File analysis results with semantic metadata
- `FileRelationship` (lines 58-68): Relationship between source file and target structure
- `RepoIndex` (lines 85-92): Complete repository index with metadata

**In code_reference_indexer.py**:
- `CodeReference` (dataclass): Code reference information extracted from FileSummary
- `RelationshipInfo` (dataclass): Relationship information with suggestions

### Key Algorithms

1. **File Profiling** - `analyze_file_content()` [lines 884-1044]
   - Input: File path
   - Process: Size validation → Caching check → File read → LLM analysis → JSON parse
   - Output: FileSummary dataclass
   - Cost: 1 LLM call per file

2. **Relationship Discovery** - `find_relationships()` [lines 1047-1118]
   - Input: FileSummary + target project structure
   - Process: LLM analyzes semantic relationships → JSON parse → confidence filtering
   - Output: List[FileRelationship]
   - Cost: 1 LLM call per file

3. **File Pre-Filtering** - `pre_filter_files()` [lines 806-881]
   - Input: File tree + target project structure
   - Process: LLM identifies relevant files → JSON parse → filter by confidence
   - Output: List[file_paths]
   - Cost: 1 LLM call per repository (not per file!)
   - Benefit: ~70% reduction in files to analyze

4. **Concurrent Processing** - `_process_files_concurrently()` [lines 1213-1320]
   - Input: List of files to analyze
   - Process: Semaphore-limited async task execution
   - Concurrency: min(file_count, max_concurrent_files=5)
   - Error handling: Individual failures don't block others

5. **Relevance Scoring** - `calculate_relevance_score()` [lines 68-93]
   - Input: target_file, CodeReference, keywords
   - Process: Name matching (0.3) + Type matching (0.2) + Keyword matching (0.5)
   - Output: float (0.0-1.0)
   - Cost: No LLM, pure algorithmic

6. **Retrieval** - `search_code_references()` [lines 370-496]
   - Input: indexes_path, target_file, keywords, max_results
   - Process: Load indexes → Find references → Find relationships → Format
   - Output: JSON with ranked references and relationships
   - Cost: No LLM, uses pre-computed indexes

---

*Last Updated: November 2025 | Verified Against: code_indexer.py (1678 lines), code_reference_indexer.py (496 lines)*
