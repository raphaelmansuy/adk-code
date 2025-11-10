#include "inference.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// resolve is a recursive function that attempts to satisfy a list of goals
// (representing the current state of the query) given a knowledge base and a substitution.
// It implements a depth-first search with backtracking.
void resolve(KnowledgeBase *kb, Term **goals, int num_goals, Substitution *sub, int *var_counter, int *solution_count) {
    // Base case: If there are no more goals, a solution has been found.
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

    // The current goal to resolve is the first one in the list.
    Term *current_goal = goals[0];

    // Attempt to resolve the current goal by unifying it with the head of each clause
    // in the knowledge base (backward chaining).
    for (int i = 0; i < kb->count; ++i) {
        Clause *clause = kb->clauses[i];

        // To prevent variable clashes across different resolution paths, we create a
        // "fresh" copy of the clause by renaming all its variables. This is crucial
        // for correctness in logic programming.
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

        // Mark the current state of the substitution. If unification fails or a branch
        // of the search tree doesn't lead to a solution, we will restore the substitution
        // to this marked state (backtracking).
        int sub_mark = mark_substitution(sub);

        // Attempt to unify the current goal with the head of the fresh clause.
        // If unification succeeds, `sub` will be updated with new bindings.
        if (unify(current_goal, fresh_clause->head, sub)) {
            // Unification successful. Construct the new list of goals:
            // The body of the fresh clause becomes new goals, followed by the remaining
            // goals from the original list (excluding the current_goal that was just resolved).
            int next_num_goals = (fresh_clause->body_len) + (num_goals - 1);
            Term **next_goals = (Term **)calloc(next_num_goals, sizeof(Term *));
            if (!next_goals) {
                fprintf(stderr, "Memory allocation failed for next_goals.\n");
                // Proper error handling/exit here
                // For now, we'll just clean up and return
                restore_substitution(sub, sub_mark);
                free_term(fresh_clause->head);
                if (fresh_clause->body) {
                    for (int j = 0; j < fresh_clause->body_len; ++j) free_term(fresh_clause->body[j]);
                    free(fresh_clause->body);
                }
                free(fresh_clause);
                return;
            }

            // Add the body of the fresh clause to the front of the goals.
            // Apply the current substitution to these new goals.
            for (int j = 0; j < fresh_clause->body_len; ++j) {
                next_goals[j] = apply_substitution(fresh_clause->body[j], sub);
            }
            // Add the remaining goals from the original list.
            // Apply the current substitution to these remaining goals as well.
            for (int j = 1; j < num_goals; ++j) {
                next_goals[fresh_clause->body_len + (j - 1)] = apply_substitution(goals[j], sub);
            }

            // Recursively call resolve with the new set of goals.
            resolve(kb, next_goals, next_num_goals, sub, var_counter, solution_count);

            // After the recursive call returns (either a solution was found or the branch
            // was exhausted), free the memory allocated for the next_goals terms.
            // These terms were created by `apply_substitution`.
            for (int j = 0; j < next_num_goals; ++j) {
                free_term(next_goals[j]);
            }
            free(next_goals);
        }
        // Backtracking: Restore the substitution to its state before attempting
        // unification with the current clause. This allows exploration of other
        // clauses with a clean substitution environment.
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

// resolve_query is the entry point for starting the resolution process for a given query.
// It initializes the substitution, variable counter, and calls the main `resolve` function.
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
