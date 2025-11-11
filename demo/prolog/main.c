#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdbool.h>

// --- Prolog Data Structures ---

typedef enum {
    TERM_ATOM,
    TERM_VARIABLE,
    TERM_NUMBER,
    TERM_STRUCTURE,
    TERM_LIST // Future: for [a, b, c]
} PrologTermType;

// Forward declaration for recursive types
typedef struct PrologTerm PrologTerm;
typedef struct PrologStructure PrologStructure;

struct PrologTerm {
    PrologTermType type;
    union {
        char *atom_name;
        char *variable_name;
        int number_value;
        PrologStructure *structure;
        // For lists, we'd have a pointer to a list structure
    } value;
};

struct PrologStructure {
    char *functor; // The name of the predicate/functor
    PrologTerm **args; // Array of pointers to arguments
    int arity; // Number of arguments
};

// Global database of facts and rules (future)
typedef struct {
    PrologStructure *head; // For facts, this is the entire fact. For rules, the head.
    PrologStructure **body; // For rules, array of body goals. NULL for facts.
    int body_count;
} Clause;

// Forward declaration for free_prolog_term (needed by free_clause)
void free_prolog_term(PrologTerm* term);

// --- Prolog Database ---

typedef struct PrologDatabase {
    Clause **clauses;
    int count;
    int capacity;
} PrologDatabase;

PrologDatabase* global_database;

void init_database() {
    global_database = (PrologDatabase*)malloc(sizeof(PrologDatabase));
    if (!global_database) { fprintf(stderr, "Memory allocation failed\n"); exit(1); }
    global_database->capacity = 4;
    global_database->clauses = (Clause**)malloc(sizeof(Clause*) * global_database->capacity);
    if (!global_database->clauses) { fprintf(stderr, "Memory allocation failed\n"); exit(1); }
    global_database->count = 0;
}

void add_clause(Clause* clause) {
    if (global_database->count >= global_database->capacity) {
        global_database->capacity *= 2;
        global_database->clauses = (Clause**)realloc(global_database->clauses, sizeof(Clause*) * global_database->capacity);
        if (!global_database->clauses) { fprintf(stderr, "Memory allocation failed\n"); exit(1); }
    }
    global_database->clauses[global_database->count++] = clause;
}

// Frees memory for a single clause (recursively frees terms)
void free_clause(Clause* clause) {
    if (!clause) return;
    if (clause->head) {
        // The head is a PrologStructure, but free_prolog_term handles freeing the entire term structure
        // We need to create a temporary PrologTerm to pass to free_prolog_term
        PrologTerm temp_term; 
        temp_term.type = TERM_STRUCTURE;
        temp_term.value.structure = clause->head;
        free_prolog_term(&temp_term); 
    }
    if (clause->body) {
        for (int i = 0; i < clause->body_count; i++) {
            // Each element in body is a PrologStructure*, so create a temp PrologTerm
            PrologTerm temp_term; 
            temp_term.type = TERM_STRUCTURE;
            temp_term.value.structure = clause->body[i];
            free_prolog_term(&temp_term); 
        }
        free(clause->body);
    }
    free(clause);
}

void free_database() {
    if (!global_database) return;
    for (int i = 0; i < global_database->count; i++) {
        free_clause(global_database->clauses[i]);
    }
    free(global_database->clauses);
    free(global_database);
    global_database = NULL;
}

// --- Substitution / Variable Bindings ---

typedef struct Substitution Substitution;

struct Substitution {
    char *variable_name;
    PrologTerm *term;
    Substitution *next;
};

// Forward declaration for copy_term
PrologTerm* copy_term(PrologTerm* original);

Substitution* create_substitution(char *var_name, PrologTerm *term) {
    Substitution* sub = (Substitution*)malloc(sizeof(Substitution));
    if (!sub) { fprintf(stderr, "Memory allocation failed\n"); exit(1); }
    sub->variable_name = strdup(var_name);
    sub->term = copy_term(term); // Now, the substitution OWNS a copy of the term
    sub->next = NULL;
    return sub;
}

// Add a binding to the front of a substitution list
Substitution* add_binding(Substitution* current_sub, char *var_name, PrologTerm *term) {
    Substitution* new_sub = create_substitution(var_name, term);
    new_sub->next = current_sub;
    return new_sub;
}

