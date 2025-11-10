#include "substitution.h"
#include <stdlib.h>
#include <string.h>

Substitution *create_substitution() {
    Substitution *sub = (Substitution *)malloc(sizeof(Substitution));
    sub->count = 0;
    sub->capacity = 5;
    sub->pairs = (SubPair *)malloc(sub->capacity * sizeof(SubPair));
    return sub;
}

void add_sub_pair(Substitution *sub, const char *var_name, Term *term) {
    if (sub->count == sub->capacity) {
        sub->capacity *= 2;
        sub->pairs = (SubPair *)realloc(sub->pairs, sub->capacity * sizeof(SubPair));
    }
    sub->pairs[sub->count].var_name = strdup(var_name);
    sub->pairs[sub->count].term = copy_term(term); // Copy the term
    sub->count++;
}

// Apply a substitution to a term
Term *apply_substitution(Term *term, Substitution *sub) {
    if (!term) return NULL;

    if (term->type == VARIABLE) {
        for (int i = 0; i < sub->count; ++i) {
            if (strcmp(term->name, sub->pairs[i].var_name) == 0) {
                // Found a substitution for this variable.
                // Recursively apply substitution to the term it's bound to.
                Term *substituted_term = apply_substitution(sub->pairs[i].term, sub);
                return substituted_term;
            }
        }
    }

    if (term->type == COMPOUND) {
        Term *new_term = create_compound_term(term->name, term->arity);
        for (int i = 0; i < term->arity; ++i) {
            new_term->args[i] = apply_substitution(term->args[i], sub);
        }
        return new_term;
    }
    return copy_term(term); // Atoms and non-substituted variables are copied directly
}

void free_substitution(Substitution *sub) {
    if (!sub) return;
    for (int i = 0; i < sub->count; ++i) {
        free(sub->pairs[i].var_name);
        free_term(sub->pairs[i].term);
    }
    free(sub->pairs);
    free(sub);
}

// Mark the current state of the substitution list for backtracking
int mark_substitution(Substitution *sub) {
    return sub->count;
}

// Restore the substitution list to a previously marked state
void restore_substitution(Substitution *sub, int mark) {
    for (int i = sub->count - 1; i >= mark; --i) {
        free(sub->pairs[i].var_name);
        free_term(sub->pairs[i].term);
    }
    sub->count = mark;
}
