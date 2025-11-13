// Package orchestration provides component dependency injection and initialization
// orchestration using the builder pattern.
//
// The Orchestrator builds all application components (display, model, agent, session)
// with proper dependency resolution and error handling. It ensures that components
// are initialized in the correct order and that dependencies are satisfied before
// use.
//
// The orchestration uses a fluent builder API that allows sequential component
// initialization with automatic error accumulation. If any step fails, the entire
// build fails with a clear error message.
//
// Component initialization order:
// 1. Display components (UI rendering, styling, formatting)
// 2. Model components (LLM provider initialization)
// 3. Agent components (tool registration, prompt building)
// 4. Session components (persistence, state management)
//
// Example:
//
//	components, err := orchestration.NewOrchestrator(ctx, cfg).
//		WithDisplay().
//		WithModel().
//		WithAgent().
//		WithSession().
//		Build()
//	if err != nil {
//		return nil, fmt.Errorf("orchestration failed: %w", err)
//	}
//
// The package contains:
// - Orchestrator: Main builder for component initialization
// - Components: Aggregated result of orchestration
// - Component-specific initializers (display.go, model.go, agent.go, session.go)
// - Factory functions for creating individual components
// - Builder utilities and error handling
package orchestration
