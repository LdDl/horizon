package horizon

import (
	"math"
	"testing"

	"github.com/LdDl/ch"
	"github.com/LdDl/horizon/spatial"
	"github.com/golang/geo/s2"
)

// TestMapMatcherSubMatchesPlanar tests sub-matching with 4 disconnected road networks using Euclidean storage
/*
digraph G {

  subgraph cluster_0 {
    style=filled;
    color=lightgrey;
    node [style=filled,color=white];
    v0 -> v1 -> v2 -> v3;
    v2 -> v0;
    v3 -> v2;
    v3 -> v1;
    v0 -> v3;
    label = "Network 1";

    v0 [label="a0" pos="1,1!"];
    v1 [label="a1" pos="2,2!"];
    v2 [label="a2" pos="2.3,-0.7!"];
    v3 [label="a3" pos="3.4,1.5!"];
  }

  subgraph cluster_1 {
    style=filled;
    color=lightblue;
    node [style=filled,color=white];
    v4 -> v5 -> v6 -> v7;
    v7 -> v5;
    v7 -> v6;
    v6 -> v4;
    label = "Network 2";

    v4 [label="a4" pos="10,2!"];
    v5 [label="a5" pos="11,-1!"];
    v6 [label="a6" pos="12,2.3!"];
    v7 [label="a7" pos="13.5,2.1!"];
  }

  subgraph cluster_2 {
    style=filled;
    color=lightyellow;
    node [style=filled,color=white];
    v8 -> v9 -> v10 -> v11;
    v8 -> v11;
    v11 -> v9;
    label = "Network 3";

    v8 [label="a8" pos="6,-11!"];
    v9 [label="a9" pos="7,-7!"];
    v10 [label="a10" pos="8.9,-9.3!"];
    v11 [label="a11" pos="13,-11!"];
  }

  subgraph cluster_3 {
    style=filled;
    color=lightgreen;
    node [style=filled,color=white];
    v12 -> v13 -> v14 -> v15;
    v12 -> v14;
    v13 -> v15;
    v15 -> v12;
    label = "Network 4";

    v12 [label="a12" pos="-1,-5!"];
    v13 [label="a13" pos="1,-6.5!"];
    v14 [label="a14" pos="2,-3!"];
    v15 [label="a15" pos="3.9,-5!"];
  }

  obsA [shape=diamond, style=filled, fillcolor=yellow, label="A" pos="1.2,1.5!"];
  obsB [shape=diamond, style=filled, fillcolor=yellow, label="B" pos="3.1,0.3!"];
  obsC [shape=diamond, style=filled, fillcolor=yellow, label="C" pos="11.2,0.8!"];
  obsD [shape=diamond, style=filled, fillcolor=yellow, label="D" pos="6.1,-8.7!"];
  obsE [shape=diamond, style=filled, fillcolor=yellow, label="E" pos="8.2,-8.2!"];
  obsF [shape=diamond, style=filled, fillcolor=yellow, label="F" pos="10.9,-9.9!"];
  obsG [shape=diamond, style=filled, fillcolor=yellow, label="G" pos="1.9,-4.2!"];
}
*/
func TestMapMatcherSubMatchesPlanar(t *testing.T) {
	// Observations (GPS measurements) with SRID=0 for Euclidean
	gpsMeasurements := GPSMeasurements{
		NewGPSMeasurementFromID(1, 1.2, 1.5, 0),   // Near 0->3 edge
		NewGPSMeasurementFromID(2, 3.1, 0.3, 0),   // Near 3->2 edge
		NewGPSMeasurementFromID(3, 11.2, 0.8, 0),  // Near 5->6 edge
		NewGPSMeasurementFromID(4, 6.1, -8.7, 0),  // Near 8->9 edge
		NewGPSMeasurementFromID(5, 8.2, -8.2, 0),  // Near 9->10 edge
		NewGPSMeasurementFromID(6, 10.9, -9.9, 0), // Near 10->11 edge
		NewGPSMeasurementFromID(7, 1.9, -4.2, 0),  // Near 13->14 edge
	}

	// Vertex positions [x, y] format (Euclidean coordinates)
	vertices := map[int64][2]float64{
		// Network 1
		0: {1, 1},
		1: {2, 2},
		2: {2.3, -0.7},
		3: {3.4, 1.5},
		// Network 2
		4: {10, 2},
		5: {11, -1},
		6: {12, 2.3},
		7: {13.5, 2.1},
		// Network 3
		8:  {6, -11},
		9:  {7, -7},
		10: {8.9, -9.3},
		11: {13, -11},
		// Network 4
		12: {-1, -5},
		13: {1, -6.5},
		14: {2, -3},
		15: {3.9, -5},
	}

	// Edge definitions: ID, Source, Target
	type edgeDef struct {
		id     int64
		source int64
		target int64
	}
	edgeDefs := []edgeDef{
		// Network 1 (7 edges)
		{1, 0, 1}, {2, 1, 2}, {3, 2, 3}, {4, 2, 0}, {5, 3, 2}, {6, 3, 1}, {7, 0, 3},
		// Network 2 (6 edges)
		{8, 4, 5}, {9, 5, 6}, {10, 6, 7}, {11, 7, 5}, {12, 7, 6}, {13, 6, 4},
		// Network 3 (5 edges)
		{14, 8, 9}, {15, 9, 10}, {16, 10, 11}, {17, 8, 11}, {18, 11, 9},
		// Network 4 (6 edges)
		{19, 12, 13}, {20, 13, 14}, {21, 14, 15}, {22, 12, 14}, {23, 13, 15}, {24, 15, 12},
	}

	// Populate graph and spatial edges, vertices storage
	graph := ch.Graph{}
	edgesSpatial := []*spatial.Edge{}
	verticesSpatial := []*spatial.Vertex{}

	for vertexID, coords := range vertices {
		err := graph.CreateVertex(vertexID)
		if err != nil {
			t.Errorf("Can't add vertex with id = '%d' to the graph: %v", vertexID, err)
			return
		}
		// Create vertex with Euclidean point
		s2Point := spatial.NewEuclideanS2Point(coords[0], coords[1])
		verticesSpatial = append(verticesSpatial, &spatial.Vertex{
			Point: &s2Point,
			ID:    vertexID,
		})
	}

	for _, edge := range edgeDefs {
		source := vertices[edge.source]
		target := vertices[edge.target]

		// Euclidean distance
		dx := target[0] - source[0]
		dy := target[1] - source[1]
		weight := math.Sqrt(dx*dx + dy*dy)

		err := graph.AddEdge(edge.source, edge.target, weight)
		if err != nil {
			t.Errorf("Can't add edge from '%d' to '%d' to the graph: %v", edge.source, edge.target, err)
			return
		}

		s2Polyline := s2.Polyline{
			spatial.NewEuclideanS2Point(source[0], source[1]),
			spatial.NewEuclideanS2Point(target[0], target[1]),
		}
		edge := spatial.Edge{
			ID:       edge.id,
			Source:   edge.source,
			Target:   edge.target,
			Weight:   weight,
			Polyline: &s2Polyline,
		}
		edgesSpatial = append(edgesSpatial, &edge)
	}

	// Create Euclidean storage
	euclideanStorage := spatial.NewStorage(spatial.StorageTypeEuclidean)

	// Prepare engine
	mapEngine := NewMapEngine(
		WithGraph(graph),
		WithStorage(euclideanStorage),
		WithEdges(edgesSpatial),
		WithVertices(verticesSpatial),
	)

	// Create matcher and set engine
	sigma := 3.0 // Adjusted for Euclidean scale
	beta := 1.0  // Adjusted for Euclidean scale
	hmmParams := NewHmmProbabilities(sigma, beta)
	matcher := NewMapMatcher(
		WithHmmParameters(hmmParams),
		WithMapEngine(mapEngine),
	)

	// Define expected results
	// Expected sub-matches:
	// - SubMatch 0 (Network 1): A (0->3), B (3->2)
	// - SubMatch 1 (Network 2): C (5->6)
	// - SubMatch 2 (Network 3): D (8->9), E (9->10), F (10->11)
	// - SubMatch 3 (Network 4): G (13->14)
	expectedSubMatches := []struct {
		observations []struct {
			source int64
			target int64
		}
	}{
		{
			observations: []struct {
				source int64
				target int64
			}{
				{0, 3},
				{3, 2},
			},
		},
		{
			observations: []struct {
				source int64
				target int64
			}{
				{5, 6},
			},
		},
		{
			observations: []struct {
				source int64
				target int64
			}{
				{8, 9},
				{9, 10},
				{10, 11},
			},
		},
		{
			observations: []struct {
				source int64
				target int64
			}{
				{13, 14},
			},
		},
	}

	statesRadiusMeters := 5.0
	maxStates := 2
	result, err := matcher.Run(gpsMeasurements, statesRadiusMeters, maxStates)
	if err != nil {
		t.Error(err)
		return
	}

	// Check number of sub-matches
	if len(result.SubMatches) != len(expectedSubMatches) {
		t.Errorf("Expected %d sub-matches, got %d", len(expectedSubMatches), len(result.SubMatches))
		return
	}

	for s := range result.SubMatches {
		resultSubMatch := result.SubMatches[s]
		expectedSubMatch := expectedSubMatches[s]

		if len(resultSubMatch.Observations) != len(expectedSubMatch.observations) {
			t.Errorf("SubMatch %d: expected %d observations, got %d",
				s, len(expectedSubMatch.observations), len(resultSubMatch.Observations))
			continue
		}

		for i := range resultSubMatch.Observations {
			expected := expectedSubMatch.observations[i]
			actual := resultSubMatch.Observations[i]

			if actual.MatchedEdge.Source != expected.source ||
				actual.MatchedEdge.Target != expected.target {
				t.Errorf("SubMatch %d, observation %d: matched edge should be %d->%d, but got %d->%d",
					s, actual.Observation.id,
					expected.source, expected.target,
					actual.MatchedEdge.Source, actual.MatchedEdge.Target,
				)
			}
		}
	}
}
