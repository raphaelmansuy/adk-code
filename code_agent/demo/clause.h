#ifndef CLAUSE_H
#define CLAUSE_H

#include "term.h"

// A clause (fact or rule)
typedef struct Clause {
    Term *head;
    Term **body; // Array of terms (goals) for the body
    int body_len;
} Clause;

// Function prototypes for clause management
Clause *create_clause(Term *head, Term **body, int body_len);
void free_clause(Clause *clause);

#endif // CLAUSE_H
