# Tree-Sitter Go: Quick Reference Examples

## Quick Start (Copy & Paste Ready)

### 1. Basic Setup & Parse

```go
package main

import (
    "context"
    "fmt"
    
    sitter "github.com/smacker/go-tree-sitter"
    "github.com/smacker/go-tree-sitter/golang"
)

func main() {
    source := []byte(`
        package main
        
        func main() {
            fmt.Println("Hello, World!")
        }
    `)
    
    // Create parser
    parser := sitter.NewParser()
    parser.SetLanguage(golang.GetLanguage())
    
    // Parse source code
    tree, err := parser.ParseCtx(context.Background(), nil, source)
    if err != nil {
        panic(err)
    }
    defer tree.Close()
    
    root := tree.RootNode()
    fmt.Println("Root type:", root.Type())
    fmt.Println("Has error:", root.HasError())
    fmt.Println("Child count:", root.ChildCount())
}
```

### 2. Inspect Syntax Tree

```go
package main

import (
    "fmt"
    sitter "github.com/smacker/go-tree-sitter"
    "github.com/smacker/go-tree-sitter/golang"
)

func printTree(node *sitter.Node, indent string) {
    fmt.Printf("%s%s [%d-%d]\n", 
        indent, 
        node.Type(),
        node.StartByte(),
        node.EndByte(),
    )
    
    for i := uint32(0); i < node.ChildCount(); i++ {
        child := node.Child(int(i))
        printTree(child, indent+"  ")
    }
}

func main() {
    source := []byte("func foo() {}")
    
    parser := sitter.NewParser()
    parser.SetLanguage(golang.GetLanguage())
    tree, _ := parser.ParseCtx(context.Background(), nil, source)
    
    printTree(tree.RootNode(), "")
    /* Output:
    source_file [0-12]
      function_declaration [0-12]
        func [0-4]
        identifier [5-8]
        parameters [8-10]
        block [10-12]
    */
}
```

### 3. Extract Function Names

```go
package main

import (
    "context"
    "fmt"
    
    sitter "github.com/smacker/go-tree-sitter"
    "github.com/smacker/go-tree-sitter/golang"
)

func extractFunctionNames(source []byte) []string {
    parser := sitter.NewParser()
    parser.SetLanguage(golang.GetLanguage())
    tree, _ := parser.ParseCtx(context.Background(), nil, source)
    
    var names []string
    
    var visit func(*sitter.Node)
    visit = func(node *sitter.Node) {
        if node.Type() == "function_declaration" {
            // Get the function name (usually first named child that's an identifier)
            for i := uint32(0); i < node.NamedChildCount(); i++ {
                child := node.NamedChild(int(i))
                if child != nil && child.Type() == "identifier" {
                    names = append(names, child.Content(source))
                    break
                }
            }
        }
        
        for i := uint32(0); i < node.ChildCount(); i++ {
            if child := node.Child(int(i)); child != nil {
                visit(child)
            }
        }
    }
    
    visit(tree.RootNode())
    return names
}

func main() {
    code := []byte(`
        func hello() {}
        func world() {}
        func foo() {}
    `)
    
    names := extractFunctionNames(code)
    fmt.Println("Functions:", names) // [hello world foo]
}
```

### 4. Query Patterns

```go
package main

import (
    "context"
    "fmt"
    
    sitter "github.com/smacker/go-tree-sitter"
    "github.com/smacker/go-tree-sitter/golang"
)

func findFunctionWithQuery(source []byte) {
    parser := sitter.NewParser()
    parser.SetLanguage(golang.GetLanguage())
    tree, _ := parser.ParseCtx(context.Background(), nil, source)
    
    // Query to find all function names
    query, _ := sitter.NewQuery([]byte(`
        (function_declaration
          name: (identifier) @func_name)
    `), golang.GetLanguage())
    
    cursor := sitter.NewQueryCursor()
    cursor.Exec(query, tree.RootNode())
    
    fmt.Println("Functions found with query:")
    for {
        match, ok := cursor.NextMatch()
        if !ok {
            break
        }
        
        for _, capture := range match.Captures {
            fmt.Println("  -", capture.Node.Content(source))
        }
    }
    
    cursor.Close()
    query.Close()
}

func main() {
    code := []byte(`
        func Parse() {}
        func Handle() {}
    `)
    
    findFunctionWithQuery(code)
}
```

