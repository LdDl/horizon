package horizon

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/LdDl/ch"
	"github.com/LdDl/horizon/spatial"
	"github.com/golang/geo/s2"
)

type sameEdgeDefinition struct {
	source, target int64
	coordinates    [][2]float64
}

type sameEdgeCase struct {
	name          string
	edges         []sameEdgeDefinition
	observations  [][2]float64
	routeLengths  []float64
	groups        [][]int
	nextEdges     [][]int64
	matchedEdge   int64
	maxCandidates int
}

// Coordinates are metres in the plane, or equatorial arc metres converted to degrees.
func sameEdgeFixture(t *testing.T, tc sameEdgeCase, srid int, costScale float64) (*MapMatcher, GPSMeasurements) {
	t.Helper()
	coordinate := func(x float64) float64 {
		if srid == 4326 {
			return x / spatial.EarthRadius * 180 / math.Pi
		}
		return x
	}
	point := func(xy [2]float64) s2.Point {
		if srid == 4326 {
			return spatial.NewWGS84Point(coordinate(xy[0]), coordinate(xy[1])).Point
		}
		return spatial.NewEuclideanS2Point(xy[0], xy[1])
	}
	var graph ch.Graph
	vertices := make([]*spatial.Vertex, 0)
	seen := make(map[int64]bool)
	for _, definition := range tc.edges {
		for i, id := range []int64{definition.source, definition.target} {
			if seen[id] {
				continue
			}
			seen[id] = true
			if err := graph.CreateVertex(id); err != nil {
				t.Fatal(err)
			}
			p := point(definition.coordinates[i*(len(definition.coordinates)-1)])
			vertices = append(vertices, &spatial.Vertex{ID: id, Point: &p})
		}
	}
	edges := make([]*spatial.Edge, 0, len(tc.edges))
	for id, definition := range tc.edges {
		line := make(s2.Polyline, len(definition.coordinates))
		length := 0.0
		for i, xy := range definition.coordinates {
			line[i] = point(xy)
			if i > 0 {
				if srid == 4326 {
					length += line[i-1].Distance(line[i]).Radians() * spatial.EarthRadius
				} else {
					length += line[i].Sub(line[i-1].Vector).Norm()
				}
			}
		}
		edge := &spatial.Edge{ID: int64(id), Source: definition.source, Target: definition.target, Polyline: &line, LengthMeters: length, Weight: length * costScale}
		if srid == 4326 {
			edge.PrecomputeCumLen()
			edge.PrecomputeBound()
		}
		edges = append(edges, edge)
		if err := graph.AddEdge(edge.Source, edge.Target, edge.Weight); err != nil {
			t.Fatal(err)
		}
	}
	storageType := spatial.StorageTypeEuclidean
	if srid == 4326 {
		storageType = spatial.StorageTypeSpherical
	}
	engine := NewMapEngine(WithGraph(graph), WithStorage(spatial.NewStorage(storageType)), WithEdges(edges), WithVertices(vertices))
	engine.vertexComponent = engine.computeWeakConnectedComponents().VertexComponent
	matcher := NewMapMatcher(WithMapEngine(engine), WithHmmParameters(NewHmmProbabilities(1, 2)))
	observations := make(GPSMeasurements, len(tc.observations))
	for i, xy := range tc.observations {
		observations[i] = NewGPSMeasurement(i, coordinate(xy[0]), coordinate(xy[1]), srid, WithGPSTime(time.Unix(int64(i)*30, 0)))
	}
	return matcher, observations
}

