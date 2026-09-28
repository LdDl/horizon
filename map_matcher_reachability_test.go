package horizon

import (
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/LdDl/ch"
	"github.com/LdDl/horizon/spatial"
	"github.com/golang/geo/s2"
)

type reachabilityEdge struct {
	source int64
	target int64
	points [][2]float64
}

// newReachabilityEngine builds 0 <-> 1 -> 2 <-> 3 and a separate 4 -> 5.
// Reverse edges bend away from the forward edges to avoid nearest-candidate ties.
func newReachabilityEngine(tb testing.TB) *MapEngine {
	tb.Helper()
	vertices := [][2]float64{{0, 0}, {10, 0}, {20, 0}, {30, 0}, {0, 20}, {10, 20}}
	definitions := []reachabilityEdge{
		{source: 0, target: 1},
		{source: 1, target: 0, points: [][2]float64{{5, -3}}},
		{source: 1, target: 2},
		{source: 2, target: 3},
		{source: 3, target: 2, points: [][2]float64{{25, -3}}},
		{source: 4, target: 5},
	}
	var graph ch.Graph
	spatialVertices := make([]*spatial.Vertex, 0, len(vertices))
	for id, coords := range vertices {
		if err := graph.CreateVertex(int64(id)); err != nil {
			tb.Fatal(err)
		}
		point := spatial.NewEuclideanS2Point(coords[0], coords[1])
		spatialVertices = append(spatialVertices, &spatial.Vertex{ID: int64(id), Point: &point})
	}
	edges := make([]*spatial.Edge, 0, len(definitions))
	for id, definition := range definitions {
		coords := append([][2]float64{vertices[definition.source]}, definition.points...)
		coords = append(coords, vertices[definition.target])
		polyline := make(s2.Polyline, 0, len(coords))
		length := 0.0
		for i, point := range coords {
			polyline = append(polyline, spatial.NewEuclideanS2Point(point[0], point[1]))
			if i > 0 {
				length += math.Hypot(point[0]-coords[i-1][0], point[1]-coords[i-1][1])
			}
		}
		if err := graph.AddEdge(definition.source, definition.target, length); err != nil {
			tb.Fatal(err)
		}
		edges = append(edges, &spatial.Edge{
			ID: int64(id), Source: definition.source, Target: definition.target,
			Weight: length, LengthMeters: length, Polyline: &polyline,
		})
	}
	engine := NewMapEngine(
		WithGraph(graph),
		WithStorage(spatial.NewStorage(spatial.StorageTypeEuclidean)),
		WithEdges(edges),
		WithVertices(spatialVertices),
	)
	engine.vertexComponent = engine.computeWeakConnectedComponents().VertexComponent
	engine.vertexStrongComponent = engine.computeStrongConnectedComponents().VertexComponent
	if engine.vertexStrongComponent[0] == engine.vertexStrongComponent[3] || engine.vertexComponent[0] != engine.vertexComponent[3] {
		tb.Fatal("fixture must have distinct SCCs in one weak component")
	}
	return engine
}

type cachedPathReachabilityCase struct {
	name       string
	from       int64
	to         int64
	components map[int64]int64
	wantCost   float64
	wantPath   []int64
	wantCached bool
}

