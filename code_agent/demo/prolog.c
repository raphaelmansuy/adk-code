#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdbool.h>

// --- Data Structures ---

// Represents a Prolog term (atom or variable)
enum TermType { ATOM, VARIABLE }; // Define enum globally

typedef struct Term {
    enum TermType type; // Use the global enum
    char *name;
} Term;

// Represents a Prolog predicate (e.g., parent(X, Y))
typedef struct Predicate {
    char *name;
    Term **args;
    int arity; // Number of arguments
} Predicate;

// Forward declarations for constructors
Term* create_term(enum TermType type, const char *name);
Predicate* create_predicate(const char *name, int arity);

// Represents a Prolog clause (fact or rule)
// For simplicity, we'll only handle facts initially
typedef struct Clause {
    Predicate *head;
    // Predicate **body; // For rules, not implemented yet
    // int body_len;
} Clause;

// Represents a variable binding in a substitution
typedef struct Binding {
    char *variable_name;
    Term *term;
} Binding;

// Represents a substitution (a list of variable bindings)
#define MAX_BINDINGS 50
typedef struct Substitution {
    Binding bindings[MAX_BINDINGS];
    int size;
} Substitution;

// A simple database to store clauses
#define MAX_CLAUSES 100
Clause *database[MAX_CLAUSES];
int db_size = 0;

// --- Memory Management and Constructors ---

// Create and free Substitution
Substitution* create_substitution() {
    Substitution *sub = (Substitution*)malloc(sizeof(Substitution));
    sub->size = 0;
    return sub;
}

void free_substitution(Substitution *sub) {
    for (int i = 0; i < sub->size; i++) {
        free(sub->bindings[i].variable_name);
        // Note: Term* in binding points to existing terms, not owned by binding
        // so we don't free them here to avoid double-freeing.
    }
    free(sub);
}

// Add a binding to the substitution
void add_binding(Substitution *sub, const char *var_name, Term *term) {
    if (sub->size < MAX_BINDINGS) {
        sub->bindings[sub->size].variable_name = strdup(var_name);
        sub->bindings[sub->size].term = term;
        sub->size++;
    } else {
        fprintf(stderr, "Substitution full! Cannot add binding for %s.\n", var_name);
    }
}

// Get the term a variable is bound to, or NULL if not bound
Term* get_binding(Substitution *sub, const char *var_name) {
    for (int i = 0; i < sub->size; i++) {
        if (strcmp(sub->bindings[i].variable_name, var_name) == 0) {
            return sub->bindings[i].term;
        }
    }
    return NULL;
}

// Deep copy a term (important for variable renaming in facts)
Term* copy_term(Term *original_term) {
    if (!original_term) return NULL;
    Term *new_term = create_term(original_term->type, original_term->name);
    return new_term;
}

// Deep copy a predicate (important for variable renaming in facts)
Predicate* copy_predicate(Predicate *original_pred) {
    if (!original_pred) return NULL;
    Predicate *new_pred = create_predicate(original_pred->name, original_pred->arity);
    for (int i = 0; i < original_pred->arity; i++) {
        new_pred->args[i] = copy_term(original_pred->args[i]);
    }
    return new_pred;
}

Term* create_term(enum TermType type, const char *name) {
    Term *term = (Term*)malloc(sizeof(Term));
    term->type = type;
    term->name = strdup(name);
    return term;
}

Predicate* create_predicate(const char *name, int arity) {
    Predicate *pred = (Predicate*)malloc(sizeof(Predicate));
    pred->name = strdup(name);
    pred->arity = arity;
    pred->args = (Term**)calloc(arity, sizeof(Term*)); // Initialize args to NULL
    return pred;
}

Clause* create_clause(Predicate *head) {
    Clause *clause = (Clause*)malloc(sizeof(Clause));
    clause->head = head;
    return clause;
}

void free_term(Term *term) {
    free(term->name);
    free(term);
}

void free_predicate(Predicate *pred) {
    free(pred->name);
    for (int i = 0; i < pred->arity; i++) {
        if (pred->args[i]) {
            free_term(pred->args[i]);
        }
    }
    free(pred->args);
    free(pred);
}

void free_clause(Clause *clause) {
    free_predicate(clause->head);
    free(clause);
}

// --- Database Operations ---

void add_clause(Clause *clause) {
    if (db_size < MAX_CLAUSES) {
        database[db_size++] = clause;
        printf("Fact added: %s(", clause->head->name);
        for (int i = 0; i < clause->head->arity; i++) {
            printf("%s%s", clause->head->args[i]->name, (i == clause->head->arity - 1) ? "" : ", ");
        }
        printf(").\n");
    } else {
        fprintf(stderr, "Database full!\n");
        free_clause(clause);
    }
}

// --- Unification (very basic for now) ---

// Forward declaration for mutual recursion
bool unify_terms(Term *t1, Term *t2, Substitution *sub);

