#ifndef PROLOG_DB_H
#define PROLOG_DB_H

#include "prolog_data.h"

// A simple database to store Prolog clauses (facts and rules)
#define MAX_CLAUSES 100
extern Clause *database[MAX_CLAUSES];
extern int db_size;

void add_clause(Clause *clause);
void free_database();

#endif // PROLOG_DB_H
