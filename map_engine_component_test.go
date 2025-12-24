package horizon

import (
	"testing"
)

func TestWeakComponents_SmallGraph(t *testing.T) {
	graphFileName := "./test_data/matcher_4326_test.csv"

	hmmParams := NewHmmProbabilities(50.0, 2.0)
	matcher, err := NewMapMatcherFromFiles(hmmParams, graphFileName)
	if err != nil {
		t.Fatal(err)
	}

	engine := matcher.engine

	// Small graph should have 6 vertices
	if len(engine.vertexComponent) != 6 {
		t.Errorf("Expected 6 vertices, got %d", len(engine.vertexComponent))
	}

	// All vertices should be in component 0 (single connected component)
	expectedVertices := []int64{101, 102, 103, 104, 105, 106}
	for _, v := range expectedVertices {
		comp, exists := engine.vertexComponent[v]
		if !exists {
			t.Errorf("Vertex %d should have component assigned", v)
			continue
		}
		if comp != 0 {
			t.Errorf("Vertex %d should be in component 0, got %d", v, comp)
		}
	}

	// Big component ID should be 0
	if engine.bigComponentID != 0 {
		t.Errorf("Expected bigComponentID 0, got %d", engine.bigComponentID)
	}
}

func TestWeakComponents_BigGraph(t *testing.T) {
	graphFileName := "./test_data/osm2ch_export.csv"

	hmmParams := NewHmmProbabilities(50.0, 30.0)
	matcher, err := NewMapMatcherFromFiles(hmmParams, graphFileName)
	if err != nil {
		t.Fatal(err)
	}

	engine := matcher.engine

	// Big graph should have 40397 vertices
	if len(engine.vertexComponent) != 40397 {
		t.Errorf("Expected 40397 vertices, got %d", len(engine.vertexComponent))
	}

	// Big component ID should be 0
	if engine.bigComponentID != 0 {
		t.Errorf("Expected bigComponentID 0, got %d", engine.bigComponentID)
	}

	// Count vertices in big component
	bigComponentCount := 0
	for _, comp := range engine.vertexComponent {
		if comp == engine.bigComponentID {
			bigComponentCount++
		}
	}

	// Big component should have 40174 vertices
	if bigComponentCount != 40174 {
		t.Errorf("Expected 40174 vertices in big component, got %d", bigComponentCount)
	}

	// Count total components
	componentSet := make(map[int64]bool)
	for _, comp := range engine.vertexComponent {
		componentSet[comp] = true
	}

	// Should have 20 components
	if len(componentSet) != 20 {
		t.Errorf("Expected 20 components, got %d", len(componentSet))
	}
}

func TestStrongComponents_SmallGraph(t *testing.T) {
	graphFileName := "./test_data/matcher_4326_test.csv"

	hmmParams := NewHmmProbabilities(50.0, 2.0)
	matcher, err := NewMapMatcherFromFiles(hmmParams, graphFileName)
	if err != nil {
		t.Fatal(err)
	}

	engine := matcher.engine

	// Small graph should have 6 vertices
	if len(engine.vertexStrongComponent) != 6 {
		t.Errorf("Expected 6 vertices with SCC, got %d", len(engine.vertexStrongComponent))
	}

	// Count SCCs (no cycles in small graph, so each vertex is its own SCC)
	sccSet := make(map[int64]bool)
	for _, comp := range engine.vertexStrongComponent {
		sccSet[comp] = true
	}
	if len(sccSet) != 6 {
		t.Errorf("Expected 6 SCCs, got %d", len(sccSet))
	}

	// Big SCC should have size 1 (no cycles)
	bigSCCCount := 0
	for _, comp := range engine.vertexStrongComponent {
		if comp == engine.bigStrongComponentID {
			bigSCCCount++
		}
	}
	if bigSCCCount != 1 {
		t.Errorf("Expected big SCC size 1, got %d", bigSCCCount)
	}

	// All components should be very small (size < SMALL_COMPONENT_SIZE)
	for compID := range sccSet {
		if !engine.isComponentVerySmall[compID] {
			t.Errorf("Component %d should be very small", compID)
		}
	}
}

func TestStrongComponents_BigGraph(t *testing.T) {
	graphFileName := "./test_data/osm2ch_export.csv"

	hmmParams := NewHmmProbabilities(50.0, 30.0)
	matcher, err := NewMapMatcherFromFiles(hmmParams, graphFileName)
	if err != nil {
		t.Fatal(err)
	}

	engine := matcher.engine

	// Big graph should have 40397 vertices with SCC assigned
	if len(engine.vertexStrongComponent) != 40397 {
		t.Errorf("Expected 40397 vertices with SCC, got %d", len(engine.vertexStrongComponent))
	}

	// Count SCCs
	sccSet := make(map[int64]bool)
	for _, comp := range engine.vertexStrongComponent {
		sccSet[comp] = true
	}
	if len(sccSet) != 6520 {
		t.Errorf("Expected 6520 SCCs, got %d", len(sccSet))
	}

	// Count vertices in big SCC
	bigSCCCount := 0
	for _, comp := range engine.vertexStrongComponent {
		if comp == engine.bigStrongComponentID {
			bigSCCCount++
		}
	}
	if bigSCCCount != 33773 {
		t.Errorf("Expected big SCC size 33773, got %d", bigSCCCount)
	}

	// Count non-tiny SCCs (size >= SMALL_COMPONENT_SIZE)
	nonTinyCount := 0
	for compID := range sccSet {
		if !engine.isComponentVerySmall[compID] {
			nonTinyCount++
		}
	}
	if nonTinyCount != 1 {
		t.Errorf("Expected 1 non-tiny SCC, got %d", nonTinyCount)
	}
}
