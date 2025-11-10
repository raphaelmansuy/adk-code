#ifndef PROLOG_PARSER_H
#define PROLOG_PARSER_H

#include "prolog_data.h"

// Function to parse a single fact string into a Predicate structure
// Returns NULL on parsing error.
Predicate* parse_fact_string(const char *fact_string);

// Function to read facts from a file and add them to the database
// Returns true on success, false on error.
bool load_facts_from_file(const char *filepath);

#endif // PROLOG_PARSER_H
