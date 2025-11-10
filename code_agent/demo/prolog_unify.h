#ifndef PROLOG_UNIFY_H
#define PROLOG_UNIFY_H

#include "prolog_data.h"

bool unify_terms(Term *t1, Term *t2, Substitution *sub);
bool unify_predicates(Predicate *query, Predicate *fact_head, Substitution *sub);

#endif // PROLOG_UNIFY_H
