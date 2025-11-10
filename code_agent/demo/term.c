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

// Prints a term recursively
void print_term(Term *term) {
    if (!term) {
        printf("NULL");
        return;
    }

    printf("%s", term->name);
    if (term->type == COMPOUND && term->arity > 0) {
        printf("(");
        for (int i = 0; i < term->arity; ++i) {
            print_term(term->args[i]);
            if (i < term->arity - 1) {
                printf(", ");
            }
        }
        printf(")");
    }
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

// Deep copy a term and rename its variables with unique names
Term *rename_variables(Term *original, int *var_counter) {
    if (!original) return NULL;

    Term *new_term = (Term *)malloc(sizeof(Term));
    new_term->type = original->type;
    new_term->arity = original->arity;
    new_term->args = NULL;

    if (original->type == VARIABLE) {
        char var_name[256];
        snprintf(var_name, sizeof(var_name), "_G%d", (*var_counter)++);
        new_term->name = strdup(var_name);
    } else {
        new_term->name = strdup(original->name);
    }

    if (original->type == COMPOUND && original->arity > 0) {
        new_term->args = (Term **)calloc(original->arity, sizeof(Term *));
        for (int i = 0; i < original->arity; ++i) {
            new_term->args[i] = rename_variables(original->args[i], var_counter);
        }
    }
    return new_term;
}
