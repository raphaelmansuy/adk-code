#ifndef PROLOG_QUERY_H
#define PROLOG_QUERY_H

#include "prolog_data.h"
#include "prolog_db.h"
#include "prolog_unify.h"

// Main function to execute a query (a list of goals)
void query(PredicateList *query_goals);

// Recursive function to attempt to prove a list of goals
// Returns true if proof is found, false otherwise.
// It modifies the substitution in place.
bool prove(PredicateList *goals, Substitution *sub);

#endif // PROLOG_QUERY_H
