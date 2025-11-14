package agents

import (
	"testing"
)

// TestNewDependencyGraph tests creating a new dependency graph
func TestNewDependencyGraph(t *testing.T) {
	dg := NewDependencyGraph()
	if dg == nil {
		t.Fatal("Expected non-nil graph")
	}

	if len(dg.Agents) != 0 {
		t.Error("Expected empty agents map")
	}

	if len(dg.Edges) != 0 {
		t.Error("Expected empty edges map")
	}
}

// TestAddAgent tests adding agents to the graph
func TestAddAgent(t *testing.T) {
	dg := NewDependencyGraph()

	agent := &Agent{
		Name:        "test-agent",
		Description: "Test agent",
	}

	err := dg.AddAgent(agent)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if dg.AgentCount() != 1 {
		t.Errorf("Expected 1 agent, got %d", dg.AgentCount())
	}
}

// TestAddAgentNil tests adding nil agent
func TestAddAgentNil(t *testing.T) {
	dg := NewDependencyGraph()
	err := dg.AddAgent(nil)
	if err == nil {
		t.Error("Expected error for nil agent")
	}
}

// TestAddAgentEmptyName tests adding agent with empty name
func TestAddAgentEmptyName(t *testing.T) {
	dg := NewDependencyGraph()
	agent := &Agent{Name: ""}
	err := dg.AddAgent(agent)
	if err == nil {
		t.Error("Expected error for empty name")
	}
}

// TestAddAgentDuplicate tests adding duplicate agent
func TestAddAgentDuplicate(t *testing.T) {
	dg := NewDependencyGraph()
	agent := &Agent{Name: "agent1"}

	err := dg.AddAgent(agent)
	if err != nil {
		t.Fatalf("First add failed: %v", err)
	}

	err = dg.AddAgent(agent)
	if err == nil {
		t.Error("Expected error for duplicate agent")
	}
}

// TestAddEdge tests adding dependency edges
func TestAddEdge(t *testing.T) {
	dg := NewDependencyGraph()

	agent1 := &Agent{Name: "agent1"}
	agent2 := &Agent{Name: "agent2"}

	dg.AddAgent(agent1)
	dg.AddAgent(agent2)

	err := dg.AddEdge("agent1", "agent2")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if dg.EdgeCount() != 1 {
		t.Errorf("Expected 1 edge, got %d", dg.EdgeCount())
	}
}

// TestAddEdgeNonexistentAgent tests adding edge with nonexistent agent
func TestAddEdgeNonexistentAgent(t *testing.T) {
	dg := NewDependencyGraph()
	agent := &Agent{Name: "agent1"}
	dg.AddAgent(agent)

	err := dg.AddEdge("agent1", "nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent agent")
	}
}

// TestAddEdgeDuplicate tests adding duplicate edge
func TestAddEdgeDuplicate(t *testing.T) {
	dg := NewDependencyGraph()

	agent1 := &Agent{Name: "agent1"}
	agent2 := &Agent{Name: "agent2"}

	dg.AddAgent(agent1)
	dg.AddAgent(agent2)

	err := dg.AddEdge("agent1", "agent2")
	if err != nil {
		t.Fatalf("First edge failed: %v", err)
	}

	err = dg.AddEdge("agent1", "agent2")
	if err == nil {
		t.Error("Expected error for duplicate edge")
	}
}

// TestResolveDependencies tests topological sorting
func TestResolveDependencies(t *testing.T) {
	dg := NewDependencyGraph()

	// Create agents
	a1 := &Agent{Name: "a1"}
	a2 := &Agent{Name: "a2"}
	a3 := &Agent{Name: "a3"}

	dg.AddAgent(a1)
	dg.AddAgent(a2)
	dg.AddAgent(a3)

	// Create dependencies: a3 depends on a2, a2 depends on a1
	dg.AddEdge("a3", "a2")
	dg.AddEdge("a2", "a1")

	// Resolve a3's dependencies
	deps, err := dg.ResolveDependencies("a3")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(deps) != 3 {
		t.Errorf("Expected 3 deps, got %d", len(deps))
	}

	// Should be in order: a1, a2, a3
	if deps[0].Name != "a1" {
		t.Errorf("Expected first dep a1, got %s", deps[0].Name)
	}
	if deps[1].Name != "a2" {
		t.Errorf("Expected second dep a2, got %s", deps[1].Name)
	}
	if deps[2].Name != "a3" {
		t.Errorf("Expected third dep a3, got %s", deps[2].Name)
	}
}

// TestResolveDependenciesNonexistent tests with nonexistent agent
func TestResolveDependenciesNonexistent(t *testing.T) {
	dg := NewDependencyGraph()
	_, err := dg.ResolveDependencies("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent agent")
	}
}

