#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "prolog_data.h"
#include "prolog_db.h"
#include "prolog_query.h"
#include "prolog_parser.h"
#include <stdarg.h>

#define MAX_QUERY_LEN 256

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
            // Free already allocated terms
            for (int i = 0; i < arity - 1; i++) {
                free_term(p->args[i]); // Assuming free_term exists
            }
            free_predicate(p); // Free the predicate itself
            va_end(args);
            return NULL; // Indicate failure
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
    Predicate *p1 = create_predicate_from_terms("parent", create_term(ATOM, "john"), create_term(ATOM, "jim"), NULL);
    if (!p1) {
        fprintf(stderr, "Failed to create predicate for parent(john, jim).\n");
        return;
    }
    add_clause(create_clause(FACT, p1));

    // parent(john, jane).
    Predicate *p2 = create_predicate_from_terms("parent", create_term(ATOM, "john"), create_term(ATOM, "jane"), NULL);
    if (!p2) {
        fprintf(stderr, "Failed to create predicate for parent(john, jane).\n");
        return;
    }
    add_clause(create_clause(FACT, p2));

    // parent(mary, john).
    Predicate *p3 = create_predicate_from_terms("parent", create_term(ATOM, "mary"), create_term(ATOM, "john"), NULL);
    if (!p3) {
        fprintf(stderr, "Failed to create predicate for parent(mary, john).\n");
        return;
    }
    add_clause(create_clause(FACT, p3));

    // male(john).
    Predicate *p4 = create_predicate_from_terms("male", create_term(ATOM, "john"), NULL);
    if (!p4) {
        fprintf(stderr, "Failed to create predicate for male(john).\n");
        return;
    }
    add_clause(create_clause(FACT, p4));

    // female(mary).
    Predicate *p5 = create_predicate_from_terms("female", create_term(ATOM, "mary"), NULL);
    if (!p5) {
        fprintf(stderr, "Failed to create predicate for female(mary).\n");
        return;
    }
    add_clause(create_clause(FACT, p5));
}

int main(int argc, char *argv[]) {
    printf("--- Simple Prolog Interpreter (C) ---\n");

    if (argc > 1) {
        printf("Loading clauses from file: %s\n", argv[1]);
        if (!load_clauses_from_file(argv[1])) {
            fprintf(stderr, "Failed to load clauses from %s. Exiting.\n", argv[1]);
            free_database();
            return EXIT_FAILURE;
        }
        printf("Loaded %d clauses from %s.\n", db_size, argv[1]);
    } else {
        printf("Loading default facts.\n");
        // Add some facts
        add_default_facts();
    }

    printf("\n--- Interactive Query Mode ---\n");
    printf("Type 'exit.' to quit.\n");

    char query_buffer[MAX_QUERY_LEN];
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

        PredicateList *query_goals = parse_query_string(query_buffer);
        if (query_goals) {
            query(query_goals);
            free_predicatelist(query_goals);
        } else {
            fprintf(stderr, "Invalid query. Please try again.\n");
        }
    }

    // Clean up database (free memory)
    free_database();

    return 0;
}
