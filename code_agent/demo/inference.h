#ifndef INFERENCE_H
#define INFERENCE_H

#include "knowledge_base.h"
#include "term.h"
#include "substitution.h"
#include "unification.h"
#include <stdbool.h>

// Function prototypes for inference engine
void resolve_query(KnowledgeBase *kb, Term *query_term, int *solution_count);
void resolve(KnowledgeBase *kb, Term **goals, int num_goals, Substitution *sub, int *var_counter, int *solution_count);

#endif // INFERENCE_H
