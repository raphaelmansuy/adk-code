#include "prolog_db.h"
#include <stdio.h>
#include <stdlib.h>

// Initialize global database variables
Clause *database[MAX_CLAUSES];
int db_size = 0;

void add_clause(Clause *clause) {
    if (db_size < MAX_CLAUSES) {
        database[db_size++] = clause;
        if (clause->type == FACT) {
            printf("Fact added: %s(", clause->content.fact->name);
        } else { // RULE
            printf("Rule added: %s(", clause->content.rule->head->name);
        }
        Predicate *display_pred = (clause->type == FACT) ? clause->content.fact : clause->content.rule->head;
        for (int i = 0; i < display_pred->arity; i++) {
            printf("%s%s", display_pred->args[i]->name, (i == display_pred->arity - 1) ? "" : ", ");
        }
        printf(").\n");
    } else {
        fprintf(stderr, "Database full!\n");
        free_clause(clause);
    }
}

void free_database() {
    for (int i = 0; i < db_size; i++) {
        free_clause(database[i]);
    }
    db_size = 0;
}
