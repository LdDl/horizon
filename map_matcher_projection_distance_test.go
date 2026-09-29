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

// projectionDistanceLine uses metres on a straight line, or arc metres on the equator.
func projectionDistanceLine(tb testing.TB, srid int, gap, costScale float64) (*MapMatcher, GPSMeasurements) {
	tb.Helper()
	xs := []float64{0, 100}
	if gap > 0 {
		xs = append(xs, 100+gap)
	}
	xs = append(xs, 200+gap, 300+gap)
	coordinate := func(x float64) float64 {
		if srid == 4326 {
			return x / spatial.EarthRadius * 180 / math.Pi
		}
		return x
	}
	point := func(x float64) s2.Point {
		if srid == 4326 {
			return spatial.NewWGS84Point(coordinate(x), 0).Point
		}
		return spatial.NewEuclideanS2Point(x, 0)
	}
	var graph ch.Graph
	vertices := make([]*spatial.Vertex, 0, len(xs))
	for i, x := range xs {
		if err := graph.CreateVertex(int64(i)); err != nil {
			tb.Fatal(err)
		}
		p := point(x)
		vertices = append(vertices, &spatial.Vertex{ID: int64(i), Point: &p})
	}
	edges := make([]*spatial.Edge, 0, len(xs)-1)
	for i := 0; i+1 < len(xs); i++ {
		length := xs[i+1] - xs[i]
		polyline := s2.Polyline{point(xs[i]), point(xs[i+1])}
		edge := &spatial.Edge{
			ID: int64(i), Source: int64(i), Target: int64(i + 1),
			LengthMeters: length, Weight: length * costScale, Polyline: &polyline,
		}
		if srid == 4326 {
			edge.PrecomputeCumLen()
			edge.PrecomputeBound()
		}
		edges = append(edges, edge)
		if err := graph.AddEdge(edge.Source, edge.Target, edge.Weight); err != nil {
			tb.Fatal(err)
		}
	}
	storageType := spatial.StorageTypeEuclidean
	if srid == 4326 {
		storageType = spatial.StorageTypeSpherical
	}
	engine := NewMapEngine(WithGraph(graph), WithStorage(spatial.NewStorage(storageType)), WithEdges(edges), WithVertices(vertices))
	engine.vertexComponent = engine.computeWeakConnectedComponents().VertexComponent
	matcher := NewMapMatcher(WithMapEngine(engine), WithHmmParameters(NewHmmProbabilities(1, 2)))
	observations := make(GPSMeasurements, 0, 3)
	for i, x := range []float64{70, 120 + gap, 220 + gap} {
		observations = append(observations, NewGPSMeasurement(i, coordinate(x), 0, srid, WithGPSTime(time.Unix(int64(i)*30, 0))))
	}
	return matcher, observations
}

type projectedRouteDistanceCase struct {
	name         string
	cost         float64
	connector    []int64
	sourceSuffix float64
	targetPrefix float64
	targetLength float64
	want         float64
}

func TestResolveRouteProjectionDistance(t *testing.T) {
	edges := NewEdgeGraph()
	edges.Set(1, 2, &spatial.Edge{LengthMeters: 20})
	edges.Set(2, 3, &spatial.Edge{LengthMeters: 30})
	matcher := &MapMatcher{engine: &MapEngine{edges: edges}}
	cases := []projectedRouteDistanceCase{
		{name: "30 plus 50 plus 20 metres", cost: 5, connector: []int64{1, 2, 3}, sourceSuffix: 30, targetPrefix: 20, targetLength: 100, want: 100},
		{name: "shared junction", connector: []int64{3}, sourceSuffix: 30, targetPrefix: 20, targetLength: 100, want: 50},
		{name: "both projections at junction", connector: []int64{3}, targetLength: 100, want: 0},
		{name: "large unused suffix preserves small distance", cost: 5, connector: []int64{1, 2, 3}, sourceSuffix: 30, targetPrefix: 20, targetLength: 1e20, want: 100},
		{name: "unreachable with positive offsets", cost: -1, sourceSuffix: 30, targetPrefix: 20, targetLength: 100, want: -1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			from := &RoadPosition{GraphEdge: &spatial.Edge{ID: 10, LengthMeters: tc.sourceSuffix}, afterProjection: tc.sourceSuffix}
			to := &RoadPosition{GraphEdge: &spatial.Edge{ID: 11, Source: 3, Target: 4, LengthMeters: tc.targetLength}, beforeProjection: tc.targetPrefix}
			if len(tc.connector) > 0 {
				from.GraphEdge.Target = tc.connector[0]
			}
			got, path := matcher.resolveRoute(tc.cost, tc.connector, from, to)
			if got != tc.want {
				t.Fatalf("distance = %g, want %g", got, tc.want)
			}
			var wantPath []int64
			if tc.cost >= 0 {
				wantPath = append(append([]int64(nil), tc.connector...), 4)
			}
			if !reflect.DeepEqual(path, wantPath) {
				t.Fatalf("response path = %v, want %v", path, wantPath)
			}
		})
	}
}

func TestTransitionProjectionDistanceInMeters(t *testing.T) {
	for _, srid := range []int{0, 4326} {
		for _, gap := range []float64{0, 50} {
			for _, scale := range []float64{1, 0.1} {
				t.Run(fmt.Sprintf("srid=%d/gap=%g/cost_scale=%g", srid, gap, scale), func(t *testing.T) {
					matcher, observations := projectionDistanceLine(t, srid, gap, scale)
					result, err := matcher.Run(observations, 0.1, 1)
					if err != nil {
						t.Fatal(err)
					}
					if len(result.SubMatches) != 1 || len(result.SubMatches[0].Observations) != 3 {
						t.Fatalf("expected one continuous three-observation match, got %+v", result)
					}
					wantEdges := []int64{0, 1, 2}
					if gap > 0 {
						wantEdges = []int64{0, 2, 3}
					}
					for i, obs := range result.SubMatches[0].Observations {
						if !obs.IsMatched || obs.MatchedEdge.ID != wantEdges[i] || obs.Observation != observations[i] {
							t.Fatalf("unexpected match for observation %d: edge=%d matched=%v", i, obs.MatchedEdge.ID, obs.IsMatched)
						}
					}
					// Both transitions follow the straight line between the observations.
					// Three emissions and two transitions contribute to the score.
					want := -1.5*math.Log(2*math.Pi) - 2*math.Log(2)
					if got := result.SubMatches[0].Probability; math.IsNaN(got) || math.Abs(got-want) > 1e-9 {
						t.Fatalf("score = %.12g, want %.12g; route offsets must use metres between projections", got, want)
					}
				})
			}
		}
	}
}
