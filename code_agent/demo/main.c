#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "prolog_data.h"
#include "prolog_db.h"
#include "prolog_query.h"
#include "prolog_parser.h"

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
        // parent(john, jim).
        Predicate *p1 = create_predicate("parent", 2);
        p1->args[0] = create_term(ATOM, "john");
        p1->args[1] = create_term(ATOM, "jim");
        add_clause(create_clause(p1));

        // parent(john, jane).
        Predicate *p_jane = create_predicate("parent", 2);
        p_jane->args[0] = create_term(ATOM, "john");
        p_jane->args[1] = create_term(ATOM, "jane");
        add_clause(create_clause(p_jane));

        // parent(mary, john).
        Predicate *p2 = create_predicate("parent", 2);
        p2->args[0] = create_term(ATOM, "mary");
        p2->args[1] = create_term(ATOM, "john");
        add_clause(create_clause(p2));

        // male(john).
        Predicate *m1 = create_predicate("male", 1);
        m1->args[0] = create_term(ATOM, "john");
        add_clause(create_clause(m1));

        // female(mary).
        Predicate *f1 = create_predicate("female", 1);
        f1->args[0] = create_term(ATOM, "mary");
        add_clause(create_clause(f1));
    }

    printf("\n--- Queries ---\n");

    // Query: parent(pam, bob)?
    Predicate *q1 = create_predicate("parent", 2);
    q1->args[0] = create_term(ATOM, "pam");
    q1->args[1] = create_term(ATOM, "bob");
    query(q1);
    free_predicate(q1);

    // Query: parent(bob, X)?
    Predicate *q2 = create_predicate("parent", 2);
    q2->args[0] = create_term(ATOM, "bob");
    q2->args[1] = create_term(VARIABLE, "X");
    query(q2);
    free_predicate(q2);

    // Query: male(tom)?
    Predicate *q3 = create_predicate("male", 1);
    q3->args[0] = create_term(ATOM, "tom");
    query(q3);
    free_predicate(q3);

    // Query: female(X)?
    Predicate *q4 = create_predicate("female", 1);
    q4->args[0] = create_term(VARIABLE, "X");
    query(q4);
    free_predicate(q4);

    // Query: parent(Y, ann)?
    Predicate *q5 = create_predicate("parent", 2);
    q5->args[0] = create_term(VARIABLE, "Y");
    q5->args[1] = create_term(ATOM, "ann");
    query(q5);
    free_predicate(q5);

    // Clean up database (free memory)
    free_database();

    return 0;
}
