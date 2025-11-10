#include "clause.h"
#include <stdlib.h>

Clause *create_clause(Term *head, Term **body, int body_len) {
    Clause *clause = (Clause *)malloc(sizeof(Clause));
    clause->head = head;
    clause->body = body;
    clause->body_len = body_len;
    return clause;
}

void free_clause(Clause *clause) {
    if (!clause) return;
    free_term(clause->head);
    for (int i = 0; i < clause->body_len; ++i) {
        free_term(clause->body[i]);
    }
    free(clause->body);
    free(clause);
}