// Free a substitution list (now frees the terms it owns)
void free_substitution(Substitution* sub) {
    Substitution* current = sub;
    while (current) {
        Substitution* next = current->next;
        free(current->variable_name);
        free_prolog_term(current->term); // Free the copied term owned by this substitution
        free(current);
        current = next;
    }
}

// Find a variable's binding in a substitution list
PrologTerm* lookup_binding(Substitution* sub, char *var_name) {
    Substitution* current = sub;
    while (current) {
        if (strcmp(current->variable_name, var_name) == 0) {
            return current->term;
        }
        current = current->next;
    }
    return NULL;
}

// --- Lexer ---

typedef enum {
    TOKEN_EOF = 0,
    TOKEN_ATOM,
    TOKEN_VARIABLE,
    TOKEN_NUMBER,
    TOKEN_LPAREN,     // (
    TOKEN_RPAREN,     // )
    TOKEN_COMMA,      // ,
    TOKEN_DOT,        // .
    TOKEN_COLON_DASH, // :-
    TOKEN_UNKNOWN
} TokenType;

typedef struct {
    TokenType type;
    char *lexeme;
    // For numbers, we might store the value directly
    // For future: line and column number for error reporting
} Token;

// Global variable to simulate input stream for now
char *input_buffer;
int current_char_index = 0;

// Placeholder for getting the next character from input
char consume_char() {
    if (input_buffer[current_char_index] == '\0') {
        return EOF;
    }
    return input_buffer[current_char_index++];
}

char peek_char() {
    if (input_buffer[current_char_index] == '\0') {
        return EOF;
    }
    return input_buffer[current_char_index];
}

// Global variable for the current token (the one consumed by parser)
Token current_token; // The token the parser is currently looking at
Token next_token_buffer; // Holds the next token if it has been peeked
bool has_next_token = false; // Flag to indicate if next_token_buffer contains a valid token

// Forward declarations
void get_next_token();
Token read_next_token_from_input(); // Internal function to actually read and create a token

// Gets the next token and makes it current_token
void get_next_token() {
    if (current_token.lexeme) {
        free(current_token.lexeme);
        current_token.lexeme = NULL;
    }

    if (has_next_token) {
        current_token = next_token_buffer;
        has_next_token = false;
    } else {
        current_token = read_next_token_from_input();
    }
}

// Peeks the next token without consuming it
Token peek_token() {
    if (!has_next_token) {
        next_token_buffer = read_next_token_from_input();
        has_next_token = true;
    }
    return next_token_buffer;
}

