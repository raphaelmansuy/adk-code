# Advanced Prompt Engineering: Lessons from DeepCode

## Introduction

DeepCode's success comes not just from architecture, but from sophisticated system prompts that guide agents without over-constraining them.

**Key Insight**: Better prompts > Larger models

---

## Principles from DeepCode

### 1. Clarity Over Length

**Bad Prompt** (vague, unclear):
```
"You are an AI assistant. Help the user with code generation.
Make good code that works. Consider the requirements and make
sure it integrates well."
```

**Good Prompt** (clear, specific):
```
"You are the Code Generation Agent. Your single responsibility
is to synthesize code implementations.

INPUTS: You receive (1) Implementation plan from Code Planning Agent,
(2) Code references from Reference Mining Agent.

YOUR ROLE:
1. Generate code following the plan exactly
2. Integrate with existing code naturally
3. Follow project coding style
4. Add error handling and inline documentation

YOU MUST:
- Follow implementation plan precisely
- Match existing code style (spacing, naming, patterns)
- Include proper error handling
- Add TODO comments for complex sections
- Return code ready to integrate

YOU MUST NOT:
- Deviate from plan without clear justification
- Generate unnecessary abstractions
- Ignore existing patterns
- Skip error handling for edge cases
"
```

### 2. Responsibility Clarity

**Technique**: Explicitly state what agent does and doesn't do

```
"You are [Role]. Your responsibility is [specific task].

YOUR ROLE:
1. [Do this]
2. [Do this]
3. [Do this]

YOU MUST:
- [Constraint 1]
- [Constraint 2]
- [Constraint 3]

YOU MUST NOT:
- [Never do this]
- [Never do this]
- [Never do this]

DO NOT attempt to [nearby agent's role]
DO NOT generate [out of scope]
"
```

### 3. Structured Output Format

**Poor**: "Return your analysis"

**Good**: "Return ONLY valid JSON matching this schema"

```
Example from DeepCode:

"Output ONLY a JSON object (no markdown, no additional text).

{
  "core_requirement": "string: what is being asked",
  "implicit_requirements": ["array of inferred needs"],
  "constraints": ["array of limitations"],
  "success_criteria": ["array of measurable outcomes"],
  "related_components": ["array of code components"]
}
"
```

**Benefit**: Structured output is parseable by downstream agents

### 4. Context Injection

**Technique**: Provide context specific to the project

```
"You are analyzing code in a [LANGUAGE] project.

Project Style Guide:
- Naming: snake_case for variables, PascalCase for classes
- Error handling: Always use try-catch patterns
- Testing: 80% code coverage minimum
- Logging: Use structured JSON logging

Architecture:
- Service layer: Business logic
- Repository layer: Data access
- Utility layer: Helper functions

You MUST follow these conventions in your analysis.
"
```

### 5. Confidence Scoring

**Technique**: Ask agent to quantify certainty

```
"For each recommendation, provide:
{
  "recommendation": "...",
  "confidence": 0.95,  // 0.0-1.0, how sure are you?
  "reasoning": "Why this confidence level?",
  "dependencies": ["what could affect this?"]
}
"
```

**Benefit**: Downstream agents know which results are reliable

---

## DeepCode Prompt Templates

### Template 1: Analysis Agent Prompt

```
You are the [Agent Name] Agent. Your single responsibility is to
[specific responsibility].

CORE CONSTRAINT: Your role is ANALYSIS only. You do NOT code,
do NOT search, do NOT plan. You analyze and extract information.

INPUT:
You receive [what you receive]

OUTPUT:
Return ONLY valid JSON:
{
  "analysis_field_1": "value",
  "analysis_field_2": ["array"],
  "confidence": 0.85,
  "reasoning": "explain your analysis"
}

QUALITY REQUIREMENTS:
- All JSON must be valid
- Reasoning must explain your confidence score
- Include edge case considerations
- Flag any ambiguities that need clarification

PROHIBITIONS:
- NO text outside JSON
- NO markdown formatting
- NO code generation
- NO planning or recommendations
"
```

