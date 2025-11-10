# Plan for Improving Prolog Interpreter in C

This document outlines a plan to enhance the Prolog interpreter implemented in C, based on the existing codebase and identified areas for improvement.

## I. Major Functional Enhancements

These are the core features that will significantly improve the interpreter's capabilities, as highlighted in the `README.md` and confirmed by code analysis.

### 1. Implement Full Support for Prolog Rules

**Current State:** The parser (`parser.c`) can parse rules (clauses with a body using `:-`), but the inference engine (`inference.c`) currently only handles simple facts effectively. The `resolve` function in `inference.c` has a basic structure for handling clause bodies but needs to be fully integrated.

**Improvements:**
*   **Goal Management:** Enhance the `resolve` function to properly manage a stack of goals. When a rule is used to resolve a goal, its body goals must be added to the current goal list for subsequent resolution.
*   **Backtracking for Rules:** Ensure that backtracking correctly explores alternative rules and different clauses for each subgoal.
*   **Variable Scope:** Carefully manage variable scopes when applying rules to prevent unintended bindings across different rule invocations.

### 2. Robust Backtracking Mechanism

**Current State:** The `inference.c` has a rudimentary backtracking mechanism by copying substitutions. However, it currently stops after finding the first solution.

**Improvements:**
*   **Explore All Solutions:** Modify the `resolve` function to continue searching for solutions after one is found, allowing the user to request more solutions (e.g., by typing `;` in interactive mode).
*   **Choice Points:** Introduce explicit "choice points" to manage the state of the search, allowing the interpreter to return to previous states and explore alternative paths.
*   **Stack-based Backtracking:** Implement a more formal stack-based approach for managing choice points and goal lists, similar to how Prolog engines typically work.

### 3. Enhanced Variable Renaming for Freshness

**Current State:** The `rename_variables` function in `term.c` generates unique variable names (e.g., `_G1`, `_G2`). This is a good start for preventing variable clashes.

**Improvements:**
*   **Consistent Application:** Ensure `rename_variables` is consistently applied to all clauses fetched from the knowledge base during resolution to guarantee variable freshness. This is already largely in place in `inference.c`, but a review is warranted.
*   **Optimization:** For very large knowledge bases, consider if the current renaming strategy has performance implications, though for a basic interpreter, it should be sufficient.

## II. Minor Enhancements and Code Quality

These improvements focus on robustness, user experience, and maintainability.

### 4. Improved Error Handling and Reporting

**Current State:** Error messages are basic and often printed to `stderr`.

**Improvements:**
*   **Specific Error Messages:** Provide more descriptive error messages, indicating the exact location (e.g., line number, character offset) of parsing or logical errors.
*   **Error Recovery:** Implement basic error recovery strategies for the parser to continue processing input after a minor error, if possible.
*   **Structured Error Codes:** Consider using error codes or a more structured error reporting mechanism.

### 5. Comprehensive Memory Management Review

**Current State:** The codebase uses `malloc`, `realloc`, and `free` extensively.

**Improvements:**
*   **Valgrind Analysis:** Perform a thorough memory leak and error analysis using `Valgrind` or similar tools to identify and fix any leaks, double frees, or invalid memory accesses.
*   **Ownership Clarity:** Clearly define memory ownership rules for data structures (e.g., who is responsible for freeing a `Term` or `Substitution`).
*   **Resource Cleanup:** Ensure all allocated resources are freed on program exit and during intermediate operations (e.g., when a branch of inference fails).

### 6. Built-in Predicates

**Current State:** No built-in predicates are supported.

**Improvements:**
*   **Basic I/O:** Add `write/1` for printing terms.
*   **Arithmetic Operations:** Implement basic arithmetic operations (e.g., `is/2`, `+/2`, `-/2`).
*   **Comparison Predicates:** Support comparison operators (e.g., `==/2`, `>/2`).

### 7. Enhanced Parser with Operator Precedence

**Current State:** The parser is a simple recursive descent parser without explicit operator precedence handling.

**Improvements:**
*   **Operator Table:** Implement an operator precedence parser to correctly handle infix, prefix, and postfix operators (e.g., `+`, `-`, `*`, `/`, `is`).
*   **Extended Syntax:** Support more complex Prolog syntax elements as needed (e.g., lists, cuts).

### 8. Refined Output Formatting

**Current State:** Output for query results can be somewhat simplistic (e.g., "No direct bindings.").

**Improvements:**
*   **Clearer Bindings:** Format variable bindings more consistently and readably.
*   **No Solution Message:** Ensure a clear "No." is printed when no solutions are found, and avoid redundant messages.
*   **Interactive Prompt:** Improve the interactive prompt for a better user experience.

## III. Testing and Verification

*   **Unit Tests:** Expand existing tests (if any) and add unit tests for individual modules (parser, unification, substitution, knowledge base).
*   **Integration Tests:** Create more comprehensive integration tests that cover new features like rules and backtracking.
*   **Regression Tests:** Ensure that new features or bug fixes do not introduce regressions in existing functionality.

This plan provides a roadmap for evolving the basic Prolog interpreter into a more feature-rich and robust system.