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

    // For facts and rule heads (when !is_query), the clause_string will contain the dot.
    // Individual predicates within a rule body or query do not have a dot.
    // The dot check for the overall clause is handled in parse_clause_string.

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

    // All clauses (facts and rules) must end with a dot
    size_t len = strlen(buffer);
    if (len == 0 || buffer[len - 1] != '.') {
        fprintf(stderr, "Parse error: Clause must end with '.': %s\n", clause_string);
        return NULL;
    }
    buffer[len - 1] = '\0'; // Remove the trailing dot for parsing

    char *rule_separator = strstr(buffer, ":-");

    if (rule_separator) { // It's a rule
        *rule_separator = '\0'; // Null-terminate head
        char *head_str = trim_whitespace(buffer);
        char *body_str = trim_whitespace(rule_separator + 2);

        Predicate *head = parse_predicate_string(head_str, false);
        if (!head) return NULL;

        // Parse body predicates
        PredicateList *body = create_predicatelist(0);
        char *current_pos = body_str;
        int paren_depth = 0;

        while (*current_pos != '\0') {
            char *pred_start = current_pos;
            char *comma_pos = NULL;
            bool found_comma = false;

            // Find the next comma that is not inside parentheses
            char *temp_pos = current_pos;
            while (*temp_pos != '\0') {
                if (*temp_pos == '(') {
                    paren_depth++;
                } else if (*temp_pos == ')') {
                    paren_depth--;
                } else if (*temp_pos == ',' && paren_depth == 0) {
                    comma_pos = temp_pos;
                    found_comma = true;
                    break;
                }
                temp_pos++;
            }

            char pred_buffer[256];
            if (found_comma) {
                strncpy(pred_buffer, pred_start, comma_pos - pred_start);
                pred_buffer[comma_pos - pred_start] = '\0';
                current_pos = comma_pos + 1;
            } else {
                strncpy(pred_buffer, pred_start, sizeof(pred_buffer) - 1);
                pred_buffer[sizeof(pred_buffer) - 1] = '\0';
                current_pos = pred_start + strlen(pred_start); // Move to end
            }

            Predicate *body_pred = parse_predicate_string(trim_whitespace(pred_buffer), true);
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

            // Trim leading whitespace for the next predicate
            while (*current_pos != '\0' && isspace((unsigned char)*current_pos)) {
                current_pos++;
            }
        }

        Rule *rule = create_rule(head, body);
        return create_clause(RULE, rule);

    } else { // It's a fact
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