// TestGetTransitiveDeps tests transitive dependency collection
func TestGetTransitiveDeps(t *testing.T) {
	dg := NewDependencyGraph()

	// Create agents
	a1 := &Agent{Name: "a1"}
	a2 := &Agent{Name: "a2"}
	a3 := &Agent{Name: "a3"}

	dg.AddAgent(a1)
	dg.AddAgent(a2)
	dg.AddAgent(a3)

	// Create dependencies: a3 -> a2, a2 -> a1
	dg.AddEdge("a3", "a2")
	dg.AddEdge("a2", "a1")

	// Get transitive deps of a3
	deps, err := dg.GetTransitiveDeps("a3")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(deps) != 2 {
		t.Errorf("Expected 2 deps, got %d", len(deps))
	}

	// Should include both a1 and a2
	if deps[0] != "a1" || deps[1] != "a2" {
		t.Errorf("Expected [a1, a2], got %v", deps)
	}
}

// TestDetectCycles tests cycle detection
func TestDetectCycles(t *testing.T) {
	dg := NewDependencyGraph()

	// Create agents with cycle
	a1 := &Agent{Name: "a1"}
	a2 := &Agent{Name: "a2"}

	dg.AddAgent(a1)
	dg.AddAgent(a2)

	// Create cycle: a1 -> a2 -> a1
	dg.AddEdge("a1", "a2")
	dg.AddEdge("a2", "a1")

	cycles := dg.DetectCycles()
	if len(cycles) == 0 {
		t.Error("Expected to detect cycles")
	}
}

// TestDetectCyclesNoCycle tests with no cycles
func TestDetectCyclesNoCycle(t *testing.T) {
	dg := NewDependencyGraph()

	a1 := &Agent{Name: "a1"}
	a2 := &Agent{Name: "a2"}

	dg.AddAgent(a1)
	dg.AddAgent(a2)

	dg.AddEdge("a1", "a2")

	cycles := dg.DetectCycles()
	if len(cycles) != 0 {
		t.Errorf("Expected no cycles, got %v", cycles)
	}
}

// TestGetAgent tests retrieving agent by name
func TestGetAgent(t *testing.T) {
	dg := NewDependencyGraph()
	agent := &Agent{Name: "test"}
	dg.AddAgent(agent)

	retrieved := dg.GetAgent("test")
	if retrieved == nil {
		t.Error("Expected non-nil agent")
	}

	if retrieved.Name != "test" {
		t.Errorf("Expected name 'test', got %q", retrieved.Name)
	}
}

// TestGetAgentNonexistent tests retrieving nonexistent agent
func TestGetAgentNonexistent(t *testing.T) {
	dg := NewDependencyGraph()
	agent := dg.GetAgent("nonexistent")
	if agent != nil {
		t.Error("Expected nil for nonexistent agent")
	}
}

// TestAgentCount tests agent count
func TestAgentCount(t *testing.T) {
	dg := NewDependencyGraph()

	if dg.AgentCount() != 0 {
		t.Error("Expected 0 agents initially")
	}

	dg.AddAgent(&Agent{Name: "a1"})
	if dg.AgentCount() != 1 {
		t.Error("Expected 1 agent")
	}

	dg.AddAgent(&Agent{Name: "a2"})
	if dg.AgentCount() != 2 {
		t.Error("Expected 2 agents")
	}
}

// TestEdgeCount tests edge count
func TestEdgeCount(t *testing.T) {
	dg := NewDependencyGraph()

	a1 := &Agent{Name: "a1"}
	a2 := &Agent{Name: "a2"}
	a3 := &Agent{Name: "a3"}

	dg.AddAgent(a1)
	dg.AddAgent(a2)
	dg.AddAgent(a3)

	if dg.EdgeCount() != 0 {
		t.Error("Expected 0 edges initially")
	}

	dg.AddEdge("a1", "a2")
	if dg.EdgeCount() != 1 {
		t.Error("Expected 1 edge")
	}

	dg.AddEdge("a1", "a3")
	if dg.EdgeCount() != 2 {
		t.Error("Expected 2 edges")
	}
}

// TestGetAllAgents tests retrieving all agents
func TestGetAllAgents(t *testing.T) {
	dg := NewDependencyGraph()

	dg.AddAgent(&Agent{Name: "b"})
	dg.AddAgent(&Agent{Name: "a"})
	dg.AddAgent(&Agent{Name: "c"})

	agents := dg.GetAllAgents()
	if len(agents) != 3 {
		t.Errorf("Expected 3 agents, got %d", len(agents))
	}

	// Should be sorted by name
	if agents[0].Name != "a" || agents[1].Name != "b" || agents[2].Name != "c" {
		t.Errorf("Expected sorted agents, got %v", agents)
	}
}

// TestResolveConflicts tests conflict resolution (placeholder)
func TestResolveConflicts(t *testing.T) {
	dg := NewDependencyGraph()
	err := dg.ResolveConflicts()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}