func TestSameEdgeTransitions(t *testing.T) {
	straight := sameEdgeDefinition{0, 1, [][2]float64{{0, 0}, {100, 0}}}
	bend := sameEdgeDefinition{0, 1, [][2]float64{{0, 0}, {100, 0}, {100, 100}}}
	loop := sameEdgeDefinition{0, 0, [][2]float64{{0, 0}, {100, 0}, {100, 100}, {0, 100}, {0, 0}}}
	cycle := []sameEdgeDefinition{
		straight,
		{1, 2, [][2]float64{{100, 0}, {100, -100}}},
		{2, 3, [][2]float64{{100, -100}, {0, -100}}},
		{3, 0, [][2]float64{{0, -100}, {0, 0}}},
	}
	cases := []sameEdgeCase{
		{name: "forward then stop", edges: []sameEdgeDefinition{straight}, observations: [][2]float64{{20, 0}, {70, 0}, {70, 0}}, routeLengths: []float64{50, 0}},
		{name: "stop then forward", edges: []sameEdgeDefinition{straight}, observations: [][2]float64{{20, 0}, {20, 0}, {70, 0}}, routeLengths: []float64{0, 50}},
		{name: "bend in first transition", edges: []sameEdgeDefinition{bend}, observations: [][2]float64{{90, 0}, {100, 10}, {100, 20}}, routeLengths: []float64{20, 10}},
		{name: "bend in later transition", edges: []sameEdgeDefinition{bend}, observations: [][2]float64{{80, 0}, {90, 0}, {100, 10}}, routeLengths: []float64{10, 20}},
		{name: "return cycle first", edges: cycle, observations: [][2]float64{{70, 0}, {20, 0}, {30, 0}}, routeLengths: []float64{350, 10}, nextEdges: [][]int64{{1, 2, 3}, nil, nil}},
		{name: "return cycle later", edges: cycle, observations: [][2]float64{{20, 0}, {70, 0}, {30, 0}}, routeLengths: []float64{50, 360}, nextEdges: [][]int64{nil, {1, 2, 3}, nil}},
		{name: "backward without cycle first", edges: []sameEdgeDefinition{straight}, observations: [][2]float64{{70, 0}, {20, 0}, {30, 0}}, routeLengths: []float64{0, 10}, groups: [][]int{{0}, {1, 2}}},
		{name: "backward without cycle later", edges: []sameEdgeDefinition{straight}, observations: [][2]float64{{20, 0}, {70, 0}, {30, 0}}, routeLengths: []float64{50, 0}, groups: [][]int{{0, 1}, {2}}},
		{name: "self loop forward bend", edges: []sameEdgeDefinition{loop}, observations: [][2]float64{{90, 0}, {100, 10}, {100, 20}}, routeLengths: []float64{20, 10}},
		{name: "self loop wraps", edges: []sameEdgeDefinition{loop}, observations: [][2]float64{{70, 0}, {20, 0}, {30, 0}}, routeLengths: []float64{350, 10}},
		{name: "reverse directed edge", edges: []sameEdgeDefinition{straight, {1, 0, [][2]float64{{100, 0}, {0, 0}}}}, observations: [][2]float64{{70, 0}, {20, 0}, {10, 0}}, routeLengths: []float64{50, 10}, matchedEdge: 1, maxCandidates: 2},
	}
	for _, tc := range cases {
		for _, srid := range []int{0, 4326} {
			for _, scale := range []float64{1, 0.1} {
				t.Run(fmt.Sprintf("%s/srid=%d/cost_scale=%g", tc.name, srid, scale), func(t *testing.T) {
					matcher, observations := sameEdgeFixture(t, tc, srid, scale)
					k := tc.maxCandidates
					if k == 0 {
						k = 1
					}
					result, err := matcher.Run(observations, 0.1, k)
					if err != nil {
						t.Fatal(err)
					}
					groups := tc.groups
					if groups == nil {
						groups = [][]int{{0, 1, 2}}
					}
					if len(result.SubMatches) != len(groups) {
						t.Fatalf("submatches = %d, want %d", len(result.SubMatches), len(groups))
					}
					for groupIndex, group := range groups {
						subMatch := result.SubMatches[groupIndex]
						if len(subMatch.Observations) != len(group) {
							t.Fatalf("submatch %d has %d observations, want %d", groupIndex, len(subMatch.Observations), len(group))
						}
						// Each observation lies on its edge and contributes one emission.
						wantScore := -float64(len(group)) * math.Log(2*math.Pi) / 2
						for i, obsIndex := range group {
							observation := subMatch.Observations[i]
							if !observation.IsMatched || observation.MatchedEdge.ID != tc.matchedEdge || observation.Observation != observations[obsIndex] {
								t.Fatalf("unexpected match for observation %d: edge=%d matched=%v", obsIndex, observation.MatchedEdge.ID, observation.IsMatched)
							}
							var gotNext, wantNext []int64
							for _, edge := range observation.NextEdges {
								gotNext = append(gotNext, edge.ID)
							}
							if tc.nextEdges != nil {
								wantNext = tc.nextEdges[obsIndex]
							}
							if !reflect.DeepEqual(gotNext, wantNext) {
								t.Errorf("observation %d: next edges = %v, want %v", obsIndex, gotNext, wantNext)
							}
							if i > 0 {
								chord := observations[obsIndex-1].GeoPoint.DistanceTo(observations[obsIndex].GeoPoint)
								wantScore -= math.Log(2) + math.Abs(tc.routeLengths[obsIndex-1]-chord)/2
							}
						}
						if math.IsNaN(subMatch.Probability) || math.Abs(subMatch.Probability-wantScore) > 1e-7 {
							t.Errorf("score = %.12f, want %.12f from arc lengths %v", subMatch.Probability, wantScore, tc.routeLengths)
						}
						if tc.nextEdges != nil {
							// A return through other edges must separate the two clipped visits.
							geometry := subMatch.ResponseGeometries()
							first, last := geometry[0].Matched, geometry[len(geometry)-1].Matched
							if len(first) == 0 || first[0] != subMatch.Observations[0].ProjectedPoint ||
								len(last) == 0 || last[len(last)-1] != subMatch.Observations[len(group)-1].ProjectedPoint {
								t.Error("return route lost the clipping boundaries of separate edge visits")
							}
						}
					}
					// Repeat the request to detect mutation of a stored edge or cached path.
					repeated, err := matcher.Run(observations, 0.1, k)
					if err != nil || !reflect.DeepEqual(result, repeated) {
						t.Fatalf("identical request changed its result: %v", err)
					}
				})
			}
		}
	}
}
