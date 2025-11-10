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

// Helper to parse a single predicate string (e.g., "parent(X,Y)")
static Predicate* parse_predicate_string(const char *pred_string, bool is_query) {
    char buffer[256];
    strncpy(buffer, pred_string, sizeof(buffer) - 1);
    buffer[sizeof(buffer) - 1] = '\0';

    char *pred_name_start = buffer;
    char *open_paren = strchr(buffer, '(');

    if (!open_paren) {
        fprintf(stderr, "Parse error: Missing '(' in predicate: %s\n", pred_string);
        return NULL;
    }

    *open_paren = '\0'; // Null-terminate predicate name
    char *pred_name = trim_whitespace(pred_name_start);

    if (strlen(pred_name) == 0) {
        fprintf(stderr, "Parse error: Empty predicate name in predicate: %s\n", pred_string);
        return NULL;
    }

    char *args_start = open_paren + 1;
    char *close_paren = strchr(args_start, ')');

    if (!close_paren) {
        fprintf(stderr, "Parse error: Missing ')' in predicate: %s\n", pred_string);
        return NULL;
    }

    // Only expect a dot at the end for facts/rules, not for predicates within a rule body or query
    if (!is_query) {
        char *dot = strchr(close_paren, '.');
        if (!dot || dot[1] != '\0') {
            fprintf(stderr, "Parse error: Missing '.' at end of fact/rule or extra characters: %s\n", pred_string);
            return NULL;
        }
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

Clause* parse_clause_string(const char *clause_string) {
    char buffer[512];
    strncpy(buffer, clause_string, sizeof(buffer) - 1);
    buffer[sizeof(buffer) - 1] = '\0';

    char *rule_separator = strstr(buffer, ":-");

    if (rule_separator) { // It's a rule
        *rule_separator = '\0'; // Null-terminate head
        char *head_str = trim_whitespace(buffer);
        char *body_str = trim_whitespace(rule_separator + 2);

        // Remove trailing dot from body string if present
        size_t body_len = strlen(body_str);
        if (body_len > 0 && body_str[body_len - 1] == '.') {
            body_str[body_len - 1] = '\0';
        }

        Predicate *head = parse_predicate_string(head_str, false);
        if (!head) return NULL;

        // Parse body predicates
        PredicateList *body = create_predicatelist(0);
        char *token = strtok(body_str, ",");
        while (token != NULL) {
            Predicate *body_pred = parse_predicate_string(trim_whitespace(token), true);
            if (!body_pred) {
                free_predicate(head);
                free_predicatelist(body);
                return NULL;
            }
            body->count++;
            body->predicates = (Predicate**)realloc(body->predicates, body->count * sizeof(Predicate*));
            if (!body->predicates) {
                perror("realloc failed during body parsing");
                free_predicate(head);
                free_predicatelist(body);
                exit(EXIT_FAILURE);
            }
            body->predicates[body->count - 1] = body_pred;
            token = strtok(NULL, ",");
        }

        Rule *rule = create_rule(head, body);
        return create_clause(RULE, rule);

    } else { // It's a fact
        // Check for trailing dot for facts
        size_t fact_len = strlen(buffer);
        if (fact_len > 0 && buffer[fact_len - 1] == '.') {
            buffer[fact_len - 1] = '\0';
        } else {
            fprintf(stderr, "Parse error: Missing '.' at end of fact: %s\n", clause_string);
            return NULL;
        }

        Predicate *fact_pred = parse_predicate_string(buffer, false);
        if (!fact_pred) return NULL;
        return create_clause(FACT, fact_pred);
    }
}

bool load_clauses_from_file(const char *filepath) {
    FILE *file = fopen(filepath, "r");
    if (!file) {
        perror("Failed to open clauses file");
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

        Clause *clause = parse_clause_string(trimmed_line);
        if (clause) {
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

PredicateList* parse_query_string(const char *query_string) {
    char buffer[512];
    strncpy(buffer, query_string, sizeof(buffer) - 1);
    buffer[sizeof(buffer) - 1] = '\0';

    char *query_start = buffer;

    // Optional: Handle leading "?- " for queries
    if (strncmp(query_start, "?-", 2) == 0) {
        query_start += 2;
        while (isspace((unsigned char)*query_start)) {
            query_start++;
        }
    }

    // Optional: Handle trailing '.' for queries
    size_t query_len = strlen(query_start);
    if (query_len > 0 && query_start[query_len - 1] == '.') {
        query_start[query_len - 1] = '\0'; // Remove the dot
    }

    char *trimmed_query = trim_whitespace(query_start);

    PredicateList *query_preds = create_predicatelist(0);
    char *token = strtok(trimmed_query, ",");
    while (token != NULL) {
        Predicate *pred = parse_predicate_string(trim_whitespace(token), true);
        if (!pred) {
            free_predicatelist(query_preds);
            return NULL;
        }
        query_preds->count++;
        query_preds->predicates = (Predicate**)realloc(query_preds->predicates, query_preds->count * sizeof(Predicate*));
        if (!query_preds->predicates) {
            perror("realloc failed during query parsing");
            free_predicatelist(query_preds);
            exit(EXIT_FAILURE);
        }
        query_preds->predicates[query_preds->count - 1] = pred;
        token = strtok(NULL, ",");
    }

    return query_preds;
}