func TestCachedPathDirectedReachability(t *testing.T) {
	engine := newReachabilityEngine(t)
	cases := []cachedPathReachabilityCase{
		{name: "one way across SCCs", from: 0, to: 3, components: engine.vertexComponent, wantCost: 30, wantPath: []int64{0, 1, 2, 3}, wantCached: true},
		{name: "reverse unavailable", from: 3, to: 0, components: engine.vertexComponent, wantCost: -1, wantCached: true},
		{name: "within SCC", from: 0, to: 1, components: engine.vertexComponent, wantCost: 10, wantPath: []int64{0, 1}, wantCached: true},
		{name: "same vertex", from: 0, to: 0, components: engine.vertexComponent, wantPath: []int64{0}, wantCached: true},
		{name: "separate weak components", from: 0, to: 4, components: engine.vertexComponent, wantCost: -1},
		{name: "no component labels", from: 0, to: 3, wantCost: 30, wantPath: []int64{0, 1, 2, 3}, wantCached: true},
		{name: "missing target label", from: 0, to: 3, components: map[int64]int64{0: 9}, wantCost: 30, wantPath: []int64{0, 1, 2, 3}, wantCached: true},
		{name: "missing source label", from: 0, to: 3, components: map[int64]int64{3: 9}, wantCost: 30, wantPath: []int64{0, 1, 2, 3}, wantCached: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cache := make(map[[2]int64]cachedRoute)
			pool := engine.queryPool
			if !tc.wantCached {
				// Disconnected weak components must be rejected without a CH query.
				pool = nil
			}
			cost, path := getCachedPath(pool, cache, tc.components, tc.from, tc.to)
			if cost != tc.wantCost || !reflect.DeepEqual(path, tc.wantPath) {
				t.Fatalf("route = (%g, %v), want (%g, %v)", cost, path, tc.wantCost, tc.wantPath)
			}
			cached, exists := cache[[2]int64{tc.from, tc.to}]
			if exists != tc.wantCached {
				t.Fatalf("cache contains route = %v, want %v", exists, tc.wantCached)
			}
			if exists && (cached.cost != cost || !reflect.DeepEqual(cached.path, path)) {
				t.Fatal("cache differs from the CH result")
			}
			// A nil pool proves the second call uses the cached success or failure.
			cost, path = getCachedPath(nil, cache, tc.components, tc.from, tc.to)
			if cost != tc.wantCost || !reflect.DeepEqual(path, tc.wantPath) {
				t.Fatalf("repeated route = (%g, %v), want (%g, %v)", cost, path, tc.wantCost, tc.wantPath)
			}
		})
	}
}

type submatchReachabilityCase struct {
	name       string
	points     [][2]float64
	edges      []int64
	segments   []int
	omitLabels bool
}

func TestMapMatcherSubMatchesDirectedReachability(t *testing.T) {
	cases := []submatchReachabilityCase{
		{name: "forward across SCCs stays continuous", points: [][2]float64{{2, 0}, {22, 0}, {28, 0}}, edges: []int64{0, 3, 3}, segments: []int{3}},
		{name: "unreachable reverse splits", points: [][2]float64{{28, -1.2}, {2, -1.2}, {1, -0.6}}, edges: []int64{4, 1, 1}, segments: []int{1, 2}},
		{name: "separate network still splits", points: [][2]float64{{2, 0}, {22, 0}, {28, 0}, {2, 20}, {8, 20}}, edges: []int64{0, 3, 3, 5, 5}, segments: []int{3, 2}},
		{name: "no labels falls back to CH", points: [][2]float64{{2, 0}, {22, 0}, {28, 0}}, edges: []int64{0, 3, 3}, segments: []int{3}, omitLabels: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			engine := newReachabilityEngine(t)
			if tc.omitLabels {
				engine.vertexComponent = nil
				engine.vertexStrongComponent = nil
			}
			matcher := NewMapMatcher(WithMapEngine(engine), WithHmmParameters(NewHmmProbabilities(1, 2)))
			observations := make(GPSMeasurements, 0, len(tc.points))
			for i, point := range tc.points {
				observations = append(observations, NewGPSMeasurement(i, point[0], point[1], 0, WithGPSTime(time.Unix(int64(i)*30, 0))))
			}
			result, err := matcher.Run(observations, 0.1, 1)
			if err != nil {
				t.Fatal(err)
			}
			if len(result.SubMatches) != len(tc.segments) {
				t.Fatalf("submatches = %d, want %d", len(result.SubMatches), len(tc.segments))
			}
			index := 0
			for i, match := range result.SubMatches {
				if len(match.Observations) != tc.segments[i] {
					t.Fatalf("submatch %d has %d observations, want %d", i, len(match.Observations), tc.segments[i])
				}
				if math.IsNaN(match.Probability) || math.IsInf(match.Probability, 0) {
					t.Fatalf("non-finite score for submatch %d: %g", i, match.Probability)
				}
				for _, observation := range match.Observations {
					if !observation.IsMatched || observation.Observation != observations[index] || observation.MatchedEdge.ID != tc.edges[index] {
						t.Fatalf("unexpected match for observation %d: matched=%v edge=%d, want %d", index, observation.IsMatched, observation.MatchedEdge.ID, tc.edges[index])
					}
					index++
				}
			}
		})
	}
}
