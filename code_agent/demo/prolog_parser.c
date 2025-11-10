#include "prolog_parser.h"
#include "prolog_data.h"
#include "prolog_db.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <ctype.h> // For isupper

// Helper to trim whitespace from a string
static char* trim_whitespace(char *str) {
    char *end;

    // Trim leading space
    while(isspace((unsigned char)*str)) str++;

    if(*str == 0)  // All spaces?
        return str;

    // Trim trailing space
    end = str + strlen(str) - 1;
    while(end > str && isspace((unsigned char)*end)) end--;

    // Write new null terminator character
    end[1] = '\0';

    return str;
}

Predicate* parse_fact_string(const char *fact_string) {
    char buffer[256];
    strncpy(buffer, fact_string, sizeof(buffer) - 1);
    buffer[sizeof(buffer) - 1] = '\0';

    char *pred_name_start = buffer;
    char *open_paren = strchr(buffer, '(');

    if (!open_paren) {
        fprintf(stderr, "Parse error: Missing '(' in fact: %s\n", fact_string);
        return NULL;
    }

    *open_paren = '\0'; // Null-terminate predicate name
    char *pred_name = trim_whitespace(pred_name_start);

    if (strlen(pred_name) == 0) {
        fprintf(stderr, "Parse error: Empty predicate name in fact: %s\n", fact_string);
        return NULL;
    }

    char *args_start = open_paren + 1;
    char *close_paren = strchr(args_start, ')');

    if (!close_paren) {
        fprintf(stderr, "Parse error: Missing ')' in fact: %s\n", fact_string);
        return NULL;
    }

    char *dot = strchr(close_paren, '.');
    if (!dot || dot[1] != '\0') {
        fprintf(stderr, "Parse error: Missing '.' at end of fact or extra characters: %s\n", fact_string);
        return NULL;
    }

    *close_paren = '\0'; // Null-terminate arguments string
    char *args_str = trim_whitespace(args_start);

    // Parse arguments
    Term **args = NULL;
    int arity = 0;

    if (strlen(args_str) > 0) {
        char *token = strtok(args_str, ",");
        while (token != NULL) {
            arity++;
            args = (Term**)realloc(args, arity * sizeof(Term*));
            if (!args) {
                perror("realloc failed during argument parsing");
                exit(EXIT_FAILURE);
            }
            char *trimmed_token = trim_whitespace(token);
            enum TermType type = isupper((unsigned char)trimmed_token[0]) ? VARIABLE : ATOM;
            args[arity - 1] = create_term(type, trimmed_token);
            token = strtok(NULL, ",");
        }
    }

    Predicate *pred = create_predicate(pred_name, arity);
    for (int i = 0; i < arity; i++) {
        pred->args[i] = args[i];
    }
    free(args); // Free the temporary array of Term pointers

    return pred;
}

bool load_facts_from_file(const char *filepath) {
    FILE *file = fopen(filepath, "r");
    if (!file) {
        perror("Failed to open facts file");
        return false;
    }

    char line[512];
    int line_num = 0;
    while (fgets(line, sizeof(line), file) != NULL) {
        line_num++;
        char *trimmed_line = trim_whitespace(line);

        // Skip empty lines or comment lines (starting with %)
        if (strlen(trimmed_line) == 0 || trimmed_line[0] == '%') {
            continue;
        }

        Predicate *pred = parse_fact_string(trimmed_line);
        if (pred) {
            Clause *clause = create_clause(pred);
            add_clause(clause);
        } else {
            fprintf(stderr, "Error parsing line %d: %s\n", line_num, trimmed_line);
            // Optionally, decide whether to continue or stop on error
            fclose(file);
            return false; // Stop on first error
        }
    }

    fclose(file);
    return true;
}

Predicate* parse_query_string(const char *query_string) {
    char buffer[256];
    strncpy(buffer, query_string, sizeof(buffer) - 1);
    buffer[sizeof(buffer) - 1] = '\0';

    char *pred_name_start = buffer;

    // Optional: Handle leading "?- " for queries
    if (strncmp(pred_name_start, "?-", 2) == 0) {
        pred_name_start += 2;
        while (isspace((unsigned char)*pred_name_start)) {
            pred_name_start++;
        }
    }

    char *open_paren = strchr(pred_name_start, '(');

    if (!open_paren) {
        fprintf(stderr, "Parse error: Missing '(' in query: %s\n", query_string);
        return NULL;
    }

    *open_paren = '\0'; // Null-terminate predicate name
    char *pred_name = trim_whitespace(pred_name_start);

    if (strlen(pred_name) == 0) {
        fprintf(stderr, "Parse error: Empty predicate name in query: %s\n", query_string);
        return NULL;
    }

    char *args_start = open_paren + 1;
    char *close_paren = strchr(args_start, ')');

    if (!close_paren) {
        fprintf(stderr, "Parse error: Missing ')' in query: %s\n", query_string);
        return NULL;
    }

    // Optional: Handle trailing '.' for queries
    char *dot = strchr(close_paren, '.');
    if (dot && dot[1] == '\0') {
        *dot = '\0'; // Remove the dot
    }

    *close_paren = '\0'; // Null-terminate arguments string
    char *args_str = trim_whitespace(args_start);

    // Parse arguments
    Term **args = NULL;
    int arity = 0;

    if (strlen(args_str) > 0) {
        char *token = strtok(args_str, ",");
        while (token != NULL) {
            arity++;
            args = (Term**)realloc(args, arity * sizeof(Term*));
            if (!args) {
                perror("realloc failed during argument parsing for query");
                exit(EXIT_FAILURE);
            }
            char *trimmed_token = trim_whitespace(token);
            enum TermType type = isupper((unsigned char)trimmed_token[0]) ? VARIABLE : ATOM;
            args[arity - 1] = create_term(type, trimmed_token);
            token = strtok(NULL, ",");
        }
    }

    Predicate *pred = create_predicate(pred_name, arity);
    for (int i = 0; i < arity; i++) {
        pred->args[i] = args[i];
    }
    free(args); // Free the temporary array of Term pointers

    return pred;
}
