package horizon

import (
	"testing"
	"time"

	"github.com/LdDl/horizon/spatial"
)

type transitionPresenceBenchmarkCase struct {
	name         string
	candidates   int
	presentEvery int
}

func BenchmarkTransitionRoutePresence(b *testing.B) {
	cases := []transitionPresenceBenchmarkCase{
		{name: "dense5", candidates: 5, presentEvery: 1},
		{name: "sparse5", candidates: 5, presentEvery: 5},
		{name: "empty5", candidates: 5},
		{name: "dense20", candidates: 20, presentEvery: 1},
		{name: "sparse20", candidates: 20, presentEvery: 5},
		{name: "empty20", candidates: 20},
	}
	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			matcher := NewMapMatcherDefault()
			matcher.hmmParams = NewHmmProbabilities(1, 2)
			first := NewGPSMeasurement(0, 0, 0, 0, WithGPSTime(time.Unix(0, 0)))
			second := NewGPSMeasurement(1, 10, 0, 0, WithGPSTime(time.Unix(30, 0)))
			prev := NewCandidateLayer(first, make(RoadPositions, tc.candidates))
			current := NewCandidateLayer(second, make(RoadPositions, tc.candidates))
			for i := 0; i < tc.candidates; i++ {
				edge := &spatial.Edge{ID: int64(i)}
				prev.States[i] = NewRoadPositionFromLonLat(i, 0, 1, edge, 0, 0, 0)
				current.States[i] = NewRoadPositionFromLonLat(tc.candidates+i, 0, 1, edge, 10, 0, 0)
			}
			routes := make(lengths)
			for i, from := range prev.States {
				for j, to := range current.States {
					if tc.presentEvery > 0 && (i*tc.candidates+j)%tc.presentEvery == 0 {
						routes.AddRouteLength(from, to, float64(10+i+j))
					}
				}
			}
			current.TransitionLogProbabilities = make([]transition, 0, tc.candidates*tc.candidates)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				current.TransitionLogProbabilities = current.TransitionLogProbabilities[:0]
				if err := matcher.computeTransitionLogProbabilities(prev, current, routes); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
