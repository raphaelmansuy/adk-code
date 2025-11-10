#ifndef KNOWLEDGE_BASE_H
#define KNOWLEDGE_BASE_H

#include "clause.h"

// Knowledge Base (list of clauses)
typedef struct KnowledgeBase {
    Clause **clauses;
    int count;
    int capacity;
} KnowledgeBase;

// Function prototypes for knowledge base management
KnowledgeBase *create_knowledge_base();
void add_clause(KnowledgeBase *kb, Clause *clause);
void free_knowledge_base(KnowledgeBase *kb);

#endif // KNOWLEDGE_BASE_H
