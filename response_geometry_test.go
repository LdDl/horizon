package horizon

import (
	"reflect"
	"testing"

	"github.com/LdDl/horizon/spatial"
	"github.com/golang/geo/s2"
)

type responseGeometryCase struct {
	name         string
	observations []ObservationResult
	matched      []s2.Polyline
	cuts         []bool
}

func responseTestLine(points ...[2]float64) s2.Polyline {
	line := make(s2.Polyline, len(points))
	for i, point := range points {
		line[i] = spatial.NewEuclideanS2Point(point[0], point[1])
	}
	return line
}

func responseTestObservation(edge spatial.Edge, x, y float64) ObservationResult {
	point := spatial.NewEuclideanS2Point(x, y)
	projected, _, next := spatial.CalcProjectionEuclidean(*edge.Polyline, point)
	return ObservationResult{
		Observation: NewGPSMeasurementFromID(0, x, y, 0), IsMatched: true,
		MatchedEdge: edge, ProjectedPoint: projected, ProjectionPointIdx: next,
	}
}

func TestResponseGeometriesBoundaryVisits(t *testing.T) {
	line := responseTestLine([2]float64{0, 0}, [2]float64{0, 10}, [2]float64{10, 10}, [2]float64{20, 10})
	otherLine := responseTestLine([2]float64{20, 10}, [2]float64{30, 10}, [2]float64{30, 20})
	first := spatial.Edge{ID: 1, Polyline: &line}
	last := spatial.Edge{ID: 2, Polyline: &otherLine}
	a := responseTestObservation(first, 5, 10)
	b := responseTestObservation(first, 10, 10)
	c := responseTestObservation(first, 15, 10)
	d := responseTestObservation(last, 25, 10)
	e := responseTestObservation(last, 30, 15)
	leading := responseTestLine([2]float64{5, 10}, [2]float64{10, 10}, [2]float64{20, 10})
	trailing := responseTestLine([2]float64{20, 10}, [2]float64{30, 10}, [2]float64{30, 15})
	interval := responseTestLine([2]float64{5, 10}, [2]float64{10, 10}, [2]float64{15, 10})
	stationary := responseTestLine([2]float64{5, 10}, [2]float64{5, 10})
	for _, test := range []responseGeometryCase{
		{"two_runs", []ObservationResult{a, c, d, e}, []s2.Polyline{leading, leading, trailing, trailing}, []bool{true, false, false, true}},
		{"one_edge", []ObservationResult{a, b, c}, []s2.Polyline{interval, interval, interval}, []bool{true, false, true}},
		{"stationary", []ObservationResult{a, a, a}, []s2.Polyline{stationary, stationary, stationary}, []bool{true, false, true}},
		{"backtrack", []ObservationResult{a, c, b}, []s2.Polyline{line, line, line}, []bool{false, false, false}},
		{"return_to_edge", []ObservationResult{a, d, c}, []s2.Polyline{leading, otherLine, responseTestLine([2]float64{0, 0}, [2]float64{0, 10}, [2]float64{10, 10}, [2]float64{15, 10})}, []bool{true, false, true}},
		{"singleton", []ObservationResult{a}, []s2.Polyline{leading}, []bool{true}},
		{"unmatched", []ObservationResult{{}}, []s2.Polyline{nil}, []bool{false}},
		{"unmatched_end", []ObservationResult{a, c, {}}, []s2.Polyline{leading, leading, nil}, []bool{true, false, false}},
		{"unmatched_start", []ObservationResult{{}, d, e}, []s2.Polyline{nil, trailing, trailing}, []bool{false, false, true}},
		{"empty", nil, nil, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			original := append(s2.Polyline(nil), line...)
			otherOriginal := append(s2.Polyline(nil), otherLine...)
			match := SubMatch{Observations: test.observations, Probability: -12}
			got := match.ResponseGeometries()
			if len(got) != len(test.matched) {
				t.Fatalf("got %d geometries, want %d", len(got), len(test.matched))
			}
			for i, geometry := range got {
				if !reflect.DeepEqual(geometry.Matched, test.matched[i]) {
					t.Fatalf("observation %d: got %v, want %v", i, geometry.Matched, test.matched[i])
				}
				if (geometry.Cut != nil) != test.cuts[i] {
					t.Fatalf("observation %d: unexpected cut presence", i)
				}
			}
			if !reflect.DeepEqual(line, original) || !reflect.DeepEqual(otherLine, otherOriginal) || match.Probability != -12 {
				t.Fatal("response preparation changed its input")
			}
			if !reflect.DeepEqual(got, match.ResponseGeometries()) {
				t.Fatal("repeated response preparation changed the geometry")
			}
		})
	}
}

