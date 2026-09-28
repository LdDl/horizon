package horizon

import (
	"testing"

	"github.com/LdDl/horizon/spatial"
)

type resolvedRouteBenchmarkCase struct {
	name string
	cost float64
	path []int64
}

var benchmarkResolvedLength float64
var benchmarkResolvedPath []int64

func BenchmarkResolveRoute(b *testing.B) {
	edges := NewEdgeGraph()
	edges.Set(0, 1, &spatial.Edge{LengthMeters: 2})
	edges.Set(1, 2, &spatial.Edge{LengthMeters: 3})
	matcher := &MapMatcher{engine: &MapEngine{edges: edges}}
	for _, tc := range []resolvedRouteBenchmarkCase{
		{name: "unreachable", cost: -1},
		{name: "reachable", cost: 10, path: []int64{0, 1}},
	} {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				benchmarkResolvedLength, benchmarkResolvedPath = matcher.resolveRoute(tc.cost, tc.path, 2)
			}
		})
	}
}