// The actual lexer logic to read from input_buffer
Token read_next_token_from_input() {
    Token token;
    token.lexeme = NULL;

    // Skip whitespace
    char c;
    while ((c = peek_char()) != EOF && (c == ' ' || c == '\t' || c == '\n' || c == '\r')) {
        consume_char(); // Consume the whitespace
    }
    c = consume_char(); // Consume the first non-whitespace character

    if (c == EOF) {
        token.type = TOKEN_EOF;
        return token;
    }

    // Handle single-character tokens
    switch (c) {
        case '(': token.type = TOKEN_LPAREN; token.lexeme = strdup("("); return token;
        case ')': token.type = TOKEN_RPAREN; token.lexeme = strdup(")"); return token;
        case ',': token.type = TOKEN_COMMA; token.lexeme = strdup(","); return token;
        case '.': token.type = TOKEN_DOT; token.lexeme = strdup("."); return token;
        case ':': {
            char next_c = peek_char();
            if (next_c == '-') {
                consume_char(); // Consume the '-'
                token.type = TOKEN_COLON_DASH;
                token.lexeme = strdup(":-");
                return token;
            } else {
                // If it's just ':', treat as unknown for now, or error
                token.type = TOKEN_UNKNOWN;
                token.lexeme = strdup(":");
                return token;
            }
        }
    }

    // Handle atoms and variables
    if ((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_') {
        char buffer[256]; // Temporary buffer for lexeme
        int i = 0;
        buffer[i++] = c;

        // Read alphanumeric characters and underscores
        char next_c;
        while ((next_c = peek_char()) != EOF &&
               ((next_c >= 'a' && next_c <= 'z') ||
                (next_c >= 'A' && next_c <= 'Z') ||
                (next_c >= '0' && next_c <= '9') ||
                next_c == '_')) {
            buffer[i++] = consume_char(); // Consume only if it's part of the token
        }
        buffer[i] = '\0';

        if (c >= 'A' && c <= 'Z') {
            token.type = TOKEN_VARIABLE;
        } else {
            token.type = TOKEN_ATOM;
        }
        token.lexeme = strdup(buffer);
        return token;
    }

    // Handle numbers
    if (c >= '0' && c <= '9') {
        char buffer[256];
        int i = 0;
        buffer[i++] = c;

        char next_c;
        while ((next_c = peek_char()) != EOF &&
               (next_c >= '0' && next_c <= '9')) {
            buffer[i++] = consume_char(); // Consume only if it's part of the token
        }
        buffer[i] = '\0';

        token.type = TOKEN_NUMBER;
        token.lexeme = strdup(buffer);
        return token;
    }

    // If none of the above, it's an unknown token
    token.type = TOKEN_UNKNOWN;
    char unknown_lexeme[2];
    unknown_lexeme[0] = c;
    unknown_lexeme[1] = '\0';
    token.lexeme = strdup(unknown_lexeme);
    return token;
}

// Utility for printing token types
const char* token_type_to_string(TokenType type) {
    switch (type) {
        case TOKEN_EOF: return "EOF";
        case TOKEN_ATOM: return "ATOM";
        case TOKEN_VARIABLE: return "VARIABLE";
        case TOKEN_NUMBER: return "NUMBER";
        case TOKEN_LPAREN: return "LPAREN";
        case TOKEN_RPAREN: return "RPAREN";
        case TOKEN_COMMA: return "COMMA";
        case TOKEN_DOT: return "DOT";
        case TOKEN_COLON_DASH: return "COLON_DASH";
        case TOKEN_UNKNOWN: return "UNKNOWN";
        default: return "<UNKNOWN_TYPE>";
    }
}

// --- Parser ---

// Helper to allocate a new PrologTerm
PrologTerm* create_term(PrologTermType type) {
    PrologTerm* term = (PrologTerm*)malloc(sizeof(PrologTerm));
    if (!term) { fprintf(stderr, "Memory allocation failed\n"); exit(1); }
    term->type = type;
    return term;
}

// Error reporting for parser
void parser_error(const char* message) {
    fprintf(stderr, "Parser Error: %s at token \"%s\" (Type: %s)\n", 
            message, current_token.lexeme ? current_token.lexeme : "EOF",
            token_type_to_string(current_token.type));
    exit(1);
}

// Consume the current token if its type matches expected_type
bool match_token(TokenType expected_type) {
    if (current_token.type == expected_type) {
        if (current_token.lexeme) {
            free(current_token.lexeme);
            current_token.lexeme = NULL;
        }
        get_next_token(); // This will now correctly use peeked_token if available
        return true;
    }
    return false;
}

// Parses an atom and returns a PrologTerm
PrologTerm* parse_atom() {
    if (current_token.type != TOKEN_ATOM) {
        parser_error("Expected an atom");
        return NULL; // Should exit
    }
    PrologTerm* term = create_term(TERM_ATOM);
    term->value.atom_name = strdup(current_token.lexeme);
    match_token(TOKEN_ATOM); // Consume the atom token
    return term;
}

// Parses a variable and returns a PrologTerm
PrologTerm* parse_variable() {
    if (current_token.type != TOKEN_VARIABLE) {
        parser_error("Expected a variable");
        return NULL; // Should exit
    }
    PrologTerm* term = create_term(TERM_VARIABLE);
    term->value.variable_name = strdup(current_token.lexeme);
    match_token(TOKEN_VARIABLE); // Consume the variable token
    return term;
}

// Parses a number and returns a PrologTerm
PrologTerm* parse_number() {
    if (current_token.type != TOKEN_NUMBER) {
        parser_error("Expected a number");
        return NULL; // Should exit
    }
    PrologTerm* term = create_term(TERM_NUMBER);
    term->value.number_value = atoi(current_token.lexeme);
    match_token(TOKEN_NUMBER); // Consume the number token
    return term;
}

// Forward declaration for parse_term, as parse_structure will call it.
PrologTerm* parse_term();

// Parses a structure (e.g., functor(arg1, arg2) or just a functor if arity 0)
PrologTerm* parse_structure() {
    // A structure starts with an atom (the functor)
    if (current_token.type != TOKEN_ATOM) {
        parser_error("Expected functor (atom) for structure");
        return NULL;
    }

    char *functor_name = strdup(current_token.lexeme);
    match_token(TOKEN_ATOM); // Consume the functor atom

    PrologStructure* structure = (PrologStructure*)malloc(sizeof(PrologStructure));
    if (!structure) { fprintf(stderr, "Memory allocation failed\n"); exit(1); }
    structure->functor = functor_name;
    structure->args = NULL;
    structure->arity = 0;

    if (current_token.type == TOKEN_LPAREN) {
        match_token(TOKEN_LPAREN); // Consume '('
        
        // Parse arguments
        int capacity = 2; // Initial capacity for arguments
        structure->args = (PrologTerm**)malloc(sizeof(PrologTerm*) * capacity);
        if (!structure->args) { fprintf(stderr, "Memory allocation failed\n"); exit(1); }

        do {
            if (structure->arity >= capacity) {
                capacity *= 2;
                structure->args = (PrologTerm**)realloc(structure->args, sizeof(PrologTerm*) * capacity);
                if (!structure->args) { fprintf(stderr, "Memory allocation failed\n"); exit(1); }
            }
            structure->args[structure->arity++] = parse_term();

        } while (current_token.type == TOKEN_COMMA && match_token(TOKEN_COMMA));

        if (!match_token(TOKEN_RPAREN)) {
            parser_error("Expected ')' after structure arguments");
        }
    }

    PrologTerm* term = create_term(TERM_STRUCTURE);
    term->value.structure = structure;
    return term;
}

// Parses any valid Prolog term
PrologTerm* parse_term() {
    if (current_token.type == TOKEN_ATOM) {
        // Look ahead to see if it's a structure or a bare atom
        Token next_token = peek_token();
        if (next_token.type == TOKEN_LPAREN) {
            return parse_structure();
        } else {
            return parse_atom();
        }
    } else if (current_token.type == TOKEN_VARIABLE) {
        return parse_variable();
    } else if (current_token.type == TOKEN_NUMBER) {
        return parse_number();
    } else {
        parser_error("Unexpected token type for term");
        return NULL; // Should exit
    }
}

// Utility for printing PrologTerm (for debugging)
void print_prolog_term(PrologTerm* term) {
    if (!term) { printf("NULL"); return; }

    switch (term->type) {
        case TERM_ATOM:
            printf("Atom(\"%s\")", term->value.atom_name);
            break;
        case TERM_VARIABLE:
            printf("Var(\"%s\")", term->value.variable_name);
            break;
        case TERM_NUMBER:
            printf("Num(%d)", term->value.number_value);
            break;
        case TERM_STRUCTURE:
            printf("Struct(\"%s\", arity=%d, args=[ ", term->value.structure->functor, term->value.structure->arity);
            for (int i = 0; i < term->value.structure->arity; i++) {
                print_prolog_term(term->value.structure->args[i]);
                if (i < term->value.structure->arity - 1) {
                    printf(", ");
                }
            }
            printf(" ])");
            break;
        case TERM_LIST:
            printf("List(TODO)"); // Not implemented yet
            break;
    }
}

// Frees memory associated with a PrologTerm
void free_prolog_term(PrologTerm* term) {
    if (!term) return;

    switch (term->type) {
        case TERM_ATOM:
            free(term->value.atom_name);
            break;
        case TERM_VARIABLE:
            free(term->value.variable_name);
            break;
        case TERM_NUMBER:
            // No dynamic memory for number_value
            break;
        case TERM_STRUCTURE:
            free(term->value.structure->functor);
            for (int i = 0; i < term->value.structure->arity; i++) {
                free_prolog_term(term->value.structure->args[i]);
            }
            free(term->value.structure->args);
            free(term->value.structure);
            break;
        case TERM_LIST:
            // TODO: Free list elements
            break;
    }
    free(term);
}

// --- Unification ---
// Forward declaration
bool unify(PrologTerm* term1, PrologTerm* term2, Substitution** sub);

// Dereferences a variable through a substitution list (returns the bound term or the original variable term)
PrologTerm* dereference_term(PrologTerm* term, Substitution* sub) {
    if (term->type == TERM_VARIABLE) {
        PrologTerm* binding = lookup_binding(sub, term->value.variable_name);
        if (binding) {
            return dereference_term(binding, sub); // Recursively dereference
        }
    }
    return term;
}

// Occur check: Does variable var occur in term? (prevents infinite terms)
bool occur_check(PrologTerm* var_term, PrologTerm* term, Substitution* sub) {
    if (var_term->type != TERM_VARIABLE) { /* This should not happen */ return false; }

    PrologTerm* dereferenced_term = dereference_term(term, sub); // Pass current sub

    if (dereferenced_term->type == TERM_VARIABLE) {
        return strcmp(var_term->value.variable_name, dereferenced_term->value.variable_name) == 0;
    } else if (dereferenced_term->type == TERM_STRUCTURE) {
        for (int i = 0; i < dereferenced_term->value.structure->arity; i++) {
            if (occur_check(var_term, dereferenced_term->value.structure->args[i], sub)) {
                return true;
            }
        }
    }
    return false;
}

// Main unification function
bool unify(PrologTerm* term1, PrologTerm* term2, Substitution** sub) {
    printf("  Unifying: "); print_prolog_term(term1); printf(" with "); print_prolog_term(term2); printf("\n");

    // Dereference terms first
    term1 = dereference_term(term1, *sub);
    term2 = dereference_term(term2, *sub);
    printf("  Dereferenced: "); print_prolog_term(term1); printf(" with "); print_prolog_term(term2); printf("\n");

    if (term1->type == TERM_VARIABLE) {
        if (occur_check(term1, term2, *sub)) { printf("  FAIL: Occur check\n"); return false; } // Fail if X = f(X)
        *sub = add_binding(*sub, term1->value.variable_name, term2);
        printf("  BINDING: %s = ", term1->value.variable_name); print_prolog_term(term2); printf("\n");
        return true;
    } else if (term2->type == TERM_VARIABLE) {
        if (occur_check(term2, term1, *sub)) { printf("  FAIL: Occur check\n"); return false; } // Fail if X = f(X)
        *sub = add_binding(*sub, term2->value.variable_name, term1);
        printf("  BINDING: %s = ", term2->value.variable_name); print_prolog_term(term1); printf("\n");
        return true;
    } else if (term1->type == TERM_ATOM && term2->type == TERM_ATOM) {
        bool result = strcmp(term1->value.atom_name, term2->value.atom_name) == 0;
        printf("  ATOM_MATCH: %s\n", result ? "SUCCESS" : "FAIL");
        return result;
    } else if (term1->type == TERM_NUMBER && term2->type == TERM_NUMBER) {
        bool result = term1->value.number_value == term2->value.number_value;
        printf("  NUMBER_MATCH: %s\n", result ? "SUCCESS" : "FAIL");
        return result;
    } else if (term1->type == TERM_STRUCTURE && term2->type == TERM_STRUCTURE) {
        // Functor and arity must match
        if (strcmp(term1->value.structure->functor, term2->value.structure->functor) != 0 ||
            term1->value.structure->arity != term2->value.structure->arity) {
            printf("  STRUCT_MATCH: FAIL (functor/arity mismatch)\n");
            return false;
        }
        printf("  STRUCT_MATCH: Functor and arity match. Unifying args...\n");
        // Recursively unify arguments
        for (int i = 0; i < term1->value.structure->arity; i++) {
            if (!unify(term1->value.structure->args[i], term2->value.structure->args[i], sub)) {
                printf("  STRUCT_MATCH: FAIL (arg %d failed)\n", i);
                return false; // If any argument fails to unify, overall unification fails
            }
        }
        printf("  STRUCT_MATCH: SUCCESS\n");
        return true;
    }

    // If types don't match and not a variable, unification fails
    printf("  FAIL: Type mismatch or unsupported combination\n");
    return false;
}

// --- Term Copying and Renaming (for Fresh Variables) ---

// Deep copy of a PrologTerm
PrologTerm* copy_term(PrologTerm* original) {
    if (!original) return NULL;

    PrologTerm* new_term = create_term(original->type);
    switch (original->type) {
        case TERM_ATOM:
            new_term->value.atom_name = strdup(original->value.atom_name);
            break;
        case TERM_VARIABLE:
            new_term->value.variable_name = strdup(original->value.variable_name);
            break;
        case TERM_NUMBER:
            new_term->value.number_value = original->value.number_value;
            break;
        case TERM_STRUCTURE:
            new_term->value.structure = (PrologStructure*)malloc(sizeof(PrologStructure));
            if (!new_term->value.structure) { fprintf(stderr, "Memory allocation failed\n"); exit(1); }
            new_term->value.structure->functor = strdup(original->value.structure->functor);
            new_term->value.structure->arity = original->value.structure->arity;
            new_term->value.structure->args = (PrologTerm**)malloc(sizeof(PrologTerm*) * original->value.structure->arity);
            if (!new_term->value.structure->args) { fprintf(stderr, "Memory allocation failed\n"); exit(1); }
            for (int i = 0; i < original->value.structure->arity; i++) {
                new_term->value.structure->args[i] = copy_term(original->value.structure->args[i]);
            }
            break;
        case TERM_LIST:
            // TODO: Copy list elements
            break;
    }
    return new_term;
}

// Rename variables in a term with a unique ID suffix (e.g., X -> X_1)
void rename_variables(PrologTerm* term, int unique_id) {
    if (!term) return;

    char suffix[16];
    sprintf(suffix, "_%d", unique_id);
    int suffix_len = strlen(suffix);

    if (term->type == TERM_VARIABLE) {
        char *original_name = term->value.variable_name;
        int original_len = strlen(original_name);
        char *new_name = (char*)malloc(original_len + suffix_len + 1);
        if (!new_name) { fprintf(stderr, "Memory allocation failed\n"); exit(1); }
        strcpy(new_name, original_name);
        strcat(new_name, suffix);
        free(original_name); // Free the old name
        term->value.variable_name = new_name;
    } else if (term->type == TERM_STRUCTURE) {
        for (int i = 0; i < term->value.structure->arity; i++) {
            rename_variables(term->value.structure->args[i], unique_id);
        }
    }
    // Atoms and numbers don't have variables to rename
}

// --- Resolution Engine (for facts) ---

// Applies a substitution to a term (returns a new, substituted term)
PrologTerm* apply_substitution_to_term(PrologTerm* term, Substitution* sub) {
    if (!term) return NULL;

    PrologTerm* dereferenced_term = dereference_term(term, sub); // Dereference first

    if (dereferenced_term->type == TERM_VARIABLE) {
        // If it's still an unbound variable after dereferencing, just copy it
        return copy_term(dereferenced_term);
    } else if (dereferenced_term->type == TERM_STRUCTURE) {
        // Recursively apply to arguments
        PrologTerm* new_term = create_term(TERM_STRUCTURE);
        new_term->value.structure = (PrologStructure*)malloc(sizeof(PrologStructure));
        if (!new_term->value.structure) { fprintf(stderr, "Memory allocation failed\n"); exit(1); }
        new_term->value.structure->functor = strdup(dereferenced_term->value.structure->functor);
        new_term->value.structure->arity = dereferenced_term->value.structure->arity;
        new_term->value.structure->args = (PrologTerm**)malloc(sizeof(PrologTerm*) * dereferenced_term->value.structure->arity);
        if (!new_term->value.structure->args) { fprintf(stderr, "Memory allocation failed\n"); exit(1); }

        for (int i = 0; i < dereferenced_term->value.structure->arity; i++) {
            new_term->value.structure->args[i] = apply_substitution_to_term(dereferenced_term->value.structure->args[i], sub);
        }
        return new_term;
    } else { // Atom or Number
        return copy_term(dereferenced_term);
    }
}

// Main resolution function (for facts only, initially)
// Returns true if goal is satisfiable, false otherwise.
// If satisfiable, 'result_sub' contains the final substitution.
bool resolve(PrologTerm* goal, Substitution** result_sub) {
    if (goal->type != TERM_STRUCTURE) {
        fprintf(stderr, "Resolve Error: Goal must be a structure.\n");
        return false;
    }

    // Try to unify the goal with each clause in the database
    for (int i = 0; i < global_database->count; i++) {
        Clause* db_clause = global_database->clauses[i];

        // Create a temporary PrologTerm to wrap the db_clause->head for copy_term
        PrologTerm temp_db_head_term; 
        temp_db_head_term.type = TERM_STRUCTURE;
        temp_db_head_term.value.structure = db_clause->head;

        PrologTerm* fresh_head_term = copy_term(&temp_db_head_term); // Pass the wrapped term
        rename_variables(fresh_head_term, i + 1); // Rename variables (e.g., X -> X_1, Y -> Y_1)

        // Attempt unification
        Substitution* current_attempt_sub = *result_sub; // Start with the current substitution
        PrologTerm* current_goal = apply_substitution_to_term(goal, current_attempt_sub); // Apply current sub to goal

        printf("\nAttempting to resolve goal: "); print_prolog_term(current_goal); printf(" with clause: "); print_prolog_term(fresh_head_term); printf("\n");

        bool success = unify(current_goal, fresh_head_term, &current_attempt_sub);

        if (success) {
            *result_sub = current_attempt_sub; // Update the result_sub
            free_prolog_term(fresh_head_term); // Free the fresh copy
            free_prolog_term(current_goal); // Free the temporary goal
            return true; // Found a solution
        } else {
            // If unification fails, free the temporary substitution (only new bindings)
            // This part needs careful memory management to only free new bindings from current_attempt_sub
            // For now, we free the entire sub, which might be incorrect if it contains parent bindings.
            free_substitution(current_attempt_sub); // This might free too much if not careful
            free_prolog_term(fresh_head_term);
            free_prolog_term(current_goal);
        }
    }

    return false; // No solution found
}

int main() {
    printf("Hello, Prolog! (Interpreter under construction)\n");

    init_database();

    // Parse and add facts to the database
    printf("\n--- Loading Facts ---\n");

    input_buffer = "parent(john, mary). parent(mary, tom). "; // Example facts
    current_char_index = 0;
    get_next_token(); // Get the first token

    while (current_token.type != TOKEN_EOF) {
        PrologTerm* fact_term = parse_term(); // Use parse_term here
        if (fact_term->type != TERM_STRUCTURE) {
            parser_error("Expected a structure for a fact");
        }
        if (!match_token(TOKEN_DOT)) {
            parser_error("Expected '.' at end of fact");
        }
        Clause* fact_clause = (Clause*)malloc(sizeof(Clause));
        if (!fact_clause) { fprintf(stderr, "Memory allocation failed\n"); exit(1); }
        fact_clause->head = fact_term->value.structure; // Assign the actual structure
        fact_clause->body = NULL;
        fact_clause->body_count = 0;
        add_clause(fact_clause);
        printf("Loaded fact: ");
        print_prolog_term(fact_term);
        printf("\n");
        free(fact_term); // Free the term wrapper, but not the contained structure
    }
    printf("Facts loaded: %d\n", global_database->count);

    printf("\n--- Query Test ---\n");
    // Example Query: parent(john, X).
    input_buffer = "parent(john, X).";
    current_char_index = 0;
    has_next_token = false; // Reset lexer state
    get_next_token();

    PrologTerm* query_goal = parse_term();
    if (query_goal->type != TERM_STRUCTURE) {
        parser_error("Expected a structure for a query goal");
    }
    if (!match_token(TOKEN_DOT)) {
        parser_error("Expected '.' at end of query");
    }

    printf("Query: ");
    print_prolog_term(query_goal);
    printf("\n");

    Substitution* final_sub = NULL;
    if (resolve(query_goal, &final_sub)) {
        printf("Solution found!\n");
        Substitution* current = final_sub;
        while (current) {
            printf("  %s = ", current->variable_name);
            print_prolog_term(current->term);
            printf("\n");
            current = current->next;
        }
    } else {
        printf("No solution found.\n");
    }

    free_prolog_term(query_goal);
    free_substitution(final_sub);

    free_database();

    return 0;
}