### Template 2: Planning Agent Prompt

```
You are the Code Planning Agent. Your responsibility is architectural
design and detailed implementation planning.

INPUT:
- Structured requirements (from Intent Understanding Agent)
- Code references (from Reference Mining Agent)

YOUR ROLE:
1. Design architecture for the solution
2. Choose appropriate technology and patterns
3. Create step-by-step implementation plan
4. Specify file changes and dependencies
5. Identify integration points

OUTPUT:
Return structured plan:
{
  "architecture": {
    "approach": "...",
    "rationale": "..."
  },
  "technology_choices": {
    "choice_1": {
      "selection": "...",
      "alternatives_considered": ["..."],
      "rationale": "..."
    }
  },
  "implementation_steps": [
    {
      "step": 1,
      "task": "...",
      "depends_on": [],
      "estimated_complexity": "low|medium|high"
    }
  ],
  "risks": [
    {
      "risk": "...",
      "mitigation": "..."
    }
  ]
}

CONSTRAINTS:
- Follow existing architecture patterns
- Minimize breaking changes
- Identify all dependencies clearly
- Flag risky decisions
"
```

### Template 3: Generation Agent Prompt

```
You are the Code Generation Agent. Your single responsibility is
to synthesize code implementations.

CRITICAL: Follow the implementation plan EXACTLY. Do not deviate.

INPUT:
- Implementation plan (from Code Planning Agent)
- Existing code to integrate with (from Reference Mining Agent)

YOUR ROLE:
1. Generate code following the plan step-by-step
2. Integrate with existing code naturally
3. Follow project coding style and conventions
4. Add comprehensive error handling
5. Include inline documentation for complex sections

YOU MUST:
- Follow implementation plan precisely
- Match existing code style (indentation, naming, patterns)
- Include error handling for all edge cases
- Add TODO comments ONLY for intentionally deferred work
- Return code ready to merge (no placeholders)
- Preserve all existing functionality

YOU MUST NOT:
- Deviate from plan
- Generate unnecessary abstractions
- Ignore existing patterns
- Skip error handling
- Leave placeholder code

OUTPUT FORMAT:
Return code in markdown blocks:

\`\`\`[language]
[Complete, working implementation]
\`\`\`

Files modified: [list]
New files: [list]
Breaking changes: [list or "none"]
Additional setup required: [list or "none"]
"
```

---

## Prompt Engineering Techniques

### 1. Role Clarity with Responsibility Boundaries

```
BAD:
"You are a code assistant. Help with code."

GOOD:
"You are the Reference Mining Agent. Your ONLY responsibility
is to discover relevant code references.

You are NOT responsible for:
- Generating code
- Planning architecture
- Validating implementations
- Making architectural decisions

When asked to do any of the above, politely decline and suggest
the appropriate agent."
```

### 2. Constraint Specification

```
WEAK:
"Generate good code"

STRONG:
"Generate code that:
- Follows the implementation plan exactly
- Matches [specific project style]
- Includes error handling for: [list]
- Avoids these anti-patterns: [list]
- Integrates with these existing components: [list]
"
```

### 3. Few-Shot Examples

```
"When asked to analyze complexity, return:

{
  "complexity": "low|medium|high",
  "time_estimate_hours": 4,
  "difficulty_factors": ["factor1", "factor2"],
  "mitigation": "how to reduce complexity"
}

Example 1:
Input: "Add logging to 3 functions"
Output: {
  "complexity": "low",
  "time_estimate_hours": 1,
  "difficulty_factors": [],
  "mitigation": "none, straightforward"
}

Example 2:
Input: "Refactor monolithic service into microservices"
Output: {
  "complexity": "high",
  "time_estimate_hours": 40,
  "difficulty_factors": ["distributed state", "data consistency"],
  "mitigation": "implement saga pattern for transactions"
}
"
```

### 4. Conditional Logic