### 5. Find Syntax Errors

```go
package main

import (
    "context"
    "fmt"
    
    sitter "github.com/smacker/go-tree-sitter"
    "github.com/smacker/go-tree-sitter/golang"
)

func findErrors(source []byte) []string {
    parser := sitter.NewParser()
    parser.SetLanguage(golang.GetLanguage())
    tree, _ := parser.ParseCtx(context.Background(), nil, source)
    
    var errors []string
    
    var visit func(*sitter.Node)
    visit = func(node *sitter.Node) {
        if node.IsError() || node.IsMissing() {
            errors = append(errors, fmt.Sprintf(
                "Error at %d:%d: %s", 
                node.StartPoint().Row,
                node.StartPoint().Column,
                node.Type(),
            ))
        }
        
        for i := uint32(0); i < node.ChildCount(); i++ {
            if child := node.Child(int(i)); child != nil {
                visit(child)
            }
        }
    }
    
    visit(tree.RootNode())
    return errors
}

func main() {
    // Invalid syntax
    code := []byte("func foo(( {}")
    
    errors := findErrors(code)
    for _, e := range errors {
        fmt.Println(e)
    }
}
```

### 6. Extract and Filter Large Function

```go
package main

import (
    "context"
    "fmt"
    
    sitter "github.com/smacker/go-tree-sitter"
    "github.com/smacker/go-tree-sitter/golang"
)

type FunctionInfo struct {
    Name      string
    LineCount int
    IsTooLong bool
}

func analyzeFunctions(source []byte) []FunctionInfo {
    parser := sitter.NewParser()
    parser.SetLanguage(golang.GetLanguage())
    tree, _ := parser.ParseCtx(context.Background(), nil, source)
    
    var functions []FunctionInfo
    
    var visit func(*sitter.Node)
    visit = func(node *sitter.Node) {
        if node.Type() == "function_declaration" {
            // Get function name
            var name string
            for i := uint32(0); i < node.NamedChildCount(); i++ {
                child := node.NamedChild(int(i))
                if child != nil && child.Type() == "identifier" {
                    name = child.Content(source)
                    break
                }
            }
            
            // Calculate line count
            lineCount := int(node.EndPoint().Row - node.StartPoint().Row)
            
            functions = append(functions, FunctionInfo{
                Name:      name,
                LineCount: lineCount,
                IsTooLong: lineCount > 50,
            })
        }
        
        for i := uint32(0); i < node.ChildCount(); i++ {
            if child := node.Child(int(i)); child != nil {
                visit(child)
            }
        }
    }
    
    visit(tree.RootNode())
    return functions
}

func main() {
    code := []byte(`
        func short() { println("hi") }
        
        func veryLongFunction() {
            // ... 60 lines of code ...
        }
    `)
    
    funcs := analyzeFunctions(code)
    for _, f := range funcs {
        status := ""
        if f.IsTooLong {
            status = " ⚠️  TOO LONG"
        }
        fmt.Printf("%s (%d lines)%s\n", f.Name, f.LineCount, status)
    }
}
```

### 7. Language Detection & Support

```go
package main

import (
    "context"
    "fmt"
    
    sitter "github.com/smacker/go-tree-sitter"
    "github.com/smacker/go-tree-sitter/golang"
    "github.com/smacker/go-tree-sitter/python"
    "github.com/smacker/go-tree-sitter/javascript"
)

func parseByLanguage(source []byte, lang string) (*sitter.Tree, error) {
    parser := sitter.NewParser()
    
    switch lang {
    case "go", "golang":
        parser.SetLanguage(golang.GetLanguage())
    case "python":
        parser.SetLanguage(python.GetLanguage())
    case "javascript", "js":
        parser.SetLanguage(javascript.GetLanguage())
    default:
        return nil, fmt.Errorf("unsupported language: %s", lang)
    }
    
    tree, err := parser.ParseCtx(context.Background(), nil, source)
    return tree, err
}

func main() {
    goCode := []byte("func main() {}")
    pythonCode := []byte("def main(): pass")
    jsCode := []byte("function main() {}")
    
    tree1, _ := parseByLanguage(goCode, "go")
    tree2, _ := parseByLanguage(pythonCode, "python")
    tree3, _ := parseByLanguage(jsCode, "javascript")
    
    fmt.Println("Go:", tree1.RootNode().Type())
    fmt.Println("Python:", tree2.RootNode().Type())
    fmt.Println("JavaScript:", tree3.RootNode().Type())
    
    tree1.Close()
    tree2.Close()
    tree3.Close()
}
```

