package horizon

import (
	"fmt"
	"math"
	"testing"
	"time"
)

// Benchmark test points for osm2ch_export.csv graph
// Source: near vertex 19234
// Target: near vertex 14886
// Coordinates slightly offset to simulate GPS noise
const (
	// Source (near vertex 19234)
	benchStartLon = 37.55009827
	benchStartLat = 55.72766118
	// Target (near vertex 14886)
	benchEndLon = 37.68867676
	benchEndLat = 55.77719452
)

// BenchmarkRouting benchmarks FindShortestPath on osm2ch_export.csv graph.
// Run with: go test -bench=BenchmarkRouting -benchmem
func BenchmarkRouting(b *testing.B) {
	graphFileName := "./test_data/osm2ch_export.csv"

	b.Log("Loading graph...")
	hmmParams := NewHmmProbabilities(50.0, 30.0)
	matcher, err := NewMapMatcherFromFiles(hmmParams, graphFileName)
	if err != nil {
		b.Fatalf("Failed to load graph: %v", err)
	}

	source := NewGPSMeasurement(1, benchStartLon, benchStartLat, 4326, WithGPSTime(time.Now()))
	target := NewGPSMeasurement(2, benchEndLon, benchEndLat, 4326, WithGPSTime(time.Now()))

	// Verify route works before benchmarking
	result, err := matcher.FindShortestPath(source, target, -1)
	if err != nil {
		b.Fatalf("Failed to find route: %v", err)
	}
	b.Logf("Route found with %d sub-matches", len(result.SubMatches))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := matcher.FindShortestPath(source, target, -1)
		if err != nil {
			b.Error(err)
		}
	}
}

// BenchmarkRouting_WithRadius benchmarks with a bounded search radius.
// Uses 2000m radius which uses FindNearestInRadius.
func BenchmarkRouting_WithRadius(b *testing.B) {
	graphFileName := "./test_data/osm2ch_export.csv"

	b.Log("Loading graph...")
	hmmParams := NewHmmProbabilities(50.0, 30.0)
	matcher, err := NewMapMatcherFromFiles(hmmParams, graphFileName)
	if err != nil {
		b.Fatalf("Failed to load graph: %v", err)
	}

	source := NewGPSMeasurement(1, benchStartLon, benchStartLat, 4326, WithGPSTime(time.Now()))
	target := NewGPSMeasurement(2, benchEndLon, benchEndLat, 4326, WithGPSTime(time.Now()))

	searchRadius := 2000.0

	// Verify route works before benchmarking
	result, err := matcher.FindShortestPath(source, target, searchRadius)
	if err != nil {
		b.Fatalf("Failed to find route: %v", err)
	}
	b.Logf("Route found with %d sub-matches (radius=%.0fm)", len(result.SubMatches), searchRadius)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := matcher.FindShortestPath(source, target, searchRadius)
		if err != nil {
			b.Error(err)
		}
	}
}

// BenchmarkRouting_Iterations runs multiple iterations to get stable measurements.
func BenchmarkRouting_Iterations(b *testing.B) {
	graphFileName := "./test_data/osm2ch_export.csv"

	b.Log("Loading graph...")
	hmmParams := NewHmmProbabilities(50.0, 30.0)
	matcher, err := NewMapMatcherFromFiles(hmmParams, graphFileName)
	if err != nil {
		b.Fatalf("Failed to load graph: %v", err)
	}

	source := NewGPSMeasurement(1, benchStartLon, benchStartLat, 4326, WithGPSTime(time.Now()))
	target := NewGPSMeasurement(2, benchEndLon, benchEndLat, 4326, WithGPSTime(time.Now()))

	b.ResetTimer()

	for k := 0.; k <= 10; k++ {
		n := int(math.Pow(2, k))
		b.Run(fmt.Sprintf("iterations-%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := matcher.FindShortestPath(source, target, -1)
				if err != nil {
					b.Error(err)
				}
			}
		})
	}
}

