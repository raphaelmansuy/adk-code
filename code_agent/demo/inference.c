#include "inference.h"
#include <stdio.h>

bool resolve_query(KnowledgeBase *kb, Term *query_term) {
    Substitution *sub = create_substitution();
    bool result = false;

    for (int i = 0; i < kb->count; ++i) {
        Clause *clause = kb->clauses[i];
        // For now, only consider facts (clauses with no body)
        if (clause->body_len == 0) {
            // Create fresh copies of clause head to avoid modifying KB terms
            Term *fresh_head = copy_term(clause->head);
            
            Substitution *current_sub = create_substitution();
            if (unify(query_term, fresh_head, current_sub)) {
                result = true;

                // Print variable bindings if any
                if (current_sub->count > 0) {
                    printf("Yes. ");
                    for (int j = 0; j < current_sub->count; ++j) {
                        printf("%s = %s ", current_sub->pairs[j].var_name, current_sub->pairs[j].term->name);
                    }
                    printf("\n");
                } else {
                    printf("Yes.\n");
                }
                free_substitution(current_sub);
                free_term(fresh_head);
                break; // Found a solution, for now just one
            }
            free_substitution(current_sub);
            free_term(fresh_head);
        }
    }
    free_substitution(sub); // Free the initial (unused) substitution
    return result;
}
