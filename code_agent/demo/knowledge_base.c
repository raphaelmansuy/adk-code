#include "knowledge_base.h"
#include <stdlib.h>

KnowledgeBase *create_knowledge_base() {
    KnowledgeBase *kb = (KnowledgeBase *)malloc(sizeof(KnowledgeBase));
    kb->count = 0;
    kb->capacity = 10;
    kb->clauses = (Clause **)malloc(kb->capacity * sizeof(Clause *));
    return kb;
}

void add_clause(KnowledgeBase *kb, Clause *clause) {
    if (kb->count == kb->capacity) {
        kb->capacity *= 2;
        kb->clauses = (Clause **)realloc(kb->clauses, kb->capacity * sizeof(Clause *));
    }
    kb->clauses[kb->count++] = clause;
}

void free_knowledge_base(KnowledgeBase *kb) {
    if (!kb) return;
    for (int i = 0; i < kb->count; ++i) {
        free_clause(kb->clauses[i]);
    }
    free(kb->clauses);
    free(kb);
}