### 8. Complex Query: Find Function Parameters

```go
package main

import (
    "context"
    "fmt"
    
    sitter "github.com/smacker/go-tree-sitter"
    "github.com/smacker/go-tree-sitter/golang"
)

type Parameter struct {
    Name string
    Type string
}

type FunctionSignature struct {
    Name       string
    Parameters []Parameter
    ReturnType string
}

func analyzeFunctionSignatures(source []byte) []FunctionSignature {
    parser := sitter.NewParser()
    parser.SetLanguage(golang.GetLanguage())
    tree, _ := parser.ParseCtx(context.Background(), nil, source)
    
    var signatures []FunctionSignature
    
    var visit func(*sitter.Node)
    visit = func(node *sitter.Node) {
        if node.Type() == "function_declaration" {
            sig := FunctionSignature{}
            
            // Extract function name
            nameNode := node.NamedChild(0)
            if nameNode != nil && nameNode.Type() == "identifier" {
                sig.Name = nameNode.Content(source)
            }
            
            // Extract parameters (simplified)
            for i := uint32(0); i < node.NamedChildCount(); i++ {
                child := node.NamedChild(int(i))
                if child != nil && child.Type() == "parameter_list" {
                    // Parse parameter_list content
                    sig.Parameters = []Parameter{
                        {Name: "param1", Type: "string"},
                        // Real parsing would traverse the parameter_list
                    }
                }
            }
            
            signatures = append(signatures, sig)
        }
        
        for i := uint32(0); i < node.ChildCount(); i++ {
            if child := node.Child(int(i)); child != nil {
                visit(child)
            }
        }
    }
    
    visit(tree.RootNode())
    return signatures
}

func main() {
    code := []byte(`
        func add(a int, b int) int {
            return a + b
        }
    `)
    
    sigs := analyzeFunctionSignatures(code)
    for _, sig := range sigs {
        fmt.Printf("Function: %s\n", sig.Name)
        for _, param := range sig.Parameters {
            fmt.Printf("  - %s: %s\n", param.Name, param.Type)
        }
    }
}
```

### 9. Incremental Parsing (Editor Use Case)

```go
package main

import (
    "context"
    "fmt"
    
    sitter "github.com/smacker/go-tree-sitter"
    "github.com/smacker/go-tree-sitter/python"
)

func main() {
    source := []byte("def foo(): pass")
    
    parser := sitter.NewParser()
    parser.SetLanguage(python.GetLanguage())
    
    // Initial parse
    tree, _ := parser.ParseCtx(context.Background(), nil, source)
    fmt.Println("Initial root:", tree.RootNode().Type())
    
    // Simulate editing: change "foo" to "bar"
    newSource := []byte("def bar(): pass")
    
    // Inform tree of changes (byte position 4-7)
    tree.Edit(sitter.EditInput{
        StartIndex:  4,
        OldEndIndex: 7,
        NewEndIndex: 7,
        StartPoint:  sitter.Point{Row: 0, Column: 4},
        OldEndPoint: sitter.Point{Row: 0, Column: 7},
        NewEndPoint: sitter.Point{Row: 0, Column: 7},
    })
    
    // Re-parse is faster now (uses previous tree)
    newTree, _ := parser.ParseCtx(context.Background(), tree, newSource)
    fmt.Println("Updated root:", newTree.RootNode().Type())
    
    newTree.Close()
}
```

### 10. Parallel Processing (Multiple Files)

