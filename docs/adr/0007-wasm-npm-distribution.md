# ADR-0007: WebAssembly (WASM) Compilation Target & npm CLI Distribution

**Status**: Proposed  
**Date**: 2025-11-15  
**Revision**: 1  
**Owner**: adk-code Team  
**Context**: Research on distributing adk-code as a cross-platform npm package via WASM  

---

## Executive Summary

This ADR explores **three approaches** to compile adk-code to WebAssembly and distribute it as an npm CLI package:

1. **Standard Go + wasm_exec.js** – Full GOOS=js/GOARCH=wasm, 2-4MB bundle
2. **TinyGo + WASI** – Smaller footprint (~100-400kB), limited stdlib
3. **Hybrid Native + WASM** – Native for core agent, WASM for embedded/web scenarios

**Recommendation**: Not recommended for primary adk-code distribution at this time due to I/O limitations, but suitable as an **optional secondary distribution channel** for specific use cases (browser-based, lightweight agents).

---

## The Problem

### Current Distribution Model
- adk-code is distributed as native Go binaries (macOS, Linux, Windows)
- Requires platform-specific builds and updates
- Installation: `brew install` (macOS), binary download, or `go install`

### Why Consider WASM?
1. **Single distributable format** – One npm package for all platforms
2. **Easy installation** – `npm install -g adk-code-wasm`
3. **Web/Node.js compatibility** – Run in browsers and Node.js servers
4. **Cross-platform guarantee** – Same code everywhere (Windows/Mac/Linux/Web)
5. **Community reach** – npm has 20M+ weekly downloads (vs Homebrew's smaller reach)

### Why WASM is Challenging for adk-code
1. **Heavy I/O operations** – File reading/writing, folder traversal
   - WASM syscalls have performance overhead vs native
   - Some syscalls unavailable (subprocess spawning requires workarounds)

2. **File system access** – Limited by sandbox in browser contexts
   - Node.js WASM has access but requires async patterns
   - Go's native syscall/fs assumptions don't map cleanly

3. **Bundle size** – Go WASM baseline ~2-4MB, adk-code ~10-16MB
   - Adds 40-50% overhead vs native binary (~8-10MB compressed)
   - npm downloads less efficient than direct binary distribution

4. **Subprocess execution** – Go's os.Exec doesn't translate to WASM naturally
   - Would need to reimpl via Node.js child_process bindings
   - Performance penalty for agent tool execution

5. **Performance parity** – WASM not ideal for long-running REPL sessions
   - Native Go binaries have better GC, memory efficiency
   - WASM adds 10-20% latency on syscall-heavy workloads

---

## Solution Architecture

### Option 1: Standard Go WASM (GOOS=js GOARCH=wasm)

#### What It Does
```bash
GOOS=js GOARCH=wasm go build -o adk-code.wasm main.go
```

Produces:
- `adk-code.wasm` (~4-8MB uncompressed, ~1-2MB gzip)
- Uses Go 1.11+ JavaScript interop (syscall/js)
- Requires `wasm_exec.js` helper file (Go runtime)
- Runs in browsers and Node.js

#### npm Package Structure
```
adk-code-wasm/
├── package.json
│   ├── "bin": { "adk-code": "./bin/cli.js" }
│   ├── "main": "./lib/index.js"
│   └── "files": ["lib/", "bin/", "adk-code.wasm", "wasm_exec.js"]
├── bin/
│   └── cli.js (Node.js wrapper script)
├── lib/
│   ├── index.js (WASM module loader & bindings)
│   ├── commands.go (compiled to WASM)
│   └── agent.go (compiled to WASM)
├── adk-code.wasm (Go WASM binary)
└── wasm_exec.js (Go runtime, copied from $GOROOT/lib/wasm)
```

#### CLI Wrapper (bin/cli.js)
```javascript
#!/usr/bin/env node
const path = require('path');
const fs = require('fs');

// Load WASM module
const wasmPath = path.join(__dirname, '../adk-code.wasm');
const wasmBuffer = fs.readFileSync(wasmPath);

// Setup runtime
const { TextEncoder, TextDecoder } = require('util');
globalThis.TextEncoder = TextEncoder;
globalThis.TextDecoder = TextDecoder;

// Load Go runtime from wasm_exec.js
const go = new Go();

// Instantiate and run WASM
WebAssembly.instantiate(wasmBuffer, go.importObject).then(result => {
  // Pass process.argv to Go program
  process.argv = ['node', 'adk-code', ...process.argv.slice(2)];
  go.run(result.instance);
});
```

#### Limitations
- ❌ File I/O slower than native (syscall translation overhead)
- ❌ `os.Exec` doesn't work (no subprocess spawning)
- ❌ Large bundle size (2-4MB uncompressed, 500KB-1MB gzip)
- ❌ No direct access to process signals or timers
- ✅ Full Go stdlib available
- ✅ Browser + Node.js compatible

#### Suitability for adk-code
**Medium risk** – Could work for agent commands that are compute-intensive but moderate I/O. REPL would be slow due to async file operations.

---

### Option 2: TinyGo + WASI (GOOS=wasip1 GOARCH=wasm)

#### What It Does
```bash
# Requires Go 1.21+
GOOS=wasip1 GOARCH=wasm go build -o adk-code.wasm main.go
```

Produces:
- `adk-code.wasm` (~100-400kB uncompressed, ~30-100kB gzip)
- Uses WebAssembly System Interface (WASI) for syscalls
- Better performance than js/wasm for I/O operations
- Runs in Node.js via wasmtime/wasmer, or WASI runtimes

#### npm Package Structure (Similar to Option 1)
```
adk-code-wasi/
├── bin/cli.js (loads WASI runtime + wasm)
├── adk-code.wasm (WASI binary)
└── node_modules include: @wasmer/wasm (or wasmtime)
```

#### CLI Wrapper Enhancement (bin/cli.js)
```javascript
#!/usr/bin/env node
const path = require('path');
const fs = require('fs');
const { WASI } = require('wasi');

const wasmPath = path.join(__dirname, '../adk-code.wasm');
const wasm = new WebAssembly.Instance(
  new WebAssembly.Module(fs.readFileSync(wasmPath)),
  new WASI({ args: process.argv.slice(2) }).wasiImport
);
wasm.instance.exports._start();
```

#### Advantages vs Option 1
- ✅ 90% smaller bundle (100-400kB vs 2-4MB)
- ✅ Better I/O performance (native WASI syscalls)
- ✅ Subprocess spawning via WASI interface
- ✅ Faster startup time
- ❌ Limited stdlib (Go 1.21 WASIP1 limitations)
- ❌ Newer, less mature than js/wasm
- ❌ Requires wasmtime/wasmer runtime in npm dependencies

#### Suitability for adk-code
**Lower risk, higher reward** – Smaller bundle, better I/O, but stdlib limitations may require code adjustments.

---

### Option 3: Hybrid Approach (Recommended Investigation Path)

#### Concept
- **Native binary** (macOS, Linux, Windows) = primary distribution
- **WASM variant** (browser + Node.js) = secondary for embedded/restricted environments
- Split functionality:
  - Agent core (compute-heavy) → WASM-friendly
  - Tool system (I/O-heavy) → Native fallback on demand

#### Implementation Pattern
```go
// main.go
func main() {
    // Detect runtime
    if isWASM() {
        // Lightweight agent mode (restricted tool set)
        runWASMAgent()
    } else {
        // Full native agent mode
        runNativeAgent()
    }
}

// Build scripts
// Native: go build -o adk-code main.go
// WASM:   GOOS=js GOARCH=wasm go build -o pkg/adk-code.wasm main.go
```

#### Advantages
- ✅ Preserves native binary as primary (no compromises)
- ✅ WASM as opt-in for specific users
- ✅ Better testing and validation path
- ✅ Can iterate on WASM separately
- ❌ Requires maintaining two build paths
- ❌ More complex CI/CD
- ❌ Potential code duplication

#### Suitability for adk-code
**Best approach** – Lets us deliver native binaries unchanged while exploring WASM adoption incrementally.

---

## Technical Deep Dive: Building adk-code for WASM

### Build Process (Standard Go + js/wasm)

#### Step 1: Verify Go Environment
```bash
go version  # Must be 1.11+
echo $GOROOT
ls $GOROOT/lib/wasm/wasm_exec.js  # Verify support files exist
```

#### Step 2: Compile to WASM
```bash
# Set environment variables
export GOOS=js
export GOARCH=wasm

# Build with optimization flags
go build \
  -o adk-code.wasm \
  -ldflags="-s -w" \  # Strip symbols, reduce size by ~30%
  -tags=wasm \        # Custom build tag if needed
  main.go

# Verify output
ls -lh adk-code.wasm  # Expected: 2-4MB
file adk-code.wasm    # Should be WebAssembly binary
```

#### Step 3: Create npm Wrapper Structure
```bash
# Create package directories
mkdir -p dist/bin dist/lib

# Copy WASM binary
cp adk-code.wasm dist/adk-code.wasm

# Copy runtime support
cp $GOROOT/lib/wasm/wasm_exec.js dist/wasm_exec.js

# Create wrapper script
cat > dist/bin/cli.js << 'EOF'
#!/usr/bin/env node
const path = require('path');
const fs = require('fs');
const go = new Go();
const wasmBuffer = fs.readFileSync(path.join(__dirname, '../adk-code.wasm'));
process.argv = ['node', 'adk-code', ...process.argv.slice(2)];
WebAssembly.instantiate(wasmBuffer, go.importObject)
  .then(r => go.run(r.instance))
  .catch(err => { console.error(err); process.exit(1); });
EOF
chmod +x dist/bin/cli.js
```

#### Step 4: Create package.json
```json
{
  "name": "@adk-code/wasm",
  "version": "0.1.0",
  "description": "adk-code CLI compiled to WebAssembly for npm",
  "type": "module",
  "bin": {
    "adk-code": "./bin/cli.js"
  },
  "main": "./lib/index.js",
  "files": [
    "bin/",
    "lib/",
    "adk-code.wasm",
    "wasm_exec.js"
  ],
  "engines": {
    "node": ">=16.0.0"
  },
  "keywords": ["cli", "agent", "wasm", "webassembly"]
}
```

#### Step 5: Test Locally
```bash
# Link package for local testing
npm link

# Test CLI invocation
adk-code version
adk-code run "list files"

# Test in another directory
cd /tmp && adk-code --help
```

#### Step 6: Publish to npm
```bash
npm publish --access public

# Or scoped:
npm publish --access public  # @adk-code/wasm
```

---

## Handling I/O and Syscalls

### File System Access via syscall/js

Go's `syscall/js` package provides JavaScript interop:

```go
// Example: Read file in WASM context
import "syscall/js"

func readFileWASM(path string) ([]byte, error) {
    // Call JavaScript fs.readFileSync
    fs := js.Global().Get("fs")
    result := fs.Call("readFileSync", path)
    
    if result.Type() == js.TypeNull {
        return nil, fmt.Errorf("file not found")
    }
    
    // Convert Uint8Array to Go byte slice
    buffer := make([]byte, result.Get("length").Int())
    js.CopyBytesToGo(buffer, result)
    return buffer, nil
}
```

### Subprocess Execution Workaround

For tool execution in WASM (e.g., spawning `curl`, `git`):

```go
// Not available in WASM directly
// Instead, use Node.js child_process via syscall/js:

import "syscall/js"

func execCommandWASM(cmd string, args []string) (string, error) {
    execSync := js.Global().Get("require").Call("eval", "require('child_process').execSync")
    
    cmdStr := cmd + " " + strings.Join(args, " ")
    result := execSync.Invoke(cmdStr, js.ValueOf(map[string]interface{}{
        "encoding": "utf-8",
    }))
    
    return result.String(), nil
}
```

### Performance Implications
- File I/O: **1.5-3x slower** than native (syscall translation overhead)
- Subprocess calls: **2-5x slower** (IPC vs direct execution)
- Memory usage: **Similar** to native (Go's GC is portable)
- Startup time: **500ms-1s slower** (WASM instantiation + module loading)

---

## Testing & Validation Strategy

### Unit Tests (WASM-specific)
```go
// Build tag for WASM tests
//go:build wasm
// +build wasm

package main

import "testing"

func TestFileOperationWASM(t *testing.T) {
    // Test file read via syscall/js
    data, err := readFileWASM("test.txt")
    if err != nil {
        t.Fatal(err)
    }
    if len(data) == 0 {
        t.Error("expected non-empty file")
    }
}

func TestAgentExecutionWASM(t *testing.T) {
    // Test agent runs in WASM context
    // Would be integration test, not unit test
}
```

### Integration Tests (Node.js)
```javascript
// test/wasm.test.js
const { spawn } = require('child_process');

describe('adk-code WASM CLI', () => {
  it('should run version command', (done) => {
    const proc = spawn('node', ['./bin/cli.js', 'version']);
    
    let output = '';
    proc.stdout.on('data', (data) => { output += data; });
    
    proc.on('close', (code) => {
      expect(code).toBe(0);
      expect(output).toMatch(/adk-code/);
      done();
    });
  });
});
```

### Performance Benchmarks
```bash
# Compare native vs WASM
time adk-code run "command"           # Native: ~100ms
time adk-code-wasm run "command"      # WASM: ~150-300ms

# File I/O benchmark
time adk-code grep-code "pattern" .   # Native: ~200ms
time adk-code-wasm grep-code "pattern" .  # WASM: ~500-1000ms
```

---

## npm Publishing & Distribution

### Package Metadata
```json
{
  "name": "@adk-code/wasm",
  "version": "0.1.0",
  "description": "adk-code AI agent compiled to WebAssembly",
  "keywords": ["cli", "agent", "ai", "wasm", "webassembly", "npm"],
  "author": "adk-code Team",
  "license": "MIT",
  "homepage": "https://github.com/raphaelmansuy/adk-code",
  "repository": {
    "type": "git",
    "url": "https://github.com/raphaelmansuy/adk-code.git",
    "directory": "pkg/npm-wasm"
  },
  "bugs": {
    "url": "https://github.com/raphaelmansuy/adk-code/issues"
  },
  "type": "module",
  "main": "./lib/index.js",
  "bin": {
    "adk-code": "./bin/cli.js"
  },
  "files": [
    "bin/",
    "lib/",
    "adk-code.wasm",
    "wasm_exec.js",
    "README.md"
  ],
  "engines": {
    "node": ">=16.0.0"
  },
  "scripts": {
    "test": "jest",
    "prepublish": "npm run build && npm run test"
  }
}
```

### CI/CD Integration
```yaml
# .github/workflows/npm-publish-wasm.yml
name: npm WASM Publish

on:
  push:
    tags:
      - 'v*'

jobs:
  build-and-publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      
      - name: Build WASM
        run: |
          export GOOS=js GOARCH=wasm
          go build -o dist/adk-code.wasm \
            -ldflags="-s -w" main.go
      
      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: '18'
          registry-url: 'https://registry.npmjs.org'
      
      - name: Publish to npm
        working-directory: dist
        run: npm publish
        env:
          NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
```

---

## Risk Assessment & Mitigations

| Risk | Impact | Likelihood | Mitigation |
|------|--------|-----------|-----------|
| **I/O performance degradation** | Users experience slow REPL | High | Test on representative workloads; set expectations |
| **Subprocess execution failure** | Tool execution breaks | Medium | Implement fallback to JavaScript child_process |
| **Large bundle size** | Poor npm download experience | High | Use brotli compression; offer TinyGo alternative |
| **Platform limitations** | Features don't work in browsers | High | Document limitations; separate WASM-safe features |
| **Version skew with native** | Bug fixes not in WASM | Medium | Automated build+publish in CI; sync releases |
| **Memory usage in long sessions** | REPL crash after many turns | Low | Similar to native; Go GC is portable |

---

## Implementation Roadmap (If Approved)

### Phase 1: Proof of Concept (2-3 days)
- [ ] Build adk-code to WASM (GOOS=js GOARCH=wasm)
- [ ] Create npm wrapper script and package.json
- [ ] Test basic CLI commands locally (version, help)
- [ ] Document limitations and performance characteristics

### Phase 2: Integration & Testing (3-4 days)
- [ ] Integrate WASM build into Makefile (`make wasm`)
- [ ] Write WASM-specific unit tests (marked with //go:build wasm)
- [ ] Create Node.js integration tests
- [ ] Benchmark vs native (file I/O, agent execution)
- [ ] Test in multiple Node.js versions (16, 18, 20)

### Phase 3: CI/CD & Publishing (2-3 days)
- [ ] Add GitHub Actions workflow for npm publish
- [ ] Set up npm account & authentication (NPM_TOKEN)
- [ ] Create separate @adk-code/wasm package
- [ ] Publish initial version (v0.1.0-alpha)
- [ ] Test installation: `npm install -g @adk-code/wasm`

### Phase 4: Documentation (1-2 days)
- [ ] Create WASM README with limitations and use cases
- [ ] Add troubleshooting guide (common errors in Node.js WASM)
- [ ] Document fallback strategies (when to use native vs WASM)
- [ ] Add examples (browser embedding, Node.js server integration)

### Total Effort: 8-12 days

---

## Alternatives Considered

### Alternative 1: Standalone WASI Binary (Not published to npm)
- Build with GOOS=wasip1 GOARCH=wasm
- Distribute as standalone .wasm file (no npm)
- Users run via `wasmtime adk-code.wasm`
- **Rejected**: Smaller reach than npm; requires extra installation step

### Alternative 2: Rust Rewrite + wasm-pack
- Rewrite adk-code in Rust
- Use wasm-pack for automated npm packaging
- **Rejected**: Massive effort (rewrite entire agent); doesn't solve I/O limitations

### Alternative 3: JavaScript Port (Node.js native)
- Rewrite agent logic in TypeScript/Node.js
- Publish as native npm package
- **Rejected**: Loses Go advantages (performance, concurrency, type safety)

### Alternative 4: Docker/Container Distribution
- Publish Docker image with adk-code
- Users pull and run in containers
- **Rejected**: Heavier than WASM; doesn't solve cross-platform issues

---

## Success Criteria

| Criterion | Metric | Acceptance |
|-----------|--------|-----------|
| Build completeness | Can build WASM binary | Pass make wasm |
| Package quality | Valid npm package | Passes npm publish validation |
| CLI functionality | Core commands work | 80%+ command success rate vs native |
| Performance | Acceptable latency | <2x slower than native for typical ops |
| Documentation | Clear guidance | Complete README, limitations, examples |
| Testing | Automated validation | >70% test coverage (WASM-specific) |
| Downloads | Community adoption | 100+ monthly downloads in first 3 months |

---

## Decision & Next Steps

### Recommendation
**Proceed with Phase 1 (PoC)** on a **feature branch** with the following scope:

1. Build and test Go WASM compilation (Option 1)
2. Create npm wrapper and test locally
3. Document findings and performance characteristics
4. Present results to team for Phase 2 decision

### Why This Approach
- **Low risk**: PoC doesn't ship to users
- **High learning**: Validates assumptions about I/O performance
- **Informed decision**: Real data before committing to full implementation
- **Flexible**: Can pivot to TinyGo or Hybrid if learnings warrant

### Phase 1 Success = Team Approval for Phase 2-4

---

## References

### Official Documentation
- Go WebAssembly Wiki: https://go.dev/wiki/WebAssembly
- Go 1.21 WASI Support: https://go.dev/blog/wasi
- TinyGo WASM Guide: https://tinygo.org/docs/guides/webassembly/
- npm Publishing: https://docs.npmjs.com/packages-and-modules/

### Tools & Examples
- wasm-pack (Rust WASM packaging): https://rustwasm.github.io/docs/wasm-pack/
- Example: OpenReplay WASM npm package: https://blog.openreplay.com/publishing-webassembly-packages-for-npm/
- Example: SWC (Rust compiler, WASM-based): https://swc.rs/

### Related ADRs
- ADR-0006: Context Management (handles long-running sessions)
- Future: ADR-0008: Browser-based REPL (if WASM path approved)

---

## Appendix A: File Size Analysis

| Component | Standard Go | TinyGo | Compressed (gzip) |
|-----------|-----------|--------|-----------------|
| WASM binary | 2-4 MB | 100-300 kB | 500 kB - 1 MB |
| wasm_exec.js | ~20 kB | N/A | ~5 kB |
| npm package (total) | ~2.5 MB | ~150 kB | ~550 kB |
| vs Native binary | +50% overhead | -80% smaller | -40% vs native |

---

## Appendix B: Node.js Version Compatibility

| Node.js Version | WASM Support | Notes |
|---|---|---|
| 12.x | ⚠️ Partial | Experimental WASM modules (--experimental-wasm-modules) |
| 14.x | ✅ Full | Stable WASM support |
| 16.x+ | ✅ Full | Recommended minimum |

Recommended: **Node.js 16.x or later**

---

## Appendix C: Syscall Support Matrix

| Operation | Standard Go WASM | WASI (wasip1) | Native |
|-----------|---|---|---|
| File read/write | ✅ (slow) | ✅ (faster) | ✅ (fast) |
| Directory listing | ✅ (slow) | ✅ | ✅ |
| Subprocess execution | ⚠️ Workaround | ✅ (via WASI) | ✅ |
| Network (HTTP) | ✅ (via fetch) | ⚠️ Limited | ✅ |
| Signals | ❌ | ⚠️ Limited | ✅ |
| Timers | ✅ | ✅ | ✅ |

---

**End of ADR-0007**
