#include "prolog_query.h"
#include <stdio.h>
#include <stdlib.h>

// Helper to generate a unique variable name for renaming
static int var_counter = 0;
static char* generate_fresh_var_name(const char *original_name) {
    char *new_name = (char*)malloc(strlen(original_name) + 10); // +10 for counter
    if (!new_name) {
        perror("Failed to allocate fresh var name");
        exit(EXIT_FAILURE);
    }
    sprintf(new_name, "%s_%d", original_name, var_counter++);
    return new_name;
}

// Helper to rename variables in a Term
static Term* rename_variables_in_term(Term *term, Substitution *renaming_sub) {
    if (!term || term->type == ATOM) return copy_term(term);

    // If it's a variable, check if it's already renamed in this context
    Term *renamed_term = get_binding(renaming_sub, term->name);
    if (renamed_term) {
        return copy_term(renamed_term);
    } else {
        char *fresh_name = generate_fresh_var_name(term->name);
        Term *new_var_term = create_term(VARIABLE, fresh_name);
        add_binding(renaming_sub, term->name, new_var_term);
        free(fresh_name); // The name is now owned by new_var_term and renaming_sub
        return new_var_term;
    }
}

// Helper to rename variables in a Predicate
static Predicate* rename_variables_in_predicate(Predicate *pred, Substitution *renaming_sub) {
    if (!pred) return NULL;
    Predicate *new_pred = create_predicate(pred->name, pred->arity);
    for (int i = 0; i < pred->arity; i++) {
        new_pred->args[i] = rename_variables_in_term(pred->args[i], renaming_sub);
    }
    return new_pred;
}

// Helper to rename variables in a PredicateList
static PredicateList* rename_variables_in_predicatelist(PredicateList *list, Substitution *renaming_sub) {
    if (!list) return NULL;
    PredicateList *new_list = create_predicatelist(list->count);
    for (int i = 0; i < list->count; i++) {
        new_list->predicates[i] = rename_variables_in_predicate(list->predicates[i], renaming_sub);
    }
    return new_list;
}

// Helper to rename variables in a Clause (fact or rule)
static Clause* rename_variables_in_clause(Clause *clause) {
    if (!clause) return NULL;
    Substitution *renaming_sub = create_substitution();
    Clause *new_clause = (Clause*)malloc(sizeof(Clause));
    if (!new_clause) {
        perror("Failed to allocate new clause for renaming");
        exit(EXIT_FAILURE);
    }
    new_clause->type = clause->type;

    if (clause->type == FACT) {
        new_clause->content.fact = rename_variables_in_predicate(clause->content.fact, renaming_sub);
    } else { // RULE
        Rule *original_rule = clause->content.rule;
        Predicate *new_head = rename_variables_in_predicate(original_rule->head, renaming_sub);
        PredicateList *new_body = rename_variables_in_predicatelist(original_rule->body, renaming_sub);
        new_clause->content.rule = create_rule(new_head, new_body);
    }
    free_substitution(renaming_sub);
    return new_clause;
}

// Global file pointer for query results
FILE *query_output_file = NULL;

// Helper to print a substitution to the output file
void print_substitution(Substitution *sub) {
    if (!query_output_file) return;

    if (sub->size == 0) {
        fprintf(query_output_file, "  Yes.\n");
        return;
    }
    fprintf(query_output_file, "  Yes, with bindings:\n");
    for (int i = 0; i < sub->size; i++) {
        fprintf(query_output_file, "    %s = ", sub->bindings[i].variable_name);
        if (sub->bindings[i].term->type == ATOM) {
            fprintf(query_output_file, "%s\n", sub->bindings[i].term->name);
        } else { // VARIABLE
            fprintf(query_output_file, "%s\n", sub->bindings[i].term->name); // Print the name of the bound variable
        }
    }
}