// Checks if two terms unify and updates the substitution.
bool unify_terms(Term *t1, Term *t2, Substitution *sub) {
    // Resolve variables through substitution
    Term *val1 = (t1->type == VARIABLE) ? get_binding(sub, t1->name) : NULL;
    Term *val2 = (t2->type == VARIABLE) ? get_binding(sub, t2->name) : NULL;

    Term *resolved_t1 = val1 ? val1 : t1;
    Term *resolved_t2 = val2 ? val2 : t2;

    // If both are the same term (after resolving), they unify
    if (resolved_t1 == resolved_t2) {
        return true;
    }

    // Case 1: Both are atoms
    if (resolved_t1->type == ATOM && resolved_t2->type == ATOM) {
        return strcmp(resolved_t1->name, resolved_t2->name) == 0;
    }

    // Case 2: One is a variable, the other is a term
    if (resolved_t1->type == VARIABLE) {
        // Occurs check (simplified: avoid binding X to X itself, or X to a term containing X)
        // For this simple interpreter, we skip a full occurs check.
        add_binding(sub, resolved_t1->name, resolved_t2);
        return true;
    }
    if (resolved_t2->type == VARIABLE) {
        add_binding(sub, resolved_t2->name, resolved_t1);
        return true;
    }

    // Mismatch: e.g., an atom with a compound term (not implemented), or different atoms.
    return false;
}

// Checks if a query predicate unifies with a clause head and updates the substitution.
bool unify_predicates(Predicate *query, Predicate *fact_head, Substitution *sub) {
    if (strcmp(query->name, fact_head->name) != 0) {
        return false; // Different predicate names
    }
    if (query->arity != fact_head->arity) {
        return false; // Different arity
    }

    for (int i = 0; i < query->arity; i++) {
        if (!unify_terms(query->args[i], fact_head->args[i], sub)) {
            return false;
        }
    }
    return true;
}

// --- Query Engine (very basic) ---

// Helper to print a substitution
void print_substitution(Substitution *sub) {
    if (sub->size == 0) {
        printf("  Yes.\n");
        return;
    }
    printf("  Yes, with bindings:\n");
    for (int i = 0; i < sub->size; i++) {
        printf("    %s = ", sub->bindings[i].variable_name);
        if (sub->bindings[i].term->type == ATOM) {
            printf("%s\n", sub->bindings[i].term->name);
        } else { // VARIABLE
            // This case is more complex in a full interpreter (e.g., X = Y)
            printf("%s (variable, not fully resolved here)\n", sub->bindings[i].term->name);
        }
    }
}

// --- Query Engine ---

void query(Predicate *query_pred) {
    printf("Query: %s(", query_pred->name);
    for (int i = 0; i < query_pred->arity; i++) {
        printf("%s%s", query_pred->args[i]->name, (i == query_pred->arity - 1) ? "" : ", ");
    }
    printf(")?\n");

    bool any_found = false;
    for (int i = 0; i < db_size; i++) {
        Clause *fact_clause = database[i];

        // Create a fresh substitution for each attempt
        Substitution *sub = create_substitution();

        // Create a copy of the fact head to rename variables (if any) and avoid modifying database
        Predicate *fact_head_copy = copy_predicate(fact_clause->head);

        if (unify_predicates(query_pred, fact_head_copy, sub)) {
            print_substitution(sub);
            any_found = true;
        }

        // Clean up the copied fact and the substitution
        free_predicate(fact_head_copy);
        free_substitution(sub);
    }

    if (!any_found) {
        printf("  No.\n");
    }
}

// --- Main Function and Example Usage ---

int main() {
    printf("--- Simple Prolog Interpreter (C) ---\n");

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

    printf("\n--- Queries ---\n");

    // Query: parent(john, jim)? -> Yes.
    Predicate *q1 = create_predicate("parent", 2);
    q1->args[0] = create_term(ATOM, "john");
    q1->args[1] = create_term(ATOM, "jim");
    query(q1);
    free_predicate(q1);

    // Query: parent(mary, jim)? -> No.
    Predicate *q2 = create_predicate("parent", 2);
    q2->args[0] = create_term(ATOM, "mary");
    q2->args[1] = create_term(ATOM, "jim");
    query(q2);
    free_predicate(q2);

    // Query: male(john)? -> Yes.
    Predicate *q3 = create_predicate("male", 1);
    q3->args[0] = create_term(ATOM, "john");
    query(q3);
    free_predicate(q3);

    // Query with a variable (simplified handling): parent(john, X)?
    // Currently, this will just match parent(john, jim) and say "Yes.".
    // A proper interpreter would bind X to jim.
    Predicate *q4 = create_predicate("parent", 2);
    q4->args[0] = create_term(ATOM, "john");
    q4->args[1] = create_term(VARIABLE, "X");
    query(q4);
    free_predicate(q4);

    // Clean up database (free memory)
    for (int i = 0; i < db_size; i++) {
        free_clause(database[i]);
    }

    return 0;
}
