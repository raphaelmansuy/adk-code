#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdbool.h>

#include "term.h"
#include "clause.h"
#include "knowledge_base.h"
#include "substitution.h"
#include "parser.h"
#include "unification.h"
#include "inference.h"

int main(int argc, char *argv[]) {
    KnowledgeBase *kb = create_knowledge_base();
    char line[256]; // Buffer for input line
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

        if (strncmp(input_ptr, "?-", 2) == 0) {
            // This is a query
            input_ptr += 2; // Skip "?-"
            Term *query_term = parse_term(&input_ptr);
             input_ptr = skip_whitespace(input_ptr);
            if (query_term && *input_ptr == '.') {
                if (!resolve_query(kb, query_term)) {
                    printf("No.\n");
                    fflush(stdout);
                }
                free_term(query_term);
            } else {
                fprintf(stderr, "Error: Invalid query syntax.\n");
                fflush(stderr);
                 if (query_term) free_term(query_term);
            }

        } else {
            // Assume it's a clause (fact or rule) to be added
            Clause *clause = parse_clause(&input_ptr);
            if (clause) {
                add_clause(kb, clause);
                if (interactive_mode) {
                    printf("Clause added.\n");
                    fflush(stdout); // Flush clause added message
                }
            } else {
                fprintf(stderr, "Error: Invalid clause syntax: %s\n", line);
                fflush(stderr);
            }
        }
    }

    if (input_file != stdin) {
        fclose(input_file);
    }

    free_knowledge_base(kb);
    return 0;
}
