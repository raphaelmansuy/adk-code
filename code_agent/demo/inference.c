#include "inference.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// Helper to copy a substitution (for backtracking)
Substitution *copy_substitution(Substitution *original) {
    Substitution *new_sub = create_substitution();
    for (int i = 0; i < original->count; ++i) {
        add_sub_pair(new_sub, original->pairs[i].var_name, original->pairs[i].term);
    }
    return new_sub;
}

// Recursive resolution function for backward chaining
void resolve(KnowledgeBase *kb, Term **goals, int num_goals, Substitution *sub, int *var_counter, int *solution_count) {
    if (num_goals == 0) {
        // All goals satisfied, a solution is found
        printf("Yes. ");
        bool has_bindings = false;
        for (int i = 0; i < sub->count; ++i) {
            // Only print bindings for original query variables (those not starting with _G)
            if (sub->pairs[i].var_name[0] != '_' || sub->pairs[i].var_name[1] != 'G') {
                printf("%s = ", sub->pairs[i].var_name);
                // Apply substitution to the bound term to get its most concrete form
                Term *bound_term = apply_substitution(sub->pairs[i].term, sub);
                if (bound_term) {
                    print_term(bound_term);
                    printf(" "); 
                    free_term(bound_term); // Free the copied term
                } else {
                    printf("NULL ");
                }
                has_bindings = true;
            }
        }
        if (!has_bindings) {
            printf("No direct bindings.");
        }
        printf("\n");
        (*solution_count)++; // Increment solution count
        return;
    }

    Term *current_goal = goals[0];
    // Iterate through all clauses in the knowledge base to find a match for the current_goal
    for (int i = 0; i < kb->count; ++i) {
        Clause *clause = kb->clauses[i];

        // Create a fresh copy of the clause with renamed variables
        Clause *fresh_clause = (Clause *)malloc(sizeof(Clause));
        fresh_clause->head = rename_variables(clause->head, var_counter);
        fresh_clause->body_len = clause->body_len;
        fresh_clause->body = NULL;
        if (clause->body_len > 0) {
            fresh_clause->body = (Term **)calloc(clause->body_len, sizeof(Term *));
            for (int j = 0; j < clause->body_len; ++j) {
                fresh_clause->body[j] = rename_variables(clause->body[j], var_counter);
            }
        }

        // Mark the current state of the substitution for backtracking
        int sub_mark = mark_substitution(sub);

        if (unify(current_goal, fresh_clause->head, sub)) {
            // Unification successful, now build the next set of goals
            int next_num_goals = (fresh_clause->body_len) + (num_goals - 1);
            Term **next_goals = (Term **)calloc(next_num_goals, sizeof(Term *));

            // Add the body of the fresh clause to the front of the goals
            for (int j = 0; j < fresh_clause->body_len; ++j) {
                next_goals[j] = apply_substitution(fresh_clause->body[j], sub);
            }
            // Add the remaining goals
            for (int j = 1; j < num_goals; ++j) {
                next_goals[fresh_clause->body_len + (j - 1)] = apply_substitution(goals[j], sub);
            }

            // Recursive call
            resolve(kb, next_goals, next_num_goals, sub, var_counter, solution_count);
            // Continue exploring other branches, don't stop after the first solution

            // Free next_goals terms (they were copied by apply_substitution)
            for (int j = 0; j < next_num_goals; ++j) {
                free_term(next_goals[j]);
            }
            free(next_goals);
        }
        // Restore substitution for the next clause attempt
        restore_substitution(sub, sub_mark);

        // Clean up fresh clause
        free_term(fresh_clause->head);
        if (fresh_clause->body) {
            for (int j = 0; j < fresh_clause->body_len; ++j) {
                free_term(fresh_clause->body[j]);
            }
            free(fresh_clause->body);
        }
        free(fresh_clause);
    }
}

// Entry point for resolving a query
void resolve_query(KnowledgeBase *kb, Term *query_term, int *solution_count) {
    Substitution *initial_sub = create_substitution();
    int var_counter = 0;
    Term *goals[1];
    goals[0] = query_term; // The initial query is our first goal

    printf("Query: ");
    if (query_term) {
        print_term(query_term);
    }
    printf("\n");

    *solution_count = 0; // Initialize solution count for this query
    resolve(kb, goals, 1, initial_sub, &var_counter, solution_count);

    if (*solution_count == 0) {
        printf("No.\n");
    }

    free_substitution(initial_sub);
    // No explicit return needed for void function
}
