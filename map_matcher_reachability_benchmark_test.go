package horizon

import (
	"testing"
	"time"
)

type reachabilityBenchmarkCase struct {
	name   string
	points [][2]float64
}

var benchmarkReachabilityResult MatcherResult

func BenchmarkMapMatcherReachability(b *testing.B) {
	cases := []reachabilityBenchmarkCase{
		{name: "same SCC", points: [][2]float64{{2, 0}, {8, 0}, {8, -1.2}}},
		{name: "forward across SCCs", points: [][2]float64{{2, 0}, {22, 0}, {28, 0}}},
		{name: "unreachable reverse", points: [][2]float64{{28, -1.2}, {2, -1.2}, {1, -0.6}}},
		{name: "separate weak components", points: [][2]float64{{2, 0}, {2, 20}, {8, 20}}},
	}
	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			engine := newReachabilityEngine(b)
			matcher := NewMapMatcher(WithMapEngine(engine), WithHmmParameters(NewHmmProbabilities(1, 2)))
			observations := make(GPSMeasurements, 0, len(tc.points))
			for i, point := range tc.points {
				observations = append(observations, NewGPSMeasurement(i, point[0], point[1], 0, WithGPSTime(time.Unix(int64(i)*30, 0))))
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				result, err := matcher.Run(observations, 0.1, 1)
				if err != nil {
					b.Fatal(err)
				}
				benchmarkReachabilityResult = result
			}
		})
	}
}