// TestRouting_Profile is a regular test that can be used with -cpuprofile.
func TestRouting_Profile(t *testing.T) {
	graphFileName := "./test_data/osm2ch_export.csv"

	t.Log("Loading graph...")
	hmmParams := NewHmmProbabilities(50.0, 30.0)
	matcher, err := NewMapMatcherFromFiles(hmmParams, graphFileName)
	if err != nil {
		t.Fatalf("Failed to load graph: %v", err)
	}

	source := NewGPSMeasurement(1, benchStartLon, benchStartLat, 4326, WithGPSTime(time.Now()))
	target := NewGPSMeasurement(2, benchEndLon, benchEndLat, 4326, WithGPSTime(time.Now()))

	iterations := 100
	t.Logf("Running %d iterations for profiling...", iterations)

	for i := 0; i < iterations; i++ {
		result, err := matcher.FindShortestPath(source, target, -1)
		if err != nil {
			t.Fatalf("Iteration %d failed: %v", i, err)
		}
		_ = result
	}

	t.Log("Profiling complete")
}

// TestRouting_SingleRun runs a single routing request with detailed timing.
func TestRouting_SingleRun(t *testing.T) {
	graphFileName := "./test_data/osm2ch_export.csv"

	t.Log("Loading graph...")
	loadStart := time.Now()
	hmmParams := NewHmmProbabilities(50.0, 30.0)
	matcher, err := NewMapMatcherFromFiles(hmmParams, graphFileName)
	if err != nil {
		t.Fatalf("Failed to load graph: %v", err)
	}
	t.Logf("Graph loaded in %v", time.Since(loadStart))
	t.Logf("Vertices: %d, Edges: %d", len(matcher.engine.vertexStrongComponent), len(matcher.engine.edges))

	source := NewGPSMeasurement(1, benchStartLon, benchStartLat, 4326, WithGPSTime(time.Now()))
	target := NewGPSMeasurement(2, benchEndLon, benchEndLat, 4326, WithGPSTime(time.Now()))

	t.Logf("Source: (%f, %f)", benchStartLat, benchStartLon)
	t.Logf("Target: (%f, %f)", benchEndLat, benchEndLon)

	routeStart := time.Now()
	result, err := matcher.FindShortestPath(source, target, -1)
	routeTime := time.Since(routeStart)

	if err != nil {
		t.Fatalf("Failed to find route: %v", err)
	}

	t.Logf("Route found in %v", routeTime)
	t.Logf("Sub-matches: %d", len(result.SubMatches))
	if len(result.SubMatches) > 0 {
		t.Logf("Observations in first sub-match: %d", len(result.SubMatches[0].Observations))
	}
}

// TestRouting_CompareRadiuses compares performance with different search radiuses.
func TestRouting_CompareRadiuses(t *testing.T) {
	graphFileName := "./test_data/osm2ch_export.csv"

	t.Log("Loading graph...")
	hmmParams := NewHmmProbabilities(50.0, 30.0)
	matcher, err := NewMapMatcherFromFiles(hmmParams, graphFileName)
	if err != nil {
		t.Fatalf("Failed to load graph: %v", err)
	}

	source := NewGPSMeasurement(1, benchStartLon, benchStartLat, 4326, WithGPSTime(time.Now()))
	target := NewGPSMeasurement(2, benchEndLon, benchEndLat, 4326, WithGPSTime(time.Now()))

	rads := []float64{-1, 100, 500, 1000, 2000}
	iterations := 10

	for _, radius := range rads {
		var totalTime time.Duration
		var success bool

		for i := 0; i < iterations; i++ {
			start := time.Now()
			_, err := matcher.FindShortestPath(source, target, radius)
			elapsed := time.Since(start)

			if err == nil {
				success = true
				totalTime += elapsed
			}
		}

		if success {
			avgTime := totalTime / time.Duration(iterations)
			t.Logf("Radius %6.0fm: avg %v", radius, avgTime)
		} else {
			t.Logf("Radius %6.0fm: FAILED", radius)
		}
	}
}
