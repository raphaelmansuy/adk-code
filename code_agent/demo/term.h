#ifndef TERM_H
#define TERM_H

#include <stdbool.h>

// Types of terms
typedef enum {
    ATOM,
    VARIABLE,
    COMPOUND
} TermType;

// A term can be an atom, a variable, or a compound term (predicate)
typedef struct Term {
    TermType type;
    char *name; // For ATOM, VARIABLE, and COMPOUND (predicate name)
    struct Term **args; // For COMPOUND terms
    int arity;          // Number of arguments for COMPOUND terms
} Term;

// Function prototypes for term management
Term *create_term(TermType type, const char *name);
Term *create_compound_term(const char *name, int arity);
Term *copy_term(Term *original);
void free_term(Term *term);

#endif // TERM_H