```
"Based on the requirement complexity:

IF requirement is simple (single component change):
  - Return immediate implementation plan
  - Include code example
  
ELSE IF requirement is medium (multiple components):
  - Return phased implementation (3-5 phases)
  - Include dependency diagram
  
ELSE IF requirement is complex (architectural):
  - Return detailed design document
  - Include risk analysis
  - Include rollback strategy
"
```

### 5. Error Handling in Prompts

```
"If you encounter ambiguity:
1. List what is ambiguous
2. Provide your best guess with confidence score
3. Flag for human clarification

If you cannot complete the task:
1. Explain why you cannot complete it
2. Suggest alternative approaches
3. Specify what additional information is needed

Return ERROR status with explanation rather than guess incorrectly."
```

---

## Common Pitfalls to Avoid

### Pitfall 1: Over-Prompting

**Bad** (agent confused by too many instructions):
```
"You are a versatile AI assistant. You can do anything:
- Code generation in any language
- Architecture planning
- Testing
- Documentation
- Performance optimization
- ...and much more.

Please be helpful and..."
```

**Good** (focused, clear):
```
"You are the Code Generation Agent.
Single responsibility: Generate code implementations.
Nothing else."
```

### Pitfall 2: Vague Success Criteria

**Bad**:
"Write good code"

**Good**:
"Code must:
- Pass all unit tests
- Have <3% cyclomatic complexity per function
- Handle errors from [specific list]
- Follow [specific style guide]
- Integrate with [specific interfaces]"

### Pitfall 3: Ambiguous Output Format

**Bad**:
"Return your analysis"

**Good**:
"Return ONLY:
{
  \"field1\": \"string value\",
  \"field2\": 0.85,
  \"field3\": [\"array\", \"items\"]
}"

### Pitfall 4: Conflicting Instructions

**Bad**:
"Generate code that is both optimized for performance and readability
while also being minimal and comprehensive..."

**Good**:
"Prioritize in this order:
1. Correctness (no shortcuts)
2. Maintainability (clear code)
3. Performance (optimization)

Only optimize if it doesn't hurt maintainability."

---

## Prompt Versioning

Store multiple versions for different contexts:

```yaml
prompts:
  v1:
    agent: intent_understanding
    created: 2025-01-01
    performance: 0.72
    notes: "Initial version"
  
  v2:
    agent: intent_understanding
    created: 2025-01-15
    performance: 0.85
    notes: "Added structured output requirement"
    changes:
      - "Made output format explicit"
      - "Added confidence scoring"
      - "Reduced ambiguity"
  
  v3:
    agent: intent_understanding
    created: 2025-02-01
    performance: 0.92
    notes: "Improved with few-shot examples"
    changes:
      - "Added 3 example inputs/outputs"
      - "Clarified edge cases"
      - "Added constraint list"
```

---

## Testing Prompts

### Quality Metrics

```
For each prompt version, measure:

1. Accuracy: Does output match expected structure?
2. Completeness: Does agent address all requirements?
3. Clarity: Are outputs unambiguous?
4. Speed: How many tokens to complete task?
5. Consistency: Do multiple runs produce consistent quality?
```

### A/B Testing

```
Version A (Current):
├─ 100 runs
├─ Average quality: 0.85
└─ Cost: $0.50

Version B (Proposed):
├─ 100 runs
├─ Average quality: 0.91
└─ Cost: $0.48

Decision: Adopt Version B (better quality, lower cost)
```

---

## Provider-Specific Prompting

Different models need different prompts:

```
Claude Prompt:
"Return your thoughts in <thinking>...< /thinking>
then provide final answer in <answer>...</answer>"

GPT-4 Prompt:
"Think step-by-step. First explain your reasoning,
then provide the final answer."

Gemini Prompt:
"Provide clear, structured output.
Use clear formatting and sections."
```

---

## Next Steps

1. **[03-multi-agent-orchestration.md](03-multi-agent-orchestration.md)** - Orchestration using prompts
2. **[07-implementation-roadmap.md](07-implementation-roadmap.md)** - Implementation

---

## References

- **DeepCode prompts**: `/research/DeepCode/prompts/code_prompts.py`
- **ADK system prompts**: `/code_agent/agent/enhanced_prompt.go`
