#ifndef SUBSTITUTION_H
#define SUBSTITUTION_H

#include "term.h"

// Substitution pair: variable -> term
typedef struct SubPair {
    char *var_name;
    Term *term;
} SubPair;

// Substitution list
typedef struct Substitution {
    SubPair *pairs;
    int count;
    int capacity;
} Substitution;

// Function prototypes for substitution management
Substitution *create_substitution();
void add_sub_pair(Substitution *sub, const char *var_name, Term *term);
Term *apply_substitution(Term *term, Substitution *sub);
void free_substitution(Substitution *sub);
int mark_substitution(Substitution *sub);
void restore_substitution(Substitution *sub, int mark);

#endif // SUBSTITUTION_H
