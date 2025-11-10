#ifndef PARSER_H
#define PARSER_H

#include "term.h"
#include "clause.h"
#include <stdbool.h>

// Function prototypes for parsing
bool is_uppercase(char c);
char *skip_whitespace(char *str);
char *parse_name(char *input, char **name_out);
Term *parse_term(char **input);
Clause *parse_clause(char **input);

#endif // PARSER_H
