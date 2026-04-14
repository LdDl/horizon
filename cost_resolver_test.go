package horizon

import (
	"math"
	"testing"
)

func TestCostResolverWeight(t *testing.T) {
	lookup := NewCSVLookup([]string{"from_vertex_id", "to_vertex_id", "weight", "geom", "edge_id"})
	r, err := NewCostResolver(lookup)
	if err != nil {
		t.Fatal(err)
	}
	if r.Strategy() != CostStrategyWeight {
		t.Fatalf("expected CostStrategyWeight, got %v", r.Strategy())
	}

	record := []string{"1", "2", "150.5", "LINESTRING(0 0,1 1)", "10"}
	cost, err := r.Resolve(record)
	if err != nil {
		t.Fatal(err)
	}
	if cost != 150.5 {
		t.Fatalf("expected 150.5, got %f", cost)
	}
}

func TestCostResolverTime(t *testing.T) {
	lookup := NewCSVLookup([]string{"from_vertex_id", "to_vertex_id", "length_meters", "free_speed", "geom", "edge_id"})
	r, err := NewCostResolver(lookup)
	if err != nil {
		t.Fatal(err)
	}
	if r.Strategy() != CostStrategyTime {
		t.Fatalf("expected CostStrategyTime, got %v", r.Strategy())
	}

	// 1000m at 36 km/h = 1000 / (36/3.6) = 1000 / 10 = 100 seconds
	record := []string{"1", "2", "1000", "36", "LINESTRING(0 0,1 1)", "10"}
	cost, err := r.Resolve(record)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(cost-100.0) > 0.001 {
		t.Fatalf("expected 100.0 seconds, got %f", cost)
	}
}

func TestCostResolverDistance(t *testing.T) {
	lookup := NewCSVLookup([]string{"from_vertex_id", "to_vertex_id", "length_meters", "geom", "edge_id"})
	r, err := NewCostResolver(lookup)
	if err != nil {
		t.Fatal(err)
	}
	if r.Strategy() != CostStrategyDistance {
		t.Fatalf("expected CostStrategyDistance, got %v", r.Strategy())
	}

	record := []string{"1", "2", "250.7", "LINESTRING(0 0,1 1)", "10"}
	cost, err := r.Resolve(record)
	if err != nil {
		t.Fatal(err)
	}
	if cost != 250.7 {
		t.Fatalf("expected 250.7, got %f", cost)
	}
}

func TestCostResolverWeightPriority(t *testing.T) {
	// When weight AND length_meters+free_speed all present, weight wins
	lookup := NewCSVLookup([]string{"from_vertex_id", "to_vertex_id", "weight", "length_meters", "free_speed", "geom", "edge_id"})
	r, err := NewCostResolver(lookup)
	if err != nil {
		t.Fatal(err)
	}
	if r.Strategy() != CostStrategyWeight {
		t.Fatalf("expected CostStrategyWeight (priority), got %v", r.Strategy())
	}
}

func TestCostResolverNoColumns(t *testing.T) {
	lookup := NewCSVLookup([]string{"from_vertex_id", "to_vertex_id", "geom", "edge_id"})
	_, err := NewCostResolver(lookup)
	if err == nil {
		t.Fatal("expected error when no cost columns are available")
	}
}

func TestCostResolverZeroSpeed(t *testing.T) {
	lookup := NewCSVLookup([]string{"length_meters", "free_speed"})
	r, err := NewCostResolver(lookup)
	if err != nil {
		t.Fatal(err)
	}

	record := []string{"1000", "0"}
	_, err = r.Resolve(record)
	if err == nil {
		t.Fatal("expected error for zero speed")
	}
}
