#include "prolog_data.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// --- Memory Management and Constructors ---

Term* create_term(enum TermType type, const char *name) {
    Term *term = (Term*)malloc(sizeof(Term));
    if (!term) {
        perror("Failed to allocate Term");
        exit(EXIT_FAILURE);
    }
    term->type = type;
    term->name = strdup(name);
    if (!term->name) {
        perror("Failed to duplicate term name");
        free(term);
        exit(EXIT_FAILURE);
    }
    return term;
}

void free_term(Term *term) {
    if (!term) return;
    free(term->name);
    free(term);
}

Term* copy_term(Term *original_term) {
    if (!original_term) return NULL;
    Term *new_term = create_term(original_term->type, original_term->name);
    return new_term;
}

Predicate* create_predicate(const char *name, int arity) {
    Predicate *pred = (Predicate*)malloc(sizeof(Predicate));
    if (!pred) {
        perror("Failed to allocate Predicate");
        exit(EXIT_FAILURE);
    }
    pred->name = strdup(name);
    if (!pred->name) {
        perror("Failed to duplicate predicate name");
        free(pred);
        exit(EXIT_FAILURE);
    }
    pred->arity = arity;
    pred->args = (Term**)calloc(arity, sizeof(Term*)); // Initialize args to NULL
    if (!pred->args && arity > 0) { // calloc can return NULL for 0 size, which is fine.
        perror("Failed to allocate predicate arguments");
        free(pred->name);
        free(pred);
        exit(EXIT_FAILURE);
    }
    return pred;
}

void free_predicate(Predicate *pred) {
    if (!pred) return;
    free(pred->name);
    for (int i = 0; i < pred->arity; i++) {
        if (pred->args[i]) {
            free_term(pred->args[i]);
        }
    }
    free(pred->args);
    free(pred);
}

Predicate* copy_predicate(Predicate *original_pred) {
    if (!original_pred) return NULL;
    Predicate *new_pred = create_predicate(original_pred->name, original_pred->arity);
    for (int i = 0; i < original_pred->arity; i++) {
        new_pred->args[i] = copy_term(original_pred->args[i]);
    }
    return new_pred;
}

Clause* create_clause(Predicate *head) {
    Clause *clause = (Clause*)malloc(sizeof(Clause));
    if (!clause) {
        perror("Failed to allocate Clause");
        exit(EXIT_FAILURE);
    }
    clause->head = head;
    return clause;
}

void free_clause(Clause *clause) {
    if (!clause) return;
    free_predicate(clause->head);
    free(clause);
}

Substitution* create_substitution() {
    Substitution *sub = (Substitution*)malloc(sizeof(Substitution));
    if (!sub) {
        perror("Failed to allocate Substitution");
        exit(EXIT_FAILURE);
    }
    sub->size = 0;
    return sub;
}

void free_substitution(Substitution *sub) {
    if (!sub) return;
    for (int i = 0; i < sub->size; i++) {
        free(sub->bindings[i].variable_name);
        // Note: Term* in binding points to existing terms, not owned by binding
        // so we don't free them here to avoid double-freeing.
    }
    free(sub);
}

void add_binding(Substitution *sub, const char *var_name, Term *term) {
    if (!sub) return;
    if (sub->size < MAX_BINDINGS) {
        sub->bindings[sub->size].variable_name = strdup(var_name);
        if (!sub->bindings[sub->size].variable_name) {
            perror("Failed to duplicate variable name for binding");
            exit(EXIT_FAILURE);
        }
        sub->bindings[sub->size].term = term;
        sub->size++;
    } else {
        fprintf(stderr, "Substitution full! Cannot add binding for %s.\n", var_name);
    }
}

Term* get_binding(Substitution *sub, const char *var_name) {
    if (!sub) return NULL;
    for (int i = 0; i < sub->size; i++) {
        if (strcmp(sub->bindings[i].variable_name, var_name) == 0) {
            return sub->bindings[i].term;
        }
    }
    return NULL;
}
