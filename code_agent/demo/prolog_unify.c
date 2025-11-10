#include "prolog_unify.h"
#include <string.h>
#include <stdio.h>

// Checks if two terms unify and updates the substitution.
bool unify_terms(Term *t1, Term *t2, Substitution *sub) {
    // Resolve variables through substitution
    Term *val1 = (t1->type == VARIABLE) ? get_binding(sub, t1->name) : NULL;
    Term *val2 = (t2->type == VARIABLE) ? get_binding(sub, t2->name) : NULL;

    Term *resolved_t1 = val1 ? val1 : t1;
    Term *resolved_t2 = val2 ? val2 : t2;

    // If both are the same term (after resolving), they unify
    if (resolved_t1 == resolved_t2) {
        return true;
    }

    // Case 1: Both are atoms
    if (resolved_t1->type == ATOM && resolved_t2->type == ATOM) {
        return strcmp(resolved_t1->name, resolved_t2->name) == 0;
    }

    // Case 2: One is a variable, the other is a term
    if (resolved_t1->type == VARIABLE) {
        // Occurs check (simplified: avoid binding X to X itself, or X to a term containing X)
        // For this simple interpreter, we skip a full occurs check.
        add_binding(sub, resolved_t1->name, resolved_t2);
        return true;
    }
    if (resolved_t2->type == VARIABLE) {
        add_binding(sub, resolved_t2->name, resolved_t1);
        return true;
    }

    // Mismatch: e.g., an atom with a compound term (not implemented), or different atoms.
    return false;
}

// Checks if a query predicate unifies with a clause head and updates the substitution.
bool unify_predicates(Predicate *query, Predicate *fact_head, Substitution *sub) {
    if (strcmp(query->name, fact_head->name) != 0) {
        return false; // Different predicate names
    }
    if (query->arity != fact_head->arity) {
        return false; // Different arity
    }

    for (int i = 0; i < query->arity; i++) {
        if (!unify_terms(query->args[i], fact_head->args[i], sub)) {
            return false;
        }
    }
    return true;
}
