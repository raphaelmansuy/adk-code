#include "prolog_db.h"
#include <stdio.h>
#include <stdlib.h>

// Initialize global database variables
Clause *database[MAX_CLAUSES];
int db_size = 0;

void add_clause(Clause *clause) {
    if (db_size < MAX_CLAUSES) {
        database[db_size++] = clause;
        printf("Fact added: %s(", clause->head->name);
        for (int i = 0; i < clause->head->arity; i++) {
            printf("%s%s", clause->head->args[i]->name, (i == clause->head->arity - 1) ? "" : ", ");
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
