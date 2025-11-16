# Token Metrics Display Guide

## Quick Reference: What Each Metric Means

### Inline Session Summary (After Each Request)
```
Session: 20K actual | 26K cached (92%) | 2K response
```

| Metric | Meaning | Formula | What It Costs |
|--------|---------|---------|---------------|
| **actual** | New tokens you're paying for | `prompt_tokens + response_tokens` | 100% cost (full price) |
| **cached** | Tokens reused from prompt cache | Tokens served from Gemini cache | ~10% cost (90% savings!) |
| **response** | Output tokens generated | `response_tokens` | Subset of actual tokens |

**Example:** If you see `20K actual | 26K cached (92%)`, that means:
- You paid for ~20,000 new tokens
- But the API reused 26,000 tokens from cache
- That's 92% of the conversation leveraging cache (excellent!)
- You saved roughly 2,600 tokens worth of cost through caching

### Session Summary Detail View
Displayed at session exit with structured breakdown:

#### 💰 Cost Metrics (What Matters)
```
├─ Actual Tokens:   20 (new prompt + response)
├─ Cached Tokens:   26 (92.9% of processed)
├─ Saved Cost:      ~3 tokens (cache reuse)
└─ Total Proc:      46 (for API billing)
```

- **Actual Tokens**: Direct API cost (input + output only)
- **Cached Tokens**: Reused from cache (costs 10% of actual)
- **Saved Cost**: Rough estimate in token units
- **Total Proc**: Sum for transparency on billing metric

#### 🔧 Token Breakdown
```
├─ Prompt (input):        14
├─ Response (output):     6
├─ Thinking:              0
└─ Cached Reuse:          26
```

Component-level details showing where tokens were used.

#### 📈 Session Efficiency
```
├─ Requests:              2
├─ Avg/Request:           23
├─ Cache Hit Rate:        92.9% (excellent!)
└─ Duration:              5s
```

- **Cache Hit Rate %**: Percentage of conversation from cache
- **Avg/Request**: Average tokens per API call
- **Duration**: Total session time

## Understanding Cache Efficiency

### Cache Hit Rate Interpretation

| Rate | Assessment | Recommendation |
|------|-----------|-----------------|
| 0-20% | ❌ No caching | Use longer prompts or enable caching |
| 20-50% | ⚠️ Minimal | Good opportunity for caching strategy |
| 50-80% | ✅ Good | Cache is helping significantly |
| 80%+ | 🚀 Excellent | Prompt caching strategy working well |

### Cost Savings Example

**Traditional approach (no cache):**
- Request 1: 10,000 tokens = $0.10
- Request 2: 10,000 tokens = $0.10
- Total: 20,000 tokens = $0.20

**With prompt caching (90% savings on cached):**
- Request 1: 10,000 tokens = $0.10
- Request 2: 9,000 new + 1,000 cached ≈ $0.091
- Total: 20,000 tokens ≈ $0.191
- **Savings: $0.009 per similar request** (5-10% overall)

## Real Examples

### Simple Single Query
```
Session: 2K actual | 0K cached (0%) | 0K response
```
- No prior context, so no cache reuse
- Small request (2K tokens)
- Cache will help on follow-up questions

### Multi-Turn Conversation
```
Session: 28K actual | 26K cached (92%) | 2K response
```
- Excellent cache hit rate (92%)
- System prompt and previous context reused
- Minimal new input needed per request

### With Extended Thinking
```
💰 Cost Metrics
├─ Actual Tokens:   300
├─ Cached Tokens:   150 (33.3%)
├─ Saved Cost:      ~15
└─ Total Proc:      450

🔧 Token Breakdown
├─ Prompt:          200
├─ Response:        100
├─ Thinking:        150
└─ Cached:          150
```
- Thinking tokens (internal reasoning) use tokens too
- Even 33% cache hit rate saves significant cost
- Reusing cached thinking tokens across turns is powerful

## Tips for Optimizing Cache Efficiency

### 1. Use Longer System Prompts
Longer instructions = more cache benefits:
```go
// Good for caching - gets reused
systemPrompt := `You are an expert coding assistant...
[long detailed instructions]...`
```

### 2. Keep Context Between Requests
Multi-turn conversations leverage cache:
```
Request 1: Setup + Question 1 → Cache filled
Request 2: [Same context reused] + Question 2 → Cache hit!
```

### 3. Batch Related Requests
Group similar queries to reuse cache:
```
Same prompt cache file → 3 related coding questions
vs.
Different prompts → No cache reuse
```

### 4. Monitor Cache Hit Rate
- Aim for 50%+ cache hit rate
- If <20%, reconsider prompt structure
- 80%+ indicates optimal caching strategy

## Common Metrics Questions

### Q: Why does "actual" seem low with "cached" high?
A: The first request primes the cache. Subsequent requests are mostly cache reuse, so "actual" only counts truly new tokens while "cached" shows what was reused.

### Q: Is "Total Proc" what I'm billed for?
A: Yes, API billing is based on total tokens processed (actual + cached), but cached tokens cost ~10% of new tokens, so your actual cost is lower than total × unit_price.

### Q: What if cached > prompt?
A: This is normal! It means you're reusing a long cached context across multiple short requests. The cache is very effective.

### Q: How do I improve cache efficiency?
A: 
1. Use longer system prompts (more to cache)
2. Keep context between multi-turn requests
3. Reuse the same conversation across related questions

## Implementation Details

### Metric Calculation
```
Actual Tokens = prompt_tokens + response_tokens
Cached Tokens = cached_content_tokens
Cache Hit Rate % = (cached / (actual + cached)) × 100
Cost Savings (est) = cached / 10  (tokens equivalent)
```

### Display Format
- Numbers ≥1000 shown in compact form: `28K` instead of `28029`
- Percentages rounded to 1 decimal: `92.1%`
- Durations shown in seconds: `5s` or minutes: `2m30s`

## Metrics Across Different Scenarios

### Code Analysis/Editing
```
Session: 15K actual | 35K cached (70%) | 5K response
```
- High cache from reused codebase context
- Cached file contents reused across edits
- Excellent efficiency for code-heavy workflows

### Documentation/Writing
```
Session: 12K actual | 8K cached (40%) | 4K response
```
- Moderate cache from document structure
- New content added each turn
- Cache helps with formatting instructions

### Brainstorming/Exploration
```
Session: 8K actual | 2K cached (20%) | 2K response
```
- Low cache from varied requests
- Each question is unique
- Focus on efficient prompts rather than caching

## Conclusion

The new metrics display helps you:
- **See actual costs** (actual tokens only)
- **Understand cache benefits** (cached %)
- **Optimize strategy** (cache hit rate)
- **Monitor efficiency** (avg tokens/request)

Use these insights to make informed decisions about your prompt engineering and caching strategy!