bool prove(PredicateList *goals, Substitution *sub) {
    // Base Case: If there are no goals, we have successfully proven the query
    if (goals->count == 0) {
        return true;
    }

    // Get the first goal to prove
    Predicate *current_goal = goals->predicates[0];

    // Try to find a clause in the database that matches the current goal
    for (int i = 0; i < db_size; i++) {
        Clause *original_clause = database[i];

        // Create a fresh copy of the clause to rename its variables
        Clause *fresh_clause = rename_variables_in_clause(original_clause);

        Substitution *new_sub = copy_substitution(sub); // Create a copy of current substitution for this path

        Predicate *clause_head = NULL;
        if (fresh_clause->type == FACT) {
            clause_head = fresh_clause->content.fact;
        } else { // RULE
            clause_head = fresh_clause->content.rule->head;
        }

        // Attempt to unify the current goal with the head of the clause
        if (unify_predicates(current_goal, clause_head, new_sub)) {
            // If unification succeeds, we have a new set of goals to prove
            PredicateList *remaining_goals = create_predicatelist(0);

            // Add the body of the rule (if it's a rule) to the front of the remaining goals
            if (fresh_clause->type == RULE) {
                Rule *rule = fresh_clause->content.rule;
                for (int j = 0; j < rule->body->count; j++) {
                    remaining_goals->count++;
                    remaining_goals->predicates = (Predicate**)realloc(remaining_goals->predicates, remaining_goals->count * sizeof(Predicate*));
                    if (!remaining_goals->predicates) { perror("realloc failed"); exit(EXIT_FAILURE); }
                    remaining_goals->predicates[remaining_goals->count - 1] = copy_predicate(rule->body->predicates[j]);
                }
            }

            // Add the rest of the original goals (after the current_goal) to the new goals list
            for (int j = 1; j < goals->count; j++) {
                remaining_goals->count++;
                remaining_goals->predicates = (Predicate**)realloc(remaining_goals->predicates, remaining_goals->count * sizeof(Predicate*));
                if (!remaining_goals->predicates) { perror("realloc failed"); exit(EXIT_FAILURE); }
                remaining_goals->predicates[remaining_goals->count - 1] = copy_predicate(goals->predicates[j]);
            }

            // Recursively try to prove the remaining goals with the new substitution
            if (prove(remaining_goals, new_sub)) {
                // If successful, propagate the substitution back up
                // We need to merge new_sub back into the original sub (or handle it in calling context)
                // For simplicity here, we'll just copy final bindings. More robust would be to pass sub by reference.
                free_substitution(sub); // Free the old substitution
                sub->size = new_sub->size;
                for(int k = 0; k < new_sub->size; k++) {
                    sub->bindings[k].variable_name = strdup(new_sub->bindings[k].variable_name);
                    sub->bindings[k].term = copy_term(new_sub->bindings[k].term);
                }
                free_predicatelist(remaining_goals);
                free_clause(fresh_clause);
                free_substitution(new_sub);
                return true;
            }
            free_predicatelist(remaining_goals);
        }
        free_clause(fresh_clause);
        free_substitution(new_sub);
    }

    // If no clause leads to a proof, backtrack
    return false;
}

void query(PredicateList *query_goals) {
    query_output_file = fopen("query_results.txt", "w");
    if (!query_output_file) {
        perror("Failed to open query_results.txt");
        return;
    }

    fprintf(query_output_file, "Query: ");
    for (int i = 0; i < query_goals->count; i++) {
        Predicate *p = query_goals->predicates[i];
        fprintf(query_output_file, "%s(", p->name);
        for (int j = 0; j < p->arity; j++) {
            fprintf(query_output_file, "%s%s", p->args[j]->name, (j == p->arity - 1) ? "" : ", ");
        }
        fprintf(query_output_file, ")%s", (i == query_goals->count - 1) ? "" : ", ");
    }
    fprintf(query_output_file, ")?\n");

    Substitution *initial_sub = create_substitution();
    if (prove(query_goals, initial_sub)) {
        print_substitution(initial_sub);
    } else {
        fprintf(query_output_file, "  No.\n");
    }
    free_substitution(initial_sub);

    fclose(query_output_file);
    query_output_file = NULL; // Reset file pointer
}
