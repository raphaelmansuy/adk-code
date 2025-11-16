# Token Metrics Display - Final Reference

## What Each Metric Means (All values in TOKENS)

### Inline Session Summary (Per-Request Display)

```
Session: cost:21K | cached:20K (48% ⚠️ modest) | response:186
```

| Label | Meaning | Unit | Example |
|-------|---------|------|---------|
| **cost:X** | New tokens you paid for | tokens (K=thousands) | cost:21K = 21,000 new tokens |
| **cached:X** | Tokens reused from cache | tokens | cached:20K = 20,000 from cache |
| **(Z% indicator)** | Cache efficiency quality | percentage + emoji | 48% ⚠️ = modest efficiency |
| **response:X** | Output tokens generated | tokens | response:186 = 186 output tokens |

### Cache Efficiency Indicators

| Range | Indicator | Meaning | Action |
|-------|-----------|---------|--------|
| 80%+ | 🚀 excellent | Outstanding cache reuse | Keep using same prompts |
| 50-79% | ✅ good | Effective caching working | Great strategy |
| 20-49% | ⚠️ modest | Some cache benefit | Could optimize more |
| <20% | ❌ minimal | Little cache reuse | Change prompt strategy |

## Complete Examples

### Example 1: First Request (No Cache)
```
Session: cost:20K | response:50
```
- **cost:20K** = 20,000 new input tokens (first prompt)
- **response:50** = 50 output tokens from AI
- No cached metrics shown (none yet)

### Example 2: Second Request (Building Cache)
```
Session: cost:18K | cached:12K (40% ⚠️ modest) | response:120
```
- **cost:18K** = 18,000 new input tokens for this request
- **cached:12K** = 12,000 tokens reused from previous context
- **(40% ⚠️ modest)** = 40% of this request came from cache (room for improvement)
- **response:120** = 120 output tokens from this request

### Example 3: Multi-Turn with Strong Caching
```
Session: cost:16K | cached:28K (63% ✅ good) | response:245
```
- **cost:16K** = Only 16,000 new tokens needed
- **cached:28K** = Reused 28,000 from conversation history
- **(63% ✅ good)** = Good cache efficiency!
- **response:245** = 245 output tokens

### Example 4: Long Response Session
```
Session: cost:22K | cached:15K (40% ⚠️ modest) | response:1200
```
- **cost:22K** = 22,000 new tokens in the request
- **cached:15K** = 15,000 reused from cache
- **(40% ⚠️ modest)** = 40% efficiency
- **response:1200** = 1,200 output tokens (lengthy response)

## Session Summary Detail View

When session ends, you see:

```
💰 Cost Metrics (What You Pay)
  ├─ New Tokens:     160 (prompt + response you paid for)
  ├─ Cache Reuse:    50 tokens (23.8% efficiency)
  ├─ Cost Savings:   ~5 tokens via caching
  └─ API Billing:    210 total tokens

🔧 Token Breakdown
  ├─ Prompt (input):   100
  ├─ Response (output):60
  ├─ Thinking:         30
  └─ Cached Reuse:     50

📈 Session Efficiency
  ├─ Requests:         3
  ├─ Avg/Request:      70 tokens
  ├─ Cache Hit Rate:   23.8%
  └─ Duration:         2m30s
```

### What Each Field Means

**💰 Cost Metrics (What You Pay)**
- **New Tokens** = All brand new input+output you had to send/receive
- **Cache Reuse** = Tokens served from cache (10% cost of new)
- **Cost Savings** = Rough estimate in tokens (cached ÷ 10)
- **API Billing** = Total tokens API counts (for billing transparency)

**🔧 Token Breakdown**
- **Prompt (input)** = Tokens you sent as questions/instructions
- **Response (output)** = Tokens AI generated as responses
- **Thinking** = Tokens used for extended reasoning (if enabled)
- **Cached Reuse** = Tokens reused from prompt cache

**📈 Session Efficiency**
- **Requests** = How many API calls you made
- **Avg/Request** = Average tokens per API call
- **Cache Hit Rate** = % of conversation from cache
- **Duration** = Total session time

## Understanding the Numbers

### What Costs Money?
Only **New Tokens** cost full price:
- Your input tokens (prompt tokens)
- AI output tokens (response tokens)

### What's Cheap (10% Cost)?
**Cached Tokens** from prompt caching:
- System instructions (reused)
- Previous context (in multi-turn conversations)
- Large documents referenced multiple times

### Example Cost Calculation
**Session with 160 new tokens + 50 cached:**
- Actual API cost for ≈160 tokens
- Cached 50 tokens cost ≈5 tokens equivalent
- Total billing: 210 tokens
- Actual cost impact: ~165 "equivalent" tokens (18% savings!)

## Tips for Better Caching

### Aim for These Cache Efficiency Targets
- **First request**: 0% (nothing to cache yet)
- **Second request**: 20-30% (context being built)
- **Third+ requests**: 50%+ (strong caching)
- **Long conversations**: 70-90% (excellent caching!)

### How to Improve Cache Hit Rate
1. **Use longer system prompts** - More to reuse across requests
2. **Keep context consistent** - Ask related questions in same session
3. **Reference same documents** - Reuse large files/contexts
4. **Enable prompt caching** - Built-in to Gemini API

### Monitor This
- If cache hit rate **<20%**: Your prompts change too much
- If cache hit rate **20-50%**: Room for improvement
- If cache hit rate **50%+**: Good strategy working!
- If cache hit rate **80%+**: Excellent optimization

## Terminology Clarification

| What You See | What It Means | What It Counts |
|---|---|---|
| **cost** | Tokens you pay for | New input + output only |
| **cached** | Tokens from cache | Reused context (10% cost) |
| **response** | AI output tokens | Words/tokens the AI generated |
| **K** | Thousands | 21K = 21,000 tokens |

## All Values Are in TOKENS

Every number you see is **tokens**:
- `cost:21K` = 21,000 tokens
- `cached:20K` = 20,000 tokens  
- `response:186` = 186 tokens
- `Cache Hit Rate: 48%` = 48% of tokens were cached

Unit is always **tokens**, abbreviated as numbers.

---

**Remember:** Larger numbers = more tokens = higher cost, BUT cache reuse saves you 90% on those tokens!
