package horizon

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/LdDl/horizon/spatial"
	"github.com/LdDl/viterbi"
)

type transitionPresenceCase struct {
	name       string
	routes     lengths
	wantCount  int
	wantLength float64
}

func TestTransitionRoutePresence(t *testing.T) {
	cases := []transitionPresenceCase{
		{name: "nil table"},
		{name: "empty table", routes: lengths{}},
		{name: "reverse pair only", routes: lengths{{1, 0}: 0}},
		{name: "other state only", routes: lengths{{2, 1}: 0}},
		{name: "explicit zero", routes: lengths{{0, 1}: 0}, wantCount: 1},
		{name: "positive length", routes: lengths{{0, 1}: 10}, wantCount: 1, wantLength: 10},
		{name: "negative length", routes: lengths{{0, 1}: -1}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			matcher := NewMapMatcherDefault()
			matcher.hmmParams = NewHmmProbabilities(1, 2)
			first := NewGPSMeasurement(0, 0, 0, 0, WithGPSTime(time.Unix(0, 0)))
			second := NewGPSMeasurement(1, 0, 0, 0, WithGPSTime(time.Unix(30, 0)))
			edge := &spatial.Edge{ID: 42}
			from := NewRoadPositionFromLonLat(0, 0, 1, edge, 0, 0, 0)
			to := NewRoadPositionFromLonLat(1, 0, 1, edge, 0, 0, 0)
			prev := NewCandidateLayer(first, RoadPositions{from})
			current := NewCandidateLayer(second, RoadPositions{to})
			if err := matcher.computeTransitionLogProbabilities(prev, current, tc.routes); err != nil {
				t.Fatal(err)
			}
			if got := len(current.TransitionLogProbabilities); got != tc.wantCount {
				t.Fatalf("transition count = %d, want %d", got, tc.wantCount)
			}
			if tc.wantCount == 0 {
				return
			}
			got := current.TransitionLogProbabilities[0]
			want := math.Log(0.5) - tc.wantLength/2
			if got.from != from || got.to != to || math.IsNaN(got.prob) || math.Abs(got.prob-want) > 1e-12 {
				t.Fatalf("transition = %+v, want the requested pair with score %.15f", got, want)
			}
		})
	}
}

type viterbiRoutePresenceCase struct {
	name      string
	routes    lengths
	wantState int
	penalty   float64
	wantError error
}

func TestPrepareViterbiRoutePresence(t *testing.T) {
	cases := []viterbiRoutePresenceCase{
		{name: "missing route cannot beat reachable candidate", routes: lengths{{0, 1}: 2}, wantState: 1, penalty: 3},
		{name: "explicit zero permits stationary candidate", routes: lengths{{0, 1}: 2, {0, 2}: 0}, wantState: 2},
		{name: "no routes means no path", wantError: viterbi.ErrPathBroken},
		{name: "negative routes remain forbidden", routes: lengths{{0, 1}: -1, {0, 2}: -1}, wantError: viterbi.ErrPathBroken},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			matcher := NewMapMatcherDefault()
			matcher.hmmParams = NewHmmProbabilities(1, 2)
			observations := GPSMeasurements{
				NewGPSMeasurement(0, 0, 0, 0, WithGPSTime(time.Unix(0, 0))),
				NewGPSMeasurement(1, 0, 0, 0, WithGPSTime(time.Unix(30, 0))),
			}
			from := NewRoadPositionFromLonLat(0, 0, 1, &spatial.Edge{ID: 10}, 0, 0, 0)
			reachable := NewRoadPositionFromLonLat(1, 1, 2, &spatial.Edge{ID: 11}, 2, 0, 0)
			stationary := NewRoadPositionFromLonLat(2, 0, 1, from.GraphEdge, 0, 0, 0)
			layers := []*CandidateLayer{
				NewCandidateLayer(observations[0], RoadPositions{from}),
				NewCandidateLayer(observations[1], RoadPositions{reachable, stationary}),
			}
			decoder, err := matcher.PrepareViterbi(layers, tc.routes, observations)
			if err != nil {
				t.Fatal(err)
			}
			path, err := decoder.EvalPathLogProbabilities()
			if !errors.Is(err, tc.wantError) {
				t.Fatalf("decode error = %v, want %v; path = %+v", err, tc.wantError, path)
			}
			if tc.wantError != nil {
				return
			}
			if len(path.Path) != 2 || path.Path[0] != from || path.Path[1].ID() != tc.wantState {
				t.Fatalf("path = %v, want states [0 %d]", path.Path, tc.wantState)
			}
			for i := 1; i < len(path.Path); i++ {
				key := [2]int{path.Path[i-1].ID(), path.Path[i].ID()}
				if _, exists := tc.routes[key]; !exists {
					t.Fatalf("decoded path uses missing route %v", key)
				}
			}
			// The existing model counts the first position in both prior and emission.
			want := 3*math.Log(1/math.Sqrt(2*math.Pi)) + math.Log(0.5) - tc.penalty
			if math.IsNaN(path.Probability) || math.Abs(path.Probability-want) > 1e-12 {
				t.Fatalf("score = %.15f, want %.15f", path.Probability, want)
			}
		})
	}
}
