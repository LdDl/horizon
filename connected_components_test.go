package horizon

import (
	"testing"

	"github.com/LdDl/horizon/spatial"
)

func TestBfsMarkWeakComponent(t *testing.T) {
	engine := &MapEngine{
		edges: make(map[int64]map[int64]*spatial.Edge),
	}

	// Create a simple graph:
	// Component 1: 1 -> 2 -> 3
	// Component 2: 4 -> 5 (isolated from component 1)
	engine.edges[1] = map[int64]*spatial.Edge{2: {}}
	engine.edges[2] = map[int64]*spatial.Edge{3: {}}
	engine.edges[4] = map[int64]*spatial.Edge{5: {}}

	visited := make(map[int64]bool)
	vertexComponent := make(map[int64]int64)

	// Mark component starting from vertex 1
	size1 := engine.bfsMarkWeakComponent(1, 0, visited, vertexComponent)
	if size1 != 3 {
		t.Errorf("Expected component size 3, got %d", size1)
	}

	// Verify all vertices in component 1 are marked
	for _, v := range []int64{1, 2, 3} {
		if vertexComponent[v] != 0 {
			t.Errorf("Vertex %d should be in component 0, got %d", v, vertexComponent[v])
		}
	}

	// Mark component starting from vertex 4
	size2 := engine.bfsMarkWeakComponent(4, 1, visited, vertexComponent)
	if size2 != 2 {
		t.Errorf("Expected component size 2, got %d", size2)
	}

	// Verify vertices in component 2 are marked
	for _, v := range []int64{4, 5} {
		if vertexComponent[v] != 1 {
			t.Errorf("Vertex %d should be in component 1, got %d", v, vertexComponent[v])
		}
	}
}

func TestBfsMarkWeakComponentUndirected(t *testing.T) {
	engine := &MapEngine{
		edges: make(map[int64]map[int64]*spatial.Edge),
	}

	// Create a graph where vertices are connected only via incoming edges:
	// 1 -> 2, 3 -> 2
	// Starting from vertex 3, we should still reach 1 and 2 (undirected traversal)
	engine.edges[1] = map[int64]*spatial.Edge{2: {}}
	engine.edges[3] = map[int64]*spatial.Edge{2: {}}

	visited := make(map[int64]bool)
	vertexComponent := make(map[int64]int64)

	// Start from vertex 2 - should find all connected vertices
	size := engine.bfsMarkWeakComponent(2, 0, visited, vertexComponent)
	if size != 3 {
		t.Errorf("Expected component size 3 (undirected), got %d", size)
	}

	// All vertices should be in the same component
	for _, v := range []int64{1, 2, 3} {
		if !visited[v] {
			t.Errorf("Vertex %d should be visited", v)
		}
		if vertexComponent[v] != 0 {
			t.Errorf("Vertex %d should be in component 0, got %d", v, vertexComponent[v])
		}
	}
}

func TestBfsMarkWeakComponentAlreadyVisited(t *testing.T) {
	engine := &MapEngine{
		edges: make(map[int64]map[int64]*spatial.Edge),
	}

	engine.edges[1] = map[int64]*spatial.Edge{2: {}}

	visited := make(map[int64]bool)
	vertexComponent := make(map[int64]int64)

	// Mark vertex 1 as already visited
	visited[1] = true

	// Should return 0 since start vertex is already visited
	size := engine.bfsMarkWeakComponent(1, 0, visited, vertexComponent)
	if size != 0 {
		t.Errorf("Expected size 0 for already visited start, got %d", size)
	}
}
