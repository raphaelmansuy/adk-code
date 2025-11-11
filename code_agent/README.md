# `demo` Directory

The `demo` directory within the `code_agent` project contains a self-contained C language implementation of a Prolog-like inference engine. This serves as a complex, real-world example for demonstrating the capabilities of the `code_agent` in understanding, navigating, modifying, and debugging codebases.

## Project Structure

The C project is organized into several modules, each handling a specific aspect of the inference engine:

```
.
├── main.c              # Main entry point, orchestrates the inference engine
├── parser.c/h          # Parses input into internal data structures
├── term.c/h            # Represents logical terms (variables, constants, functions)
├── clause.c/h          # Represents logical clauses (head and body literals)
├── knowledge_base.c/h  # Stores and manages facts and rules
├── substitution.c/h    # Handles variable substitutions
├── unification.c/h     # Implements the unification algorithm
└── inference.c/h       # Core inference engine (resolution, backtracking)
```

*   **`main.c`**: The primary entry point, orchestrates the overall inference process.
*   **`parser.c` / `parser.h`**: Responsible for parsing input queries and knowledge base entries into an internal representation.
*   **`term.c` / `term.h`**: Defines the data structures and operations for logical terms (variables, constants, functions).
*   **`clause.c` / `clause.h`**: Manages the representation and manipulation of logical clauses.
*   **`knowledge_base.c` / `knowledge_base.h`**: Handles the storage and retrieval of facts and rules (the knowledge base).
*   **`substitution.c` / `substitution.h`**: Manages variable substitutions during unification and inference.
*   **`unification.c` / `unification.h`**: Implements the unification algorithm, a core component for matching terms and clauses.
*   **`inference.c` / `inference.h`**: Implements the core inference mechanisms, likely including resolution and backtracking.

## Inference Process Overview

The `prolog` engine processes queries by interacting with the knowledge base through a series of steps involving parsing, unification, and inference.

```mermaid
graph TD
    A[Input Query<br>(e.g., father(X,Y)?)] --> B(Parser<br>(parser.c/h))
    B --> C{Internal Query<br>Representation}
    C --> D[Inference<br>(inference.c/h)<br>(Resolution, Backtracking)]
    D --> E[Knowledge Base<br>(knowledge_base.c/h)<br>(Facts & Rules)]
    D --> F[Unification Engine<br>(unification.c/h)<br>(Matches terms, generates subs.)]
    F --> G[Substitution Engine<br>(substitution.c/h)<br>(Applies bindings)]
    G --> H[Goal/Subgoal<br>Management]
    H --> D
```

## Building and Running

The directory includes a `Makefile` to simplify the compilation process. The primary executable generated is `prolog`.

To build the project:
```bash
make -C demo
```

To run the `prolog` executable (assuming you are in the root `code_agent` directory):
```bash
./demo/prolog < input.txt
```

Various `input_*.txt` files are provided for testing different scenarios, such as `input.txt`, `input_unification.txt`, `test_multi_arg.txt`, and `test_occurs_check.txt`.

## Purpose within `code_agent`

This `demo` project is intentionally complex enough to pose realistic challenges for an automated code agent. It allows for:

*   **File System Navigation**: Exploring a multi-file C project.
*   **Code Understanding**: Analyzing C code, data structures, and algorithms.
*   **Debugging**: Identifying and fixing logical errors or performance issues.
*   **Feature Implementation**: Adding new features or modifying existing logic within the inference engine.
*   **Testing**: Running the `prolog` executable with various inputs to verify changes.

By working with this demo, the `code_agent` can showcase its ability to interact with and modify a non-trivial codebase.
