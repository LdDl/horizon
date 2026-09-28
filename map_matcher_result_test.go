package horizon

import (
	"reflect"
	"testing"

	"github.com/LdDl/horizon/spatial"
	"github.com/LdDl/viterbi"
	"github.com/golang/geo/s2"
)

type intermediateEdgeCase struct {
	name   string
	states []int64
	paths  map[[2]int][]int64
	want   [][]int64
}

type intermediateEdgeFixture struct {
	matcher *MapMatcher
	path    viterbi.ViterbiPath
	gps     GPSMeasurements
}

func newIntermediateEdgeFixture(states []int64) intermediateEdgeFixture {
	edges := NewEdgeGraph()
	vertices := make(map[int64]*spatial.Vertex)
	byID := make(map[int64]*spatial.Edge)
	for _, definition := range []spatial.Edge{
		{ID: 10, Source: 0, Target: 1},
		{ID: 20, Source: 1, Target: 2},
		{ID: 30, Source: 2, Target: 1},
		{ID: 40, Source: 1, Target: 3},
		{ID: 50, Source: 3, Target: 1},
		{ID: 60, Source: 2, Target: 4},
	} {
		edge := definition
		line := s2.Polyline{
			spatial.NewEuclideanS2Point(float64(edge.Source), 0),
			spatial.NewEuclideanS2Point(float64(edge.Target), 0),
		}
		edge.Polyline = &line
		edge.Weight = float64(edge.ID)
		edges.Set(edge.Source, edge.Target, &edge)
		byID[edge.ID] = &edge
		vertices[edge.Source] = &spatial.Vertex{ID: edge.Source, Point: &line[0]}
		vertices[edge.Target] = &spatial.Vertex{ID: edge.Target, Point: &line[1]}
	}
	f := intermediateEdgeFixture{
		matcher: &MapMatcher{engine: &MapEngine{edges: edges, vertices: vertices}},
		path:    viterbi.ViterbiPath{Probability: -12},
	}
	for i, id := range states {
		edge := byID[id]
		x := (float64(edge.Source) + float64(edge.Target)) / 2
		position := NewRoadPositionFromLonLat(i, edge.Source, edge.Target, edge, x, 0, 0)
		position.next = 1
		f.path.Path = append(f.path.Path, position)
		f.gps = append(f.gps, NewGPSMeasurementFromID(i, x, 0, 0))
	}
	return f
}

func TestPrepareSubMatchIntermediateEdges(t *testing.T) {
	for _, test := range []intermediateEdgeCase{
		{"adjacent_edges_before_middle_observation", []int64{10, 20, 60}, map[[2]int][]int64{{0, 1}: {1, 2}, {1, 2}: {2, 4}}, [][]int64{nil, nil, nil}},
		{"adjacent_edges_before_last_observation", []int64{10, 20}, map[[2]int][]int64{{0, 1}: {1, 2}}, [][]int64{nil, nil}},
		{"intermediate_edges", []int64{10, 20, 60}, map[[2]int][]int64{{0, 1}: {1, 3, 1, 2}, {1, 2}: {2, 4}}, [][]int64{{40, 50}, nil, nil}},
		{"earlier_visit_and_reverse_edge", []int64{10, 20, 60}, map[[2]int][]int64{{0, 1}: {1, 2, 1, 2}, {1, 2}: {2, 4}}, [][]int64{{20, 30}, nil, nil}},
		{"last_connector_is_not_the_matched_edge", []int64{10, 20}, map[[2]int][]int64{{0, 1}: {1, 3, 1}}, [][]int64{{40, 50}, nil}},
		{"repeated_observations", []int64{10, 20, 20}, map[[2]int][]int64{{0, 1}: {1, 2}}, [][]int64{nil, nil, nil}},
		{"empty_route", []int64{10, 20}, nil, [][]int64{nil, nil}},
		{"singleton", []int64{10}, nil, [][]int64{nil}},
	} {
		t.Run(test.name, func(t *testing.T) {
			f := newIntermediateEdgeFixture(test.states)
			originalPaths := make(map[[2]int][]int64, len(test.paths))
			for key, route := range test.paths {
				originalPaths[key] = append([]int64(nil), route...)
			}
			got := f.matcher.prepareSubMatch(f.path, f.gps, nil, test.paths)
			if len(got.Observations) != len(test.states) || got.Probability != f.path.Probability {
				t.Fatal("response construction changed the matched path or score")
			}
			for i, observation := range got.Observations {
				state := f.path.Path[i].(*RoadPosition)
				if observation.MatchedEdge.ID != test.states[i] || observation.ProjectedPoint != state.Projected.Point || observation.MatchedVertex.ID != state.PickedGraphVertex {
					t.Fatalf("observation %d changed its matched position", i)
				}
				var ids []int64
				for _, edge := range observation.NextEdges {
					ids = append(ids, edge.ID)
					if edge.Weight != float64(edge.ID) || len(edge.Geom) != 2 {
						t.Fatal("intermediate edge metadata was changed")
					}
				}
				if !reflect.DeepEqual(ids, test.want[i]) {
					t.Errorf("observation %d: intermediate edges %v, want %v", i, ids, test.want[i])
				}
			}
			for key, route := range originalPaths {
				if !reflect.DeepEqual(test.paths[key], route) {
					t.Fatal("response construction changed the cached route")
				}
			}
		})
	}
}

func BenchmarkPrepareSubMatchIntermediateEdges(b *testing.B) {
	f := newIntermediateEdgeFixture([]int64{10, 20, 60})
	routes := map[[2]int][]int64{{0, 1}: {1, 3, 1, 2}, {1, 2}: {2, 4}}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.matcher.prepareSubMatch(f.path, f.gps, nil, routes)
	}
}
