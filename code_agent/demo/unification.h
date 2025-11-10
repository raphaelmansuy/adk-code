#ifndef UNIFICATION_H
#define UNIFICATION_H

#include "term.h"
#include "substitution.h"
#include <stdbool.h>

// Function prototypes for unification
bool occurs_check(const char *var_name, Term *term);
bool unify(Term *t1, Term *t2, Substitution *sub);

#endif // UNIFICATION_H
