#ifndef PROLOG_DATA_H
#define PROLOG_DATA_H

#include <stdbool.h>

// --- Data Structures ---

// Represents a Prolog term (atom or variable)
enum TermType { ATOM, VARIABLE };

typedef struct Term {
    enum TermType type;
    char *name;
} Term;

// Represents a Prolog predicate (e.g., parent(X, Y))
typedef struct Predicate {
    char *name;
    Term **args;
    int arity; // Number of arguments
} Predicate;

// Represents a Prolog clause (fact or rule)
// For simplicity, we'll only handle facts initially
typedef struct Clause {
    Predicate *head;
    // Predicate **body; // For rules, not implemented yet
    // int body_len;
} Clause;

// Represents a variable binding in a substitution
typedef struct Binding {
    char *variable_name;
    Term *term;
} Binding;

// Represents a substitution (a list of variable bindings)
#define MAX_BINDINGS 50
typedef struct Substitution {
    Binding bindings[MAX_BINDINGS];
    int size;
} Substitution;

// --- Memory Management and Constructors ---

Term* create_term(enum TermType type, const char *name);
void free_term(Term *term);
Term* copy_term(Term *original_term);

Predicate* create_predicate(const char *name, int arity);
void free_predicate(Predicate *pred);
Predicate* copy_predicate(Predicate *original_pred);

Clause* create_clause(Predicate *head);
void free_clause(Clause *clause);

Substitution* create_substitution();
void free_substitution(Substitution *sub);
void add_binding(Substitution *sub, const char *var_name, Term *term);
Term* get_binding(Substitution *sub, const char *var_name);

#endif // PROLOG_DATA_H
