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
