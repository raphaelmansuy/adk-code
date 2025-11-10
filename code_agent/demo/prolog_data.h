#ifndef PROLOG_DATA_H
#define PROLOG_DATA_H

#include <stdbool.h>

// --- Data Structures ---

// Represents a Prolog term (atom or variable)
enum TermType { ATOM, VARIABLE };

typedef struct Term {
    enum TermType type;
    char *name;
} Term;

// Represents a Prolog predicate (e.g., parent(X, Y))
typedef struct Predicate {
    char *name;
    Term **args;
    int arity; // Number of arguments
} Predicate;

// Represents a list of predicates (for rule bodies)
typedef struct PredicateList {
    Predicate **predicates;
    int count;
} PredicateList;

// Represents a Prolog rule
typedef struct Rule {
    Predicate *head;
    PredicateList *body;
} Rule;

// Represents a Prolog clause (fact or rule)
enum ClauseType { FACT, RULE };

typedef struct Clause {
    enum ClauseType type;
    union {
        Predicate *fact; // For facts
        Rule *rule;      // For rules
    } content;
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

// --- Memory Management and Constructors ---

Term* create_term(enum TermType type, const char *name);
void free_term(Term *term);
Term* copy_term(Term *original_term);

Predicate* create_predicate(const char *name, int arity);
void free_predicate(Predicate *pred);
Predicate* copy_predicate(Predicate *original_pred);

PredicateList* create_predicatelist(int count);
void free_predicatelist(PredicateList *list);
PredicateList* copy_predicatelist(PredicateList *original_list);

Rule* create_rule(Predicate *head, PredicateList *body);
void free_rule(Rule *rule);
Rule* copy_rule(Rule *original_rule);

Clause* create_clause(enum ClauseType type, void *content_ptr);
void free_clause(Clause *clause);

Substitution* create_substitution();
void free_substitution(Substitution *sub);
void add_binding(Substitution *sub, const char *var_name, Term *term);
Term* get_binding(Substitution *sub, const char *var_name);
Substitution* copy_substitution(Substitution *original_sub);

#endif // PROLOG_DATA_H
