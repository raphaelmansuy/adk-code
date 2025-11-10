# ☕️ Prolog Interpreter Demo

Welcome to the Prolog Interpreter Demo! This project showcases a basic Prolog-like inference engine implemented in C. Prolog is a declarative logic programming language, where programs are expressed in terms of facts and rules, and computation proceeds by asking queries against this knowledge base. This C implementation is designed to be modular, extensible, and easy to understand, serving as a robust example for the `code_agent` to interact with.

## ✨ Features

- **Term Management**: Create, copy, and free atoms, variables, and compound terms.
- **Clause Handling**: Define and manage facts (simple clauses).
- **Knowledge Base**: Store and retrieve clauses efficiently.
- **Substitution**: Apply and manage variable substitutions during unification.
- **Parsing**: A simple parser for facts and queries.
- **Unification**: Core logic for pattern matching with occurs check.
- **Inference Engine**: Resolves basic queries against the knowledge base.
- **Interactive Mode**: Interact with the interpreter directly.
- **File Input**: Load facts and queries from a file.
- **Modular Design**: Code is split into logical units for clarity and maintainability.
- **Enhanced Makefile**: Provides convenient build, run, test, and clean operations with colorized output for improved user experience.

## 🏗️ Project Structure (ASCII Diagram)

The project is organized into modular components, making it easier to navigate and extend.

```text
.
├── Makefile
├── main.c              # Main application logic
├── parser.h            # Header for parsing functions
├── parser.c            # Source for parsing input
├── term.h              # Header for term data structures
├── term.c              # Source for term manipulation
├── clause.h            # Header for clause data structures
├── clause.c            # Source for clause manipulation
├── knowledge_base.h    # Header for knowledge base management
├── knowledge_base.c    # Source for knowledge base operations
├── substitution.h      # Header for substitution data structures
├── substitution.c      # Source for substitution operations
├── unification.h       # Header for unification algorithm
├── unification.c       # Source for unification logic
├── inference.h         # Header for inference engine
├── inference.c         # Source for inference logic
├── input.txt           # Example input file for facts and queries
└── input_unification.txt # Example input file for unification tests
```

## 🚀 How It Works (ASCII Flowchart)

The following diagram illustrates the high-level flow of how the Prolog interpreter processes input and resolves queries.

```text
+----------+     +--------+     +---------+
|  Start   |--->|  Input |--->|  Parse  |
|          |     | (Fact/ |     | (parser)|
|          |     | Query) |     |         |
+----------+     +--------+     +----+----+
                                     |
             +-----------------------+----------------------+
             |                                              |
             V (Fact)                                       V (Query)
      +------------+                                 +------------+
      | Add Fact   |                                 | Resolve    |
      | to KB      |                                 | Query      |
      | (kb_base)  |                                 | (inference)|
      +------------+                                 +-----+------+
             |                                              |
             V                                              V
+-----------------------+                         +---------------------+
| Ready for Next Input  |                         | Iterate KB Clauses  |
+-----------------------+                         | (inference)         |
             ^                                   +----------+----------+
             |                                              |
             +----------------------------------------------+
             |                                              V
             |                                   +---------------------+
             |                                   | Copy Clause &       |
             |                                   | Create Substitution |
             |                                   +----------+----------+
             |                                              |
             |                                              V
             |                                   +---------------------+
             |                                   | Unify Query Term    |
             |                                   | with Clause Head    |
             |                                   | (unification)       |
             |                                   +----------+----------+
             |                                              |
             |                    +-------------------------+-------------------------+
             |                    |                                                   |
             |                    V (Success)                                       V (Fail)
             |             +----------+----------+                         +--------------------+
             |             | Occurs Check        |                         | Backtrack /        |
             |             +----------+----------+                         | Next Clause        |
             |                    |                                         +--------------------+
             |                    V (Pass)                                       ^
             |             +----------+----------+                                   |
             |             | Apply Substitution  |-----------------------------------+
             |             | & Report Solution   |
             |             | (substitution)      |
             |             +---------------------+
             |                    |
             +--------------------+------------------------------------------+
                                  V
                                (Loop or End)
```

## 🛠️ Building the Project

The `Makefile` simplifies the build process. Navigate to the `demo` directory and use `make`.

```bash
cd demo
make
```

### Makefile Targets

- `make all`: Compiles all source files and links them into the `prolog` executable.
- `make run`: Executes the compiled `prolog` interpreter in interactive mode.
- `make test`: Runs predefined tests using `input.txt` and `input_unification.txt`.
- `make clean`: Removes all compiled object files (`.o`) and the `prolog` executable.
- `make help`: Displays a detailed help message with all available targets.

## 🏃 Running the Interpreter

### Interactive Mode

To run the interpreter in interactive mode:

```bash
cd demo
make run
```

You can then enter facts (e.g., `parent(john, mary).`) and queries (e.g., `?- parent(john, X).`). Type `exit.` to quit.

### File Input

To load facts and queries from a file:

```bash
cd demo
./prolog input.txt
```

Example `input.txt` content:
```prolog
parent(john, mary).
parent(mary, peter).
?- parent(john, X).
?- parent(X, Y).
?- parent(peter, Z).
```

## 🧪 Testing

The `make test` command will run the interpreter with two predefined input files:

```bash
cd demo
make test
```

This will demonstrate basic fact assertion and query resolution, including unification with variables and occurs check.

## 🔮 Future Enhancements

- **Rules**: Implement support for Prolog rules (e.g., `grandparent(X, Y) :- parent(X, Z), parent(Z, Y).`).
- **Backtracking**: Introduce a proper backtracking mechanism for finding all solutions to a query.
- **Operator Precedence**: Enhance the parser to handle more complex Prolog syntax with operator precedence.
- **Built-in Predicates**: Add support for common built-in predicates (e.g., `write/1`, arithmetic operations).
- **Error Handling**: More robust error reporting and recovery.
- **Variable Renaming**: Implement safe variable renaming for clauses during resolution to prevent clashes.

This project serves as a solid foundation for building a more complete Prolog interpreter. Contributions and suggestions are welcome!
