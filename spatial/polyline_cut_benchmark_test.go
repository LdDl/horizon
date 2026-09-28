package spatial

import (
	"fmt"
	"testing"

	"github.com/golang/geo/s2"
)

var benchmarkCutFirst, benchmarkCutSecond s2.Polyline

type polylineCutBenchmarkCase struct {
	name string
	call func(s2.Polyline, s2.Point, int) (s2.Polyline, s2.Polyline)
}

func BenchmarkPolylineCuts(b *testing.B) {
	for _, cut := range []polylineCutBenchmarkCase{
		{"up_to", ExtractCutUpTo},
		{"up_from", ExtractCutUpFrom},
	} {
		for _, vertices := range []int{4, 64, 1024} {
			b.Run(fmt.Sprintf("%s/vertices_%d", cut.name, vertices), func(b *testing.B) {
				line := make(s2.Polyline, vertices)
				for i := range line {
					line[i] = NewEuclideanS2Point(float64(i), float64(i%7))
				}
				next := len(line) / 2
				previous := line[next-1]
				projected := NewEuclideanS2Point((previous.X+line[next].X)/2, (previous.Y+line[next].Y)/2)
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					benchmarkCutFirst, benchmarkCutSecond = cut.call(line, projected, next)
					// Restore the vertex overwritten by the old implementation when comparing revisions.
					line[next-1] = previous
				}
			})
		}
	}
}
