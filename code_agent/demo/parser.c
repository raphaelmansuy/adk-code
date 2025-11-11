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
    *name_out = NULL; // Initialize to NULL
    char *start = input;
    while (*input && ((input == start && isalpha(*input)) || isalnum(*input) || *input == '_')) {
        input++;
    }
    int len = input - start;
    if (len == 0) {
        return NULL; // No name parsed, return NULL to indicate failure
    }
    *name_out = (char *)malloc(len + 1);
    if (!*name_out) {
        fprintf(stderr, "Parser Error: Memory allocation failed for name.\n");
        return NULL; // Indicate critical error
    }
    strncpy(*name_out, start, len);
    (*name_out)[len] = '\0';
    return input;
}

Term *parse_term(char **input) {
    *input = skip_whitespace(*input);
    if (!**input) return NULL;

    char *name;

    char *next_char_after_name = parse_name(*input, &name);
    if (!next_char_after_name) { // parse_name returns NULL on parsing failure (e.g., empty name)
        return NULL;
    }
    if (!name) { // parse_name sets name_out to NULL on memory allocation failure
        // This case should ideally be covered by next_char_after_name check now if parse_name returns NULL
        // for len == 0, but good to keep for malloc failure.
        return NULL; 
    }

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
            fprintf(stderr, "Parser Error: Memory allocation failed for arguments.\n");
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
                        fprintf(stderr, "Parser Error: Memory re-allocation failed for arguments.\n");
                        for (int i = 0; i < arity; ++i) free_term(args[i]);
                        free(args);
                        free(name);
                        return NULL;
                    }
                }

                Term *arg = parse_term(input); // Recursive call for nested terms
                if (!arg) {
                    fprintf(stderr, "Parser Error: Expected argument in compound term at %s\n", *input);
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
                    fprintf(stderr, "Parser Error: Expected ',' or ')' in compound term at %s\n", *input);
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

Clause *parse_clause(char **input) {
    *input = skip_whitespace(*input);
    if (!**input) return NULL;

    Term *head = parse_term(input);
    if (!head) return NULL;

    Term **body_goals = NULL;
    int num_body_goals = 0;
    int capacity = 2; // Initial capacity for body goals

    *input = skip_whitespace(*input);

    if (strncmp(*input, ":-", 2) == 0) {
        *input += 2; // Consume ":-"
        body_goals = (Term **)malloc(capacity * sizeof(Term *));
        if (!body_goals) {
            fprintf(stderr, "Parser Error: Memory allocation failed for clause body.\n");
            free_term(head);
            return NULL;
        }

        while (true) {
            *input = skip_whitespace(*input);
            Term *goal = parse_term(input);
            if (!goal) {
                fprintf(stderr, "Parser Error: Expected term in clause body at %s\n", *input);
                for (int i = 0; i < num_body_goals; ++i) free_term(body_goals[i]);
                free(body_goals);
                free_term(head);
                return NULL;
            }

            if (num_body_goals == capacity) {
                capacity *= 2;
                body_goals = (Term **)realloc(body_goals, capacity * sizeof(Term *));
                if (!body_goals) {
                    fprintf(stderr, "Parser Error: Memory re-allocation failed for clause body.\n");
                    for (int i = 0; i < num_body_goals; ++i) free_term(body_goals[i]);
                    free(body_goals);
                    free_term(head);
                    return NULL;
                }
            }
            body_goals[num_body_goals++] = goal;

            *input = skip_whitespace(*input);
            if (**input == '.') {
                break; // End of clause
            } else if (**input == ',') {
                (*input)++; // Consume ',', more goals to follow
            } else {
                fprintf(stderr, "Parser Error: Expected ',' or '.' after goal in clause body at %s\n", *input);
                for (int i = 0; i < num_body_goals; ++i) free_term(body_goals[i]);
                free(body_goals);
                free_term(head);
                return NULL;
            }
        }
    } else if (**input != '.') {
        fprintf(stderr, "Parser Error: Expected ':-' or '.' after head in clause at %s\n", *input);
        free_term(head);
        return NULL;
    }
    
    (*input)++; // Consume '.'

    // If no body was parsed, it's a fact, so body_goals remains NULL and num_body_goals is 0
    return create_clause(head, body_goals, num_body_goals);
}
