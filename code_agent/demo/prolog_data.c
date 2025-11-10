#include "prolog_data.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// --- Memory Management and Constructors ---

/**
 * @brief Creates a new Term object.
 * @param type The type of the term (ATOM or VARIABLE).
 * @param name The name of the term.
 * @return A pointer to the newly created Term, or exits on allocation failure.
 */
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

/**
 * @brief Frees a Term object and its associated memory.
 * @param term A pointer to the Term to free. If NULL, does nothing.
 */
void free_term(Term *term) {
    if (!term) return;
    free(term->name);
    free(term);
}

/**
 * @brief Creates a deep copy of a Term object.
 * @param original_term A pointer to the Term to copy. Can be NULL.
 * @return A pointer to the new, copied Term, or NULL if original_term is NULL.
 */
Term* copy_term(Term *original_term) {
    if (!original_term) return NULL;
    Term *new_term = create_term(original_term->type, original_term->name);
    return new_term;
}

/**
 * @brief Creates a new Predicate object.
 * @param name The name of the predicate.
 * @param arity The number of arguments the predicate takes.
 * @return A pointer to the newly created Predicate, or exits on allocation failure.
 */
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

/**
 * @brief Frees a Predicate object and its associated memory, including its arguments.
 * @param pred A pointer to the Predicate to free. If NULL, does nothing.
 */
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

/**
 * @brief Creates a deep copy of a Predicate object.
 * @param original_pred A pointer to the Predicate to copy. Can be NULL.
 * @return A pointer to the new, copied Predicate, or NULL if original_pred is NULL.
 */
Predicate* copy_predicate(Predicate *original_pred) {
    if (!original_pred) return NULL;
    Predicate *new_pred = create_predicate(original_pred->name, original_pred->arity);
    for (int i = 0; i < original_pred->arity; i++) {
        new_pred->args[i] = copy_term(original_pred->args[i]);
    }
    return new_pred;
}

/**
 * @brief Creates a new Clause object.
 * @param head A pointer to the head predicate of the clause.
 * @return A pointer to the newly created Clause, or exits on allocation failure.
 */
Clause* create_clause(Predicate *head) {
    Clause *clause = (Clause*)malloc(sizeof(Clause));
    if (!clause) {
        perror("Failed to allocate Clause");
        exit(EXIT_FAILURE);
    }
    clause->head = head;
    return clause;
}

/**
 * @brief Frees a Clause object and its associated memory.
 * @param clause A pointer to the Clause to free. If NULL, does nothing.
 */
void free_clause(Clause *clause) {
    if (!clause) return;
    free_predicate(clause->head);
    free(clause);
}

/**
 * @brief Creates a new Substitution object.
 * @return A pointer to the newly created Substitution, or exits on allocation failure.
 */
Substitution* create_substitution() {
    Substitution *sub = (Substitution*)malloc(sizeof(Substitution));
    if (!sub) {
        perror("Failed to allocate Substitution");
        exit(EXIT_FAILURE);
    }
    sub->size = 0;
    return sub;
}

/**
 * @brief Frees a Substitution object and its associated memory.
 *        Note: It frees the variable names in bindings but not the terms,
 *        as terms are not owned by the substitution to prevent double-freeing.
 * @param sub A pointer to the Substitution to free. If NULL, does nothing.
 */
void free_substitution(Substitution *sub) {
    if (!sub) return;
    for (int i = 0; i < sub->size; i++) {
        free(sub->bindings[i].variable_name);
        // Note: Term* in binding points to existing terms, not owned by binding
        // so we don't free them here to avoid double-freeing.
    }
    free(sub);
}

/**
 * @brief Adds a new binding to a Substitution.
 * @param sub A pointer to the Substitution to add the binding to.
 * @param var_name The name of the variable to bind.
 * @param term A pointer to the Term to bind the variable to.
 */
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

/**
 * @brief Retrieves the Term bound to a given variable name in a Substitution.
 * @param sub A pointer to the Substitution to search.
 * @param var_name The name of the variable to look up.
 * @return A pointer to the bound Term, or NULL if no binding is found.
 */
Term* get_binding(Substitution *sub, const char *var_name) {
    if (!sub) return NULL;
    for (int i = 0; i < sub->size; i++) {
        if (strcmp(sub->bindings[i].variable_name, var_name) == 0) {
            return sub->bindings[i].term;
        }
    }
    return NULL;
}
