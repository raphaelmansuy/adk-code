#ifndef PROLOG_PARSER_H
#define PROLOG_PARSER_H

#include "prolog_data.h"

// Function to parse a clause string (fact or rule) into a Clause structure
// Returns NULL on parsing error.
Clause* parse_clause_string(const char *clause_string);

// Function to parse a query string into a PredicateList structure.
// Returns NULL on parsing error.
PredicateList* parse_query_string(const char *query_string);

// Function to read clauses from a file and add them to the database
// Returns true on success, false on error.
bool load_clauses_from_file(const char *filepath);

#endif // PROLOG_PARSER_H
