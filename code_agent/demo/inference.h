#ifndef INFERENCE_H
#define INFERENCE_H

#include "knowledge_base.h"
#include "term.h"
#include "substitution.h"
#include "unification.h"
#include <stdbool.h>

// Function prototypes for inference engine
bool resolve_query(KnowledgeBase *kb, Term *query_term);
bool resolve(KnowledgeBase *kb, Term **goals, int num_goals, Substitution *sub, int *var_counter);

#endif // INFERENCE_H