```go
package main

import (
    "context"
    "fmt"
    "sync"
    
    sitter "github.com/smacker/go-tree-sitter"
    "github.com/smacker/go-tree-sitter/golang"
)

func analyzeMultipleFiles(files map[string][]byte) map[string]int {
    parser := sitter.NewParser()
    parser.SetLanguage(golang.GetLanguage())
    
    results := make(map[string]int)
    var mu sync.Mutex
    var wg sync.WaitGroup
    
    for name, source := range files {
        wg.Add(1)
        
        go func(filename string, code []byte) {
            defer wg.Done()
            
            // Parse file
            tree, _ := parser.ParseCtx(context.Background(), nil, code)
            
            // Count functions
            count := 0
            var visit func(*sitter.Node)
            visit = func(node *sitter.Node) {
                if node.Type() == "function_declaration" {
                    count++
                }
                for i := uint32(0); i < node.ChildCount(); i++ {
                    if child := node.Child(int(i)); child != nil {
                        visit(child)
                    }
                }
            }
            visit(tree.RootNode())
            
            mu.Lock()
            results[filename] = count
            mu.Unlock()
            
            tree.Close()
        }(name, source)
    }
    
    wg.Wait()
    return results
}

func main() {
    files := map[string][]byte{
        "main.go": []byte("func main() {} func foo() {}"),
        "util.go": []byte("func bar() {} func baz() {} func qux() {}"),
    }
    
    results := analyzeMultipleFiles(files)
    for file, count := range results {
        fmt.Printf("%s: %d functions\n", file, count)
    }
}
```

---

## Installation Verification

Run this to verify installation:

```bash
# Create test.go
cat > test.go << 'EOF'
package main

import (
    sitter "github.com/smacker/go-tree-sitter"
    "github.com/smacker/go-tree-sitter/golang"
)

func main() {
    parser := sitter.NewParser()
    parser.SetLanguage(golang.GetLanguage())
    tree, _ := parser.ParseCtx(nil, nil, []byte("func main() {}"))
    println(tree.RootNode().Type())
}
EOF

# Run
go mod init test
go get github.com/smacker/go-tree-sitter
go run test.go

# Should output: source_file
```

---

## Common Patterns & Recipes

### Extract All Imports

```go
func extractImports(source []byte) []string {
    // ... setup parser ...
    
    query, _ := sitter.NewQuery([]byte(`
        (import_spec
          path: (interpreted_string_literal) @import)
    `), golang.GetLanguage())
    
    var imports []string
    // ... execute query and collect ...
    return imports
}
```

### Find Unused Variables

```go
func findUnusedVariables(source []byte) []string {
    // Complex pattern involving:
    // 1. Find all var_declaration nodes
    // 2. Count uses via identifier nodes
    // 3. Report those with 1 reference (the declaration)
}
```

### Detect Dead Code

```go
func findDeadCode(source []byte) []string {
    // Patterns:
    // 1. Unreachable code after return
    // 2. Unreachable code after panic
    // 3. Unreachable code after break/continue
}
```

### Extract Comments

```go
func extractComments(source []byte) []string {
    // Query for comment nodes specifically
    // tree-sitter marks them as "comment" type
}
```

---

## Performance Tips

1. **Cache parsers** - Create once, reuse many times
2. **Cache queries** - Compile once, reuse many times
3. **Use incremental updates** - Much faster than full re-parse
4. **Set context timeout** - Prevent hangs on large files
5. **Batch operations** - Parse multiple files in parallel

---

## Troubleshooting

| Problem | Solution |
|---------|----------|
| "No language set" | Call `parser.SetLanguage()` first |
| Query not matching | Debug with `root.String()` to see S-expressions |
| Memory growing | Call `.Close()` on trees and queries |
| Parse timeout | Use `context.WithTimeout()` |
| Slow parsing | Use incremental `.Edit()` instead of full parse |

---

## References

- **Go Bindings**: https://pkg.go.dev/github.com/smacker/go-tree-sitter
- **Tree-Sitter Docs**: https://tree-sitter.github.io/tree-sitter/
- **Query Syntax**: https://tree-sitter.github.io/tree-sitter/using-parsers#queries
- **GitHub**: https://github.com/smacker/go-tree-sitter
