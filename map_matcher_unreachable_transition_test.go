package horizon

import (
	"errors"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/LdDl/ch"
	"github.com/LdDl/horizon/spatial"
	"github.com/LdDl/viterbi"
)

type resolvedRouteCase struct {
	name       string
	cost       float64
	path       []int64
	target     int64
	wantLength float64
	wantPath   []int64
}

func TestResolveRoutePreservesUnreachability(t *testing.T) {
	edges := NewEdgeGraph()
	edges.Set(0, 1, &spatial.Edge{LengthMeters: 2})
	edges.Set(1, 2, &spatial.Edge{LengthMeters: 3})
	edges.Set(3, 4, &spatial.Edge{LengthMeters: 0})
	matcher := &MapMatcher{engine: &MapEngine{edges: edges}}
	cases := []resolvedRouteCase{
		{name: "unreachable", cost: -1, target: 2, wantLength: -1},
		{name: "negative cost ignores stale path", cost: -42, path: []int64{0, 1}, target: 2, wantLength: -1},
		{name: "positive route", cost: 10, path: []int64{0, 1}, target: 2, wantLength: 5, wantPath: []int64{0, 1, 2}},
		{name: "large positive cost is not absence", cost: math.MaxFloat64, path: []int64{0, 1}, target: 2, wantLength: 5, wantPath: []int64{0, 1, 2}},
		{name: "zero connector cost", cost: 0, path: []int64{1}, target: 2, wantLength: 3, wantPath: []int64{1, 2}},
		{name: "zero route length", cost: 0, path: []int64{3}, target: 4, wantLength: 0, wantPath: []int64{3, 4}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Spare capacity exposes accidental writes to the cached route's backing array.
			backing := []int64{71, 72, 73, 74}
			copy(backing, tc.path)
			before := append([]int64(nil), backing...)
			raw := backing[:len(tc.path)]
			length, path := matcher.resolveRoute(tc.cost, raw, tc.target)
			if length != tc.wantLength || !reflect.DeepEqual(path, tc.wantPath) {
				t.Fatalf("resolved route = (%g, %v), want (%g, %v)", length, path, tc.wantLength, tc.wantPath)
			}
			if len(path) > 0 {
				path[0] = 99
			}
			if !reflect.DeepEqual(backing, before) {
				t.Fatalf("cached route was modified: %v, want %v", backing, before)
			}
		})
	}
}

type unreachableTransitionCase struct {
	name       string
	validRoute bool
	length     float64
	beta       float64
}

func TestUnreachableRouteCannotWinViterbi(t *testing.T) {
	var graph ch.Graph
	for id := int64(0); id < 4; id++ {
		if err := graph.CreateVertex(id); err != nil {
			t.Fatal(err)
		}
	}
	for _, pair := range [][2]int64{{0, 1}, {2, 3}} {
		if err := graph.AddEdge(pair[0], pair[1], 1); err != nil {
			t.Fatal(err)
		}
	}
	graph.PrepareContractionHierarchies()
	pool := graph.NewQueryPool()
	rawCost, rawPath := pool.ShortestPath(1, 2)
	if rawCost >= 0 {
		t.Fatalf("disconnected CH query returned a route: cost=%g path=%v", rawCost, rawPath)
	}
	cases := []unreachableTransitionCase{
		{name: "valid score below the old finite penalty", validRoute: true, length: 2, beta: 1e-10},
		{name: "no valid route", beta: 2},
		{name: "zero length remains valid", validRoute: true, length: 0, beta: 2},
		{name: "positive length remains valid", validRoute: true, length: 2, beta: 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			matcher := NewMapMatcherDefault()
			matcher.hmmParams = NewHmmProbabilities(1, tc.beta)
			edges := NewEdgeGraph()
			edges.Set(0, 1, &spatial.Edge{ID: 10, Source: 0, Target: 1, LengthMeters: tc.length})
			matcher.engine = &MapEngine{edges: edges}
			observations := GPSMeasurements{
				NewGPSMeasurement(0, 0, 0, 0, WithGPSTime(time.Unix(0, 0))),
				NewGPSMeasurement(1, 0, 0, 0, WithGPSTime(time.Unix(30, 0))),
			}
			from := NewRoadPositionFromLonLat(0, 0, 1, &spatial.Edge{ID: 9}, 0, 0, 0)
			valid := NewRoadPositionFromLonLat(1, 0, 1, edges.Get(0, 1), tc.length, 0, 0)
			unreachable := NewRoadPositionFromLonLat(2, 2, 3, &spatial.Edge{ID: 20}, 0, 0, 0)
			layers := []*CandidateLayer{
				NewCandidateLayer(observations[0], RoadPositions{from}),
				NewCandidateLayer(observations[1], RoadPositions{valid, unreachable}),
			}
			routes := make(lengths)
			unreachableLength, _ := matcher.resolveRoute(rawCost, rawPath, 3)
			routes.AddRouteLength(from, unreachable, unreachableLength)
			if tc.validRoute {
				validLength, _ := matcher.resolveRoute(0, []int64{0}, 1)
				routes.AddRouteLength(from, valid, validLength)
			}
			decoder, err := matcher.PrepareViterbi(layers, routes, observations)
			if err != nil {
				t.Fatal(err)
			}
			path, err := decoder.EvalPathLogProbabilities()
			if !tc.validRoute {
				if !errors.Is(err, viterbi.ErrPathBroken) {
					t.Fatalf("decode error = %v, want ErrPathBroken; score = %g", err, path.Probability)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(path.Path) != 2 || path.Path[0] != from || path.Path[1] != valid {
				t.Fatalf("decoder selected an unreachable route; score = %.15g", path.Probability)
			}
			for _, transition := range layers[1].TransitionLogProbabilities {
				if transition.to == unreachable {
					t.Fatal("unreachable route received a transition probability")
				}
			}
			// Preserve the current model's initial prior and both positional emissions.
			want := 3*math.Log(1/math.Sqrt(2*math.Pi)) - tc.length*tc.length/2 + math.Log(1/tc.beta) - tc.length/tc.beta
			if math.IsNaN(path.Probability) || math.Abs(path.Probability-want) > 1e-12*math.Max(1, math.Abs(want)) {
				t.Fatalf("score = %.15g, want %.15g", path.Probability, want)
			}
		})
	}
}
