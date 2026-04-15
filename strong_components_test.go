package horizon

import (
	"testing"

	"github.com/LdDl/horizon/spatial"
)

func TestTarjanSCC_SimpleChain(t *testing.T) {
	engine := &MapEngine{
		edges: NewEdgeGraph(),
	}

	// Create a simple chain: 1 -> 2 -> 3
	// Each vertex is its own SCC (no cycles)
	engine.edges.Set(1, 2, &spatial.Edge{})
	engine.edges.Set(2, 3, &spatial.Edge{})

	result := engine.computeStrongConnectedComponents()

	// Should have 3 SCCs (each vertex is its own component)
	if result.TotalComponents != 3 {
		t.Errorf("Expected 3 SCCs, got %d", result.TotalComponents)
	}

	// Each vertex should be in a different component
	comp1 := result.VertexComponent[1]
	comp2 := result.VertexComponent[2]
	comp3 := result.VertexComponent[3]

	if comp1 == comp2 || comp2 == comp3 || comp1 == comp3 {
		t.Errorf("Each vertex should be in different SCC: %d, %d, %d", comp1, comp2, comp3)
	}

	// All components should be very small (size 1 < SMALL_COMPONENT_SIZE)
	for compID := range result.ComponentSizes {
		if !result.IsComponentVerySmall[compID] {
			t.Errorf("Component %d with size %d should be very small", compID, result.ComponentSizes[compID])
		}
	}
}

func TestTarjanSCC_SimpleCycle(t *testing.T) {
	engine := &MapEngine{
		edges: NewEdgeGraph(),
	}

	// Create a cycle: 1 -> 2 -> 3 -> 1
	// All vertices should be in one SCC
	engine.edges.Set(1, 2, &spatial.Edge{})
	engine.edges.Set(2, 3, &spatial.Edge{})
	engine.edges.Set(3, 1, &spatial.Edge{})

	result := engine.computeStrongConnectedComponents()

	// Should have 1 SCC
	if result.TotalComponents != 1 {
		t.Errorf("Expected 1 SCC, got %d", result.TotalComponents)
	}

	// All vertices should be in the same component
	comp1 := result.VertexComponent[1]
	comp2 := result.VertexComponent[2]
	comp3 := result.VertexComponent[3]

	if comp1 != comp2 || comp2 != comp3 {
		t.Errorf("All vertices should be in same SCC: %d, %d, %d", comp1, comp2, comp3)
	}

	// Component size should be 3
	if result.ComponentSizes[comp1] != 3 {
		t.Errorf("Expected component size 3, got %d", result.ComponentSizes[comp1])
	}
}

func TestTarjanSCC_TwoSeparateCycles(t *testing.T) {
	engine := &MapEngine{
		edges: NewEdgeGraph(),
	}

	// Create two separate cycles:
	// Cycle 1: 1 -> 2 -> 1
	// Cycle 2: 3 -> 4 -> 3
	engine.edges.Set(1, 2, &spatial.Edge{})
	engine.edges.Set(2, 1, &spatial.Edge{})
	engine.edges.Set(3, 4, &spatial.Edge{})
	engine.edges.Set(4, 3, &spatial.Edge{})

	result := engine.computeStrongConnectedComponents()

	// Should have 2 SCCs
	if result.TotalComponents != 2 {
		t.Errorf("Expected 2 SCCs, got %d", result.TotalComponents)
	}

	// Vertices 1,2 should be in same component
	if result.VertexComponent[1] != result.VertexComponent[2] {
		t.Errorf("Vertices 1 and 2 should be in same SCC")
	}

	// Vertices 3,4 should be in same component
	if result.VertexComponent[3] != result.VertexComponent[4] {
		t.Errorf("Vertices 3 and 4 should be in same SCC")
	}

	// But different from cycle 1
	if result.VertexComponent[1] == result.VertexComponent[3] {
		t.Errorf("Cycles 1-2 and 3-4 should be in different SCCs")
	}
}

func TestTarjanSCC_CycleWithTail(t *testing.T) {
	engine := &MapEngine{
		edges: NewEdgeGraph(),
	}

	// Create: 1 -> 2 -> 3 -> 2 (cycle 2-3) with entry from 1
	// Vertex 1 is separate, vertices 2,3 form an SCC
	engine.edges.Set(1, 2, &spatial.Edge{})
	engine.edges.Set(2, 3, &spatial.Edge{})
	engine.edges.Set(3, 2, &spatial.Edge{})

	result := engine.computeStrongConnectedComponents()

	// Should have 2 SCCs: {1} and {2,3}
	if result.TotalComponents != 2 {
		t.Errorf("Expected 2 SCCs, got %d", result.TotalComponents)
	}

	// Vertices 2,3 should be in same component
	if result.VertexComponent[2] != result.VertexComponent[3] {
		t.Errorf("Vertices 2 and 3 should be in same SCC")
	}

	// Vertex 1 should be in different component
	if result.VertexComponent[1] == result.VertexComponent[2] {
		t.Errorf("Vertex 1 should be in different SCC from 2,3")
	}
}

func TestTarjanSCC_BigComponent(t *testing.T) {
	engine := &MapEngine{
		edges: NewEdgeGraph(),
	}

	// Create a large cycle (size > SMALL_COMPONENT_SIZE would not be that small)
	// For test, create a cycle of 5 vertices and check bigComponentID
	for i := int64(1); i <= 5; i++ {
		next := i + 1
		if next > 5 {
			next = 1
		}
		engine.edges.Set(i, next, &spatial.Edge{})
	}

	result := engine.computeStrongConnectedComponents()

	// Should have 1 SCC with all 5 vertices
	if result.TotalComponents != 1 {
		t.Errorf("Expected 1 SCC, got %d", result.TotalComponents)
	}

	// Big component should have size 5
	if result.ComponentSizes[result.BigComponentID] != 5 {
		t.Errorf("Expected big component size 5, got %d", result.ComponentSizes[result.BigComponentID])
	}

	// Component should be very small (5 < SMALL_COMPONENT_SIZE)
	if !result.IsComponentVerySmall[result.BigComponentID] {
		t.Errorf("Component with size 5 should be very small")
	}
}
