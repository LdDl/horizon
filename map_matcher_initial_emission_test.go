package horizon

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/LdDl/horizon/spatial"
)

func TestPrepareViterbiCountsEachEmissionOnce(t *testing.T) {
	for _, count := range []int{1, 2, 6} {
		for _, accuracy := range []float64{0, 2} {
			t.Run(fmt.Sprintf("points=%d/accuracy=%g", count, accuracy), func(t *testing.T) {
				matcher := NewMapMatcher(WithHmmParameters(NewHmmProbabilities(1, 2)))
				observations := make(GPSMeasurements, count)
				layers := make([]*CandidateLayer, count)
				routes := make(lengths)
				want := 0.0
				for i := range observations {
					x, distance := float64(i)*10, float64(i+1)
					observations[i] = NewGPSMeasurement(i, x, 0, 0, WithGPSTime(time.Unix(int64(i)*30, 0)), WithGPSAccuracy(accuracy))
					state := NewRoadPositionFromLonLat(i, 0, 1, &spatial.Edge{ID: int64(i)}, x, distance, 0)
					layers[i] = NewCandidateLayer(observations[i], RoadPositions{state})
					sigma := 1.0
					if accuracy > 0 {
						sigma = accuracy
					}
					want += -math.Log(sigma*math.Sqrt(2*math.Pi)) - distance*distance/(2*sigma*sigma)
					if i > 0 {
						routes[[2]int{i - 1, i}] = 20
						want += -math.Log(2) - 5
					}
				}
				decoder, err := matcher.PrepareViterbi(layers, routes, observations)
				if err != nil {
					t.Fatal(err)
				}
				path, err := decoder.EvalPathLogProbabilities()
				if err != nil {
					t.Fatal(err)
				}
				if len(path.Path) != count || math.IsNaN(path.Probability) || math.Abs(path.Probability-want) > 1e-12 {
					t.Fatalf("path length=%d score=%.15f, want length=%d score=%.15f", len(path.Path), path.Probability, count, want)
				}
			})
		}
	}
}

func TestInitialEmissionDoesNotReversePathChoice(t *testing.T) {
	// Equal transition scores leave squared position errors A=(0,9,0), B=(4,4,0).
	// A repeated first emission chooses A with cost 9 instead of B with cost 8.
	for permutation := 0; permutation < 8; permutation++ {
		t.Run(fmt.Sprintf("permutation=%d", permutation), func(t *testing.T) {
			matcher := NewMapMatcher(WithHmmParameters(NewHmmProbabilities(1, 2)))
			observations := make(GPSMeasurements, 3)
			layers := make([]*CandidateLayer, 3)
			routes := lengths{{0, 2}: 20, {2, 4}: 20, {1, 3}: 20, {3, 5}: 20}
			errorsA, errorsB := []float64{0, 3, 0}, []float64{2, 2, 0}
			for i := range observations {
				x := float64(i) * 10
				observations[i] = NewGPSMeasurement(i, x, 0, 0, WithGPSTime(time.Unix(int64(i)*30, 0)))
				a := NewRoadPositionFromLonLat(2*i, 0, 1, &spatial.Edge{ID: 10}, x, errorsA[i], 0)
				b := NewRoadPositionFromLonLat(2*i+1, 2, 3, &spatial.Edge{ID: 20}, x, errorsB[i], 0)
				states := RoadPositions{a, b}
				if permutation&(1<<i) != 0 {
					states[0], states[1] = states[1], states[0]
				}
				layers[i] = NewCandidateLayer(observations[i], states)
			}
			decoder, err := matcher.PrepareViterbi(layers, routes, observations)
			if err != nil {
				t.Fatal(err)
			}
			path, err := decoder.EvalPathLogProbabilities()
			if err != nil {
				t.Fatal(err)
			}
			if len(path.Path) != 3 {
				t.Fatalf("path length=%d, want 3", len(path.Path))
			}
			for i, state := range path.Path {
				if state.ID() != 2*i+1 {
					t.Fatalf("state[%d]=%d, want %d from path B (cost 8 < 9)", i, state.ID(), 2*i+1)
				}
			}
			want := -3*math.Log(math.Sqrt(2*math.Pi)) - 4 - 2*math.Log(2) - 10
			if math.IsNaN(path.Probability) || math.Abs(path.Probability-want) > 1e-12 {
				t.Fatalf("score=%.15f, want %.15f", path.Probability, want)
			}
		})
	}
}

func TestRunCountsInitialEmissionOncePerSegment(t *testing.T) {
	cases := []sameEdgeCase{
		{name: "continuous", observations: [][2]float64{{20, 2}, {40, 2}, {60, 2}}, groups: [][]int{{0, 1, 2}}},
		{name: "singletons", observations: [][2]float64{{80, 2}, {50, 2}, {20, 2}}, groups: [][]int{{0}, {1}, {2}}},
		{name: "singleton and pair", observations: [][2]float64{{80, 2}, {20, 2}, {40, 2}}, groups: [][]int{{0}, {1, 2}}},
		{name: "two pairs", observations: [][2]float64{{20, 2}, {70, 2}, {40, 2}, {50, 2}}, groups: [][]int{{0, 1}, {2, 3}}},
	}
	for _, tc := range cases {
		for _, accuracy := range []float64{0, 2} {
			t.Run(fmt.Sprintf("%s/accuracy=%g", tc.name, accuracy), func(t *testing.T) {
				tc.edges = []sameEdgeDefinition{{source: 0, target: 1, coordinates: [][2]float64{{0, 0}, {100, 0}}}}
				matcher, observations := sameEdgeFixture(t, tc, 0, 1)
				for _, observation := range observations {
					WithGPSAccuracy(accuracy)(observation)
				}
				result, err := matcher.Run(observations, 10, 1)
				if err != nil {
					t.Fatal(err)
				}
				if len(result.SubMatches) != len(tc.groups) {
					t.Fatalf("submatches=%d, want %d", len(result.SubMatches), len(tc.groups))
				}
				sigma := 1.0
				if accuracy > 0 {
					sigma = accuracy
				}
				emission := -math.Log(sigma*math.Sqrt(2*math.Pi)) - 4/(2*sigma*sigma)
				for i, match := range result.SubMatches {
					count := len(tc.groups[i])
					want := float64(count)*emission - float64(count-1)*math.Log(2)
					if len(match.Observations) != count || math.IsNaN(match.Probability) || math.Abs(match.Probability-want) > 1e-12 {
						t.Fatalf("submatch %d length=%d score=%.15f, want length=%d score=%.15f", i, len(match.Observations), match.Probability, count, want)
					}
					for j, observation := range match.Observations {
						if !observation.IsMatched || observation.Observation != observations[tc.groups[i][j]] {
							t.Fatalf("unexpected observation in submatch %d position %d", i, j)
						}
					}
				}
			})
		}
	}
}
