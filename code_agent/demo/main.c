#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "prolog_data.h"
#include "prolog_db.h"
#include "prolog_query.h"
#include "prolog_parser.h"
#include <stdarg.h>

// Helper function to create a predicate from a list of terms
Predicate *create_predicate_from_terms(const char *name, ...) {
    Predicate *p = create_predicate(name, 0); // Start with 0 arity
    va_list args;
    va_start(args, name);
    Term *term;
    int arity = 0;
    while ((term = va_arg(args, Term *)) != NULL) {
        arity++;
        Term **new_args = realloc(p->args, arity * sizeof(Term *));
        if (new_args == NULL) {
            perror("Failed to reallocate memory for predicate arguments");
            // Free already allocated terms before exiting
            for (int i = 0; i < arity - 1; i++) {
                free_term(p->args[i]); // Assuming free_term exists
            }
            free_predicate(p); // Free the predicate itself
            va_end(args);
            exit(EXIT_FAILURE);
        }
        p->args = new_args;
        p->args[arity - 1] = term;
    }
    va_end(args);
    p->arity = arity;
    return p;
}

// Function to add default facts to the database
void add_default_facts() {
    // parent(john, jim).
    add_clause(create_clause(create_predicate_from_terms("parent", create_term(ATOM, "john"), create_term(ATOM, "jim"), NULL)));

    // parent(john, jane).
    add_clause(create_clause(create_predicate_from_terms("parent", create_term(ATOM, "john"), create_term(ATOM, "jane"), NULL)));

    // parent(mary, john).
    add_clause(create_clause(create_predicate_from_terms("parent", create_term(ATOM, "mary"), create_term(ATOM, "john"), NULL)));

    // male(john).
    add_clause(create_clause(create_predicate_from_terms("male", create_term(ATOM, "john"), NULL)));

    // female(mary).
    add_clause(create_clause(create_predicate_from_terms("female", create_term(ATOM, "mary"), NULL)));
}

// Helper function to run a query and free the predicate
void run_query_and_free(Predicate *query_predicate) {
    query(query_predicate);
    free_predicate(query_predicate);
}

int main(int argc, char *argv[]) {
    printf("--- Simple Prolog Interpreter (C) ---\n");

    if (argc > 1) {
        printf("Loading facts from file: %s\n", argv[1]);
        if (!load_facts_from_file(argv[1])) {
            fprintf(stderr, "Failed to load facts from %s. Exiting.\n", argv[1]);
            free_database();
            return EXIT_FAILURE;
        }
        printf("Loaded %d facts from %s.\n", db_size, argv[1]);
    } else {
        printf("Loading default facts.\n");
        // Add some facts
        add_default_facts();
    }

    printf("\n--- Interactive Query Mode ---\n");
    printf("Type 'exit.' to quit.\n");

    char query_buffer[256];
    while (1) {
        printf("?- ");
        if (fgets(query_buffer, sizeof(query_buffer), stdin) == NULL) {
            break; // EOF or error
        }

        // Remove trailing newline character if present
        query_buffer[strcspn(query_buffer, "\n")] = 0;

        if (strcmp(query_buffer, "exit.") == 0 || strcmp(query_buffer, "exit") == 0) {
            break;
        }

        Predicate *query_predicate = parse_query_string(query_buffer);
        if (query_predicate) {
            query(query_predicate);
            free_predicate(query_predicate);
        } else {
            fprintf(stderr, "Invalid query. Please try again.\n");
        }
    }

    // Clean up database (free memory)
    free_database();

    return 0;
}