func TestResponseGeometriesSameSegmentAndEndpoint(t *testing.T) {
	line := responseTestLine([2]float64{0, 0}, [2]float64{100, 0})
	edge := spatial.Edge{ID: 1, Polyline: &line}
	match := SubMatch{Observations: []ObservationResult{
		responseTestObservation(edge, 30, 0), responseTestObservation(edge, 50, 0), responseTestObservation(edge, 80, 0),
	}}
	got := match.ResponseGeometries()
	want := responseTestLine([2]float64{30, 0}, [2]float64{80, 0})
	for i := range got {
		if !reflect.DeepEqual(got[i].Matched, want) {
			t.Fatalf("observation %d redrew an unobserved part: %v", i, got[i].Matched)
		}
	}
	if !reflect.DeepEqual(got[0].Cut, responseTestLine([2]float64{0, 0}, [2]float64{30, 0})) ||
		!reflect.DeepEqual(got[2].Cut, responseTestLine([2]float64{80, 0}, [2]float64{100, 0})) {
		t.Fatal("removed geometry lost its boundary location")
	}
	end := responseTestObservation(edge, 100, 0)
	// Spherical projection uses len(polyline) as the index at the final vertex.
	end.ProjectionPointIdx = len(line)
	stationary := (SubMatch{Observations: []ObservationResult{end, end, end}}).ResponseGeometries()
	for _, geometry := range stationary {
		if len(geometry.Matched) != 2 || geometry.Matched[0] != line[1] || geometry.Matched[1] != line[1] {
			t.Fatal("stationary endpoint has an invalid retained interval")
		}
	}
}

func TestResponseGeometriesIncomingEdgeAndReturn(t *testing.T) {
	line := responseTestLine([2]float64{0, 0}, [2]float64{100, 0})
	edge := spatial.Edge{ID: 1, Polyline: &line}
	otherLine := responseTestLine([2]float64{-100, 0}, [2]float64{0, 0})
	other := spatial.Edge{ID: 2, Polyline: &otherLine}
	incoming := responseTestObservation(other, -50, 0)
	incoming.NextEdges = []EdgeResult{{ID: 1, Geom: line, Weight: 10}, {ID: 1, Geom: line, Weight: 20}}
	match := SubMatch{Observations: []ObservationResult{incoming, responseTestObservation(edge, 30, 0), responseTestObservation(edge, 80, 0)}}
	got := match.ResponseGeometries()
	if !reflect.DeepEqual(got[0].NextEdges[1].Geom, responseTestLine([2]float64{0, 0}, [2]float64{80, 0})) || got[0].NextEdges[1].Weight != 20 {
		t.Fatal("last incoming edge redrew the terminal tail or changed its weight")
	}
	if !reflect.DeepEqual(got[0].NextEdges[0].Geom, line) || !reflect.DeepEqual(incoming.NextEdges[1].Geom, line) {
		t.Fatal("an earlier traversal or input route was changed")
	}
	departing := responseTestObservation(edge, 30, 0)
	departing.NextEdges = []EdgeResult{{ID: 2, Geom: otherLine}}
	returned := (SubMatch{Observations: []ObservationResult{departing, responseTestObservation(edge, 50, 0), responseTestObservation(edge, 80, 0)}}).ResponseGeometries()
	if returned[1].Matched[0] != line[0] {
		t.Fatal("a return through intermediate edges inherited the first visit's cut")
	}
}

func TestResponseGeometriesDirectedOrder(t *testing.T) {
	forward := responseTestLine([2]float64{0, 0}, [2]float64{100, 0})
	reverse := responseTestLine([2]float64{100, 0}, [2]float64{0, 0})
	forwardEdge := spatial.Edge{ID: 1, Polyline: &forward}
	reverseEdge := spatial.Edge{ID: 2, Polyline: &reverse}
	match := SubMatch{Observations: []ObservationResult{
		responseTestObservation(forwardEdge, 30, 0), responseTestObservation(forwardEdge, 80, 0),
		responseTestObservation(reverseEdge, 80, 0), responseTestObservation(reverseEdge, 30, 0),
	}}
	got := match.ResponseGeometries()
	if !reflect.DeepEqual(got[1].Matched, responseTestLine([2]float64{30, 0}, [2]float64{100, 0})) ||
		!reflect.DeepEqual(got[2].Matched, responseTestLine([2]float64{100, 0}, [2]float64{30, 0})) {
		t.Fatal("opposite directed edges were merged or ordered by coordinates")
	}
	backward := SubMatch{Observations: []ObservationResult{
		responseTestObservation(reverseEdge, 80, 0), responseTestObservation(reverseEdge, 30, 0),
	}}
	for _, geometry := range backward.ResponseGeometries() {
		if !reflect.DeepEqual(geometry.Matched, responseTestLine([2]float64{80, 0}, [2]float64{30, 0})) {
			t.Fatal("retained interval does not follow the directed geometry")
		}
	}
}
