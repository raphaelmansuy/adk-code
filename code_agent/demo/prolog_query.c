#include "prolog_query.h"
#include <stdio.h>
#include <stdlib.h>

// Helper to print a substitution
void print_substitution(Substitution *sub) {
    if (sub->size == 0) {
        printf("  Yes.\n");
        return;
    }
    printf("  Yes, with bindings:\n");
    for (int i = 0; i < sub->size; i++) {
        printf("    %s = ", sub->bindings[i].variable_name);
        if (sub->bindings[i].term->type == ATOM) {
            printf("%s\n", sub->bindings[i].term->name);
        } else { // VARIABLE
            // This case is more complex in a full interpreter (e.g., X = Y)
            printf("%s (variable, not fully resolved here)\n", sub->bindings[i].term->name);
        }
    }
}

void query(Predicate *query_pred) {
    printf("Query: %s(", query_pred->name);
    for (int i = 0; i < query_pred->arity; i++) {
        printf("%s%s", query_pred->args[i]->name, (i == query_pred->arity - 1) ? "" : ", ");
    }
    printf(")?\n");

    bool any_found = false;
    for (int i = 0; i < db_size; i++) {
        Clause *fact_clause = database[i];

        // Create a fresh substitution for each attempt
        Substitution *sub = create_substitution();

        // Create a copy of the fact head to rename variables (if any) and avoid modifying database
        Predicate *fact_head_copy = copy_predicate(fact_clause->head);

        if (unify_predicates(query_pred, fact_head_copy, sub)) {
            print_substitution(sub);
            any_found = true;
        }

        // Clean up the copied fact and the substitution
        free_predicate(fact_head_copy);
        free_substitution(sub);
    }

    if (!any_found) {
        printf("  No.\n");
    }
}
