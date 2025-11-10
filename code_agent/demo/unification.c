#include "unification.h"
#include <string.h>

// Helper for occurs check: returns true if var_name occurs in term
bool occurs_check(const char *var_name, Term *term) {
    if (!term) return false;

    if (term->type == VARIABLE) {
        return strcmp(var_name, term->name) == 0;
    }

    if (term->type == COMPOUND) {
        for (int i = 0; i < term->arity; ++i) {
            if (occurs_check(var_name, term->args[i])) {
                return true;
            }
        }
    }
    return false;
}

// Helper to dereference a term: returns the term a variable is bound to, or the term itself
Term *dereference(Term *term, Substitution *sub) {
    if (!term || term->type != VARIABLE) {
        return term;
    }
    for (int i = 0; i < sub->count; ++i) {
        if (strcmp(term->name, sub->pairs[i].var_name) == 0) {
            return dereference(sub->pairs[i].term, sub);
        }
    }
    return term;
}

bool unify(Term *t1, Term *t2, Substitution *sub) {
    // Dereference terms to their most concrete form
    t1 = dereference(t1, sub);
    t2 = dereference(t2, sub);

    // Cases for unification
    if (!t1 || !t2) return false; // Should not happen with well-formed terms

    if (t1->type == VARIABLE) {
        if (strcmp(t1->name, t2->name) == 0 && t2->type == VARIABLE) {
            return true; // Same variable
        }
        if (occurs_check(t1->name, t2)) {
            return false; // Occurs check fails
        }
        add_sub_pair(sub, t1->name, t2);
        return true;
    }

    if (t2->type == VARIABLE) {
        // No need for occurs check on t2->name in t1 if t1 is an ATOM or COMPOUND
        // If t1 is a VARIABLE, it's handled by the previous if block
        if (occurs_check(t2->name, t1)) {
            return false; // Occurs check fails
        }
        add_sub_pair(sub, t2->name, t1);
        return true;
    }

    if (t1->type == ATOM && t2->type == ATOM) {
        return strcmp(t1->name, t2->name) == 0;
    }

    if (t1->type == COMPOUND && t2->type == COMPOUND) {
        if (strcmp(t1->name, t2->name) != 0 || t1->arity != t2->arity) {
            return false;
        }
        for (int i = 0; i < t1->arity; ++i) {
            if (!unify(t1->args[i], t2->args[i], sub)) {
                return false;
            }
        }
        return true;
    }

    return false; // Mismatching types (e.g., atom and compound)
}
