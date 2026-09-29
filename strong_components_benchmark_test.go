package horizon

import (
	"testing"
)

// BenchmarkSCC benchmarks computeStrongConnectedComponents on osm2ch_export.csv graph.
func BenchmarkSCC(b *testing.B) {
	graphFileName := "./test_data/osm2ch_export.csv"

	b.Log("Loading graph...")
	hmmParams := NewHmmProbabilities(50.0, 30.0)
	matcher, err := NewMapMatcherFromFiles(hmmParams, graphFileName)
	if err != nil {
		b.Fatalf("Failed to load graph: %v", err)
	}

	b.Logf("Graph has %d source vertices in edges map", matcher.engine.edges.Len())

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result := matcher.engine.computeStrongConnectedComponents()
		if result.TotalComponents == 0 {
			b.Fatal("Expected at least 1 component")
		}
	}
}
