#include "term.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

Term *create_term(TermType type, const char *name) {
    Term *term = (Term *)malloc(sizeof(Term));
    term->type = type;
    term->name = strdup(name);
    term->args = NULL;
    term->arity = 0;
    return term;
}

Term *create_compound_term(const char *name, int arity) {
    Term *term = create_term(COMPOUND, name);
    term->arity = arity;
    term->args = (Term **)calloc(arity, sizeof(Term *));
    return term;
}

// Deep copy a term
Term *copy_term(Term *original) {
    if (!original) return NULL;
    Term *new_term = create_term(original->type, original->name);
    new_term->arity = original->arity;
    if (original->type == COMPOUND && original->arity > 0) {
        new_term->args = (Term **)calloc(original->arity, sizeof(Term *));
        for (int i = 0; i < original->arity; ++i) {
            new_term->args[i] = copy_term(original->args[i]);
        }
    }
    return new_term;
}


void free_term(Term *term) {
    if (!term) return;
    free(term->name);
    if (term->type == COMPOUND) {
        for (int i = 0; i < term->arity; ++i) {
            free_term(term->args[i]);
        }
        free(term->args);
    }
    free(term);
}
