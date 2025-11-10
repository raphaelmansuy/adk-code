# Simple Prolog Interpreter (C) Demo

This directory contains a simple Prolog interpreter implemented in C. It demonstrates basic Prolog fact loading, storage, and querying capabilities.

## Files

*   `Makefile`: Used to build the project.
*   `main.c`: The main application logic, including interactive query mode.
*   `facts.pl`: A sample Prolog file containing facts (e.g., parent, male, female relationships).
*   `prolog_data.h`, `prolog_data.c`: Data structures for Prolog terms, predicates, and clauses.
*   `prolog_db.h`, `prolog_db.c`: Functions for managing the Prolog fact database.
*   `prolog_parser.h`, `prolog_parser.c`: Handles parsing Prolog queries from strings.
*   `prolog_query.h`, `prolog_query.c`: Implements the Prolog query evaluation logic.
*   `prolog_unify.h`, `prolog_unify.c`: Implements the unification algorithm.

## How to Build

To build the executable, navigate to this directory and run `make`:

```bash
cd demo
make
```

This will compile the C source files and create an executable named `my_prolog_exe`.

## How to Run

### Using Default Facts

If you run the executable without any arguments, it will load a set of default facts (defined in `main.c`) and enter an interactive query mode:

```bash
./my_prolog_exe
```

### Loading Facts from a File

You can also provide a `.pl` file (like `facts.pl`) as a command-line argument to load facts from it:

```bash
./my_prolog_exe facts.pl
```

### Interactive Query Mode

Once in interactive query mode, you can type Prolog-like queries. For example, using `facts.pl`:

```prolog
?- parent(pam, X).
X = bob
Yes
?- parent(X, bob).
X = pam
X = tom
Yes
?- male(jim).
Yes
?- male(X).
X = tom
X = bob
X = jim
Yes
?- exit.
```

Type `exit.` or `exit` to quit the interpreter.

## Cleaning Up

To remove compiled object files and the executable:

```bash
make clean
```
