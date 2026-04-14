package horizon

import (
	"testing"
)

func TestCSVLookupBasic(t *testing.T) {
	header := []string{"id", "name", "weight", "count"}
	lookup := NewCSVLookup(header)

	record := []string{"42", "test_edge", "3.14", "7"}

	// Has
	if !lookup.Has("id") {
		t.Fatal("expected 'id' column to exist")
	}
	if lookup.Has("missing") {
		t.Fatal("expected 'missing' column to not exist")
	}

	// String
	s, err := lookup.String(record, "name")
	if err != nil {
		t.Fatal(err)
	}
	if s != "test_edge" {
		t.Fatalf("expected 'test_edge', got %q", s)
	}

	// Int64
	v, err := lookup.Int64(record, "id")
	if err != nil {
		t.Fatal(err)
	}
	if v != 42 {
		t.Fatalf("expected 42, got %d", v)
	}

	// Float64
	f, err := lookup.Float64(record, "weight")
	if err != nil {
		t.Fatal(err)
	}
	if f != 3.14 {
		t.Fatalf("expected 3.14, got %f", f)
	}

	// Int
	i, err := lookup.Int(record, "count")
	if err != nil {
		t.Fatal(err)
	}
	if i != 7 {
		t.Fatalf("expected 7, got %d", i)
	}
}

func TestCSVLookupMissingColumn(t *testing.T) {
	header := []string{"id", "name"}
	lookup := NewCSVLookup(header)
	record := []string{"1", "foo"}

	_, err := lookup.Int64(record, "weight")
	if err == nil {
		t.Fatal("expected error for missing column")
	}
}

func TestCSVLookupMustHave(t *testing.T) {
	header := []string{"id", "name", "geom"}
	lookup := NewCSVLookup(header)

	if err := lookup.MustHave("id", "name"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err := lookup.MustHave("id", "weight", "cost")
	if err == nil {
		t.Fatal("expected error for missing required columns")
	}
}

func TestCSVLookupColumnOrder(t *testing.T) {
	// Columns in different order than expected - lookup should still work
	header := []string{"geom", "edge_id", "weight", "to_vertex_id", "from_vertex_id"}
	lookup := NewCSVLookup(header)

	record := []string{"LINESTRING(0 0,1 1)", "99", "150.5", "20", "10"}

	from, err := lookup.Int64(record, "from_vertex_id")
	if err != nil {
		t.Fatal(err)
	}
	to, err := lookup.Int64(record, "to_vertex_id")
	if err != nil {
		t.Fatal(err)
	}
	w, err := lookup.Float64(record, "weight")
	if err != nil {
		t.Fatal(err)
	}

	if from != 10 || to != 20 || w != 150.5 {
		t.Fatalf("unexpected values: from=%d to=%d weight=%f", from, to, w)
	}
}
