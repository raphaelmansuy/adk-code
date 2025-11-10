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

bool unify(Term *t1, Term *t2, Substitution *sub) {
    // Cases for unification
    if (!t1 || !t2) return false; // Should not happen with well-formed terms

    if (t1->type == VARIABLE) {
        // Apply existing substitutions to t1
        for (int i = 0; i < sub->count; ++i) {
            if (strcmp(t1->name, sub->pairs[i].var_name) == 0) {
                return unify(sub->pairs[i].term, t2, sub);
            }
        }
        // Occurs check: if t1 occurs in t2, cannot unify
        if (occurs_check(t1->name, t2)) {
            return false;
        }
        // t1 is a fresh variable, add t1 = t2 to substitution
        add_sub_pair(sub, t1->name, t2);
        return true;
    }

    if (t2->type == VARIABLE) {
        // Apply existing substitutions to t2
        for (int i = 0; i < sub->count; ++i) {
            if (strcmp(t2->name, sub->pairs[i].var_name) == 0) {
                return unify(t1, sub->pairs[i].term, sub);
            }
        }
        // Occurs check: if t2 occurs in t1, cannot unify
        if (occurs_check(t2->name, t1)) {
            return false;
        }
        // t2 is a fresh variable, add t2 = t1 to substitution
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
