#include "parser.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <ctype.h> // Required for isalpha, isalnum

bool is_uppercase(char c) {
    return c >= 'A' && c <= 'Z';
}

char *skip_whitespace(char *str) {
    while (*str && (*str == ' ' || *str == '\t' || *str == '\n' || *str == '\r')) {
        str++;
    }
    return str;
}

char *parse_name(char *input, char **name_out) {
    char *start = input;
    while (*input && ((input == start && isalpha(*input)) || isalnum(*input) || *input == '_')) {
        input++;
    }
    int len = input - start;
    *name_out = (char *)malloc(len + 1);
    strncpy(*name_out, start, len);
    (*name_out)[len] = '\0';
    return input;
}

Term *parse_term(char **input) {
    *input = skip_whitespace(*input);
    if (!**input) return NULL;

    char *name;

    char *next_char_after_name = parse_name(*input, &name);
    if (!name) return NULL;

    *input = next_char_after_name;
    *input = skip_whitespace(*input);

    if (**input == '(') {
        // It's a compound term
        (*input)++; // Consume '('

        Term **args = NULL;
        int arity = 0;
        int capacity = 2; 

        args = (Term **)malloc(capacity * sizeof(Term *));
        if (!args) {
            fprintf(stderr, "Memory allocation failed for arguments.\n");
            free(name);
            return NULL;
        }

        *input = skip_whitespace(*input);

        if (**input != ')') { // Check if there are arguments
            while (true) {
                if (arity == capacity) {
                    capacity *= 2;
                    args = (Term **)realloc(args, capacity * sizeof(Term *));
                    if (!args) {
                        fprintf(stderr, "Memory re-allocation failed for arguments.\n");
                        for (int i = 0; i < arity; ++i) free_term(args[i]);
                        free(args);
                        free(name);
                        return NULL;
                    }
                }

                Term *arg = parse_term(input); // Recursive call for nested terms
                if (!arg) {
                    fprintf(stderr, "Error: Expected argument in compound term at %s\n", *input);
                    for (int i = 0; i < arity; ++i) free_term(args[i]);
                    free(args);
                    free(name);
                    return NULL;
                }
                args[arity++] = arg;

                *input = skip_whitespace(*input);
                if (**input == ')') {
                    break; // End of arguments
                } else if (**input == ',') {
                    (*input)++; // Consume ','
                    *input = skip_whitespace(*input);
                } else {
                    fprintf(stderr, "Error: Expected ',' or ')' in compound term at %s\n", *input);
                    for (int i = 0; i < arity; ++i) free_term(args[i]);
                    free(args);
                    free(name);
                    return NULL;
                }
            }
        }
        (*input)++; // Consume ')'

        Term *compound = create_compound_term(name, arity);
        for (int i = 0; i < arity; ++i) {
            compound->args[i] = args[i];
        }
        free(args); 
        free(name); 
        return compound;

    } else {
        // It's an atom or a variable
        Term *term;
        if (is_uppercase(name[0])) {
            term = create_term(VARIABLE, name);
        } else {
            term = create_term(ATOM, name);
        }
        free(name);
        return term;
    }
}

Clause *parse_fact(char **input) {
    *input = skip_whitespace(*input);
    if (!**input) return NULL;

    Term *head = parse_term(input);
    if (!head) return NULL;

    *input = skip_whitespace(*input);
    if (**input != '.') {
        fprintf(stderr, "Error: Expected '.' at end of fact at %s\n", *input);
        free_term(head);
        return NULL;
    }
    (*input)++; // Consume '.'

    return create_clause(head, NULL, 0); // No body for facts
}
