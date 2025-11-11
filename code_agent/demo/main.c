#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdbool.h>

#define LINE_BUFFER_SIZE 1024 // Increased buffer size for longer inputs

#include "term.h"
#include "clause.h"
#include "knowledge_base.h"
#include "substitution.h"
#include "parser.h"
#include "unification.h"
#include "inference.h"

// Function prototypes for better modularity
void process_query_input(KnowledgeBase *kb, char *input_line);
void process_clause_input(KnowledgeBase *kb, char *input_line, bool interactive_mode);
void run_interpreter_loop(KnowledgeBase *kb, FILE *input_file, bool interactive_mode);

int main(int argc, char *argv[]) {
    KnowledgeBase *kb = create_knowledge_base();
    FILE *input_file = stdin;
    bool interactive_mode = true;

    if (argc > 1) {
        input_file = fopen(argv[1], "r");
        if (!input_file) {
            fprintf(stderr, "Error: Could not open file %s\n", argv[1]);
            return 1;
        }
        interactive_mode = false;
    }

    if (interactive_mode) {
        printf("Prolog Interpreter (Very Basic)\n");
        printf("Enter facts (e.g., p(a).), then queries (e.g., ?- p(a).)\n");
        printf("Type 'exit.' to quit.\n");
        fflush(stdout);
    }

    run_interpreter_loop(kb, input_file, interactive_mode);

    if (input_file != stdin) {
        fclose(input_file);
    }

    free_knowledge_base(kb);
    return 0;
}

void run_interpreter_loop(KnowledgeBase *kb, FILE *input_file, bool interactive_mode) {
    char line[LINE_BUFFER_SIZE]; // Buffer for input line

    while (true) {
        if (interactive_mode) {
            printf("?- ");
            fflush(stdout);
        }
        
        if (fgets(line, sizeof(line), input_file) == NULL) {
            break; // EOF or error
        }

        // Remove trailing newline
        line[strcspn(line, "\n")] = 0;

        if (strcmp(line, "exit.") == 0) {
            break;
        }

        char *input_ptr = line;
        input_ptr = skip_whitespace(input_ptr);

        if (*input_ptr == '\0') { // Skip empty lines or lines with only whitespace
            continue;
        }

        if (strncmp(input_ptr, "?-", 2) == 0) {
            process_query_input(kb, input_ptr + 2);

        } else {
            process_clause_input(kb, input_ptr, interactive_mode);
        }
    }
}

void process_query_input(KnowledgeBase *kb, char *input_ptr) {
    Term *query_term = parse_term(&input_ptr);
    input_ptr = skip_whitespace(input_ptr);
    if (query_term && *input_ptr == '.') {
        int solution_count = 0;
        resolve_query(kb, query_term, &solution_count);
        free_term(query_term);
    } else {
        fprintf(stderr, "Error: Invalid query syntax.\n");
        fflush(stderr);
        if (query_term) free_term(query_term);
    }
}

void process_clause_input(KnowledgeBase *kb, char *input_ptr, bool interactive_mode) {
    Clause *clause = parse_clause(&input_ptr);
    if (clause) {
        add_clause(kb, clause);
        if (interactive_mode) {
            printf("Clause added.\n");
            fflush(stdout);
        }
    } else {
        // input_ptr points to the beginning of the line buffer `line` in main
        // We should print the original line for better context if parsing failed
        fprintf(stderr, "Error: Invalid clause syntax. Problem near: \"%s\"\n", input_ptr);
        fflush(stderr);
    }
}
