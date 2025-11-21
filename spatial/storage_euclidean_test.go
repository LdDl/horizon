package spatial

import (
	"testing"

	"github.com/golang/geo/s2"
)

func TestEuclideanStorageBasic(t *testing.T) {
	storage := NewEuclideanStorage()

	// Edge 1: from (0, 0) to (1, 0) - horizontal line at y=0
	poly1 := make(s2.Polyline, 2)
	poly1[0] = NewEuclideanS2Point(0, 0)
	poly1[1] = NewEuclideanS2Point(1, 0)
	edge1 := &Edge{
		ID:       1,
		Source:   0,
		Target:   1,
		Polyline: &poly1,
	}

	// Edge 2: from (0, 1) to (1, 1) - horizontal line at y=1
	poly2 := make(s2.Polyline, 2)
	poly2[0] = NewEuclideanS2Point(0, 1)
	poly2[1] = NewEuclideanS2Point(1, 1)
	edge2 := &Edge{
		ID:       2,
		Source:   1,
		Target:   2,
		Polyline: &poly2,
	}

	// Add edges
	storage.AddEdge(1, edge1)
	storage.AddEdge(2, edge2)

	// Test GetEdge
	if storage.GetEdge(1) != edge1 {
		t.Error("GetEdge(1) should return edge1")
	}
	if storage.GetEdge(2) != edge2 {
		t.Error("GetEdge(2) should return edge2")
	}

	// Point at (0.5, 0.2) with radius 0.3 should find only edge1 (distance 0.2)
	pt := NewEuclideanS2Point(0.5, 0.2)
	found, err := storage.FindInRadius(pt, 0.3)
	if err != nil {
		t.Error(err)
	}
	if len(found) != 1 {
		t.Errorf("Expected 1 edge, got %d", len(found))
	}
	if _, ok := found[1]; !ok {
		t.Error("Expected to find edge 1")
	}

	// Point at (0.5, 0.5) with radius 1 should find both edges
	// Distance to edge1 (y=0): 0.5
	// Distance to edge2 (y=1): 0.5
	pt2 := NewEuclideanS2Point(0.5, 0.5)
	found2, err := storage.FindInRadius(pt2, 1)
	if err != nil {
		t.Error(err)
	}
	if len(found2) != 2 {
		t.Errorf("Expected 2 edges, got %d", len(found2))
	}

	// Point at (0.5, 0.3) - closer to edge1 (distance 0.3) than edge2 (distance 0.7)
	pt3 := NewEuclideanS2Point(0.5, 0.3)
	nearest, err := storage.FindNearestInRadius(pt3, 1, 1)
	if err != nil {
		t.Error(err)
	}
	if len(nearest) != 1 {
		t.Errorf("Expected 1 nearest, got %d", len(nearest))
	}
	// Edge 1 should be closer
	if nearest[0].EdgeID != 1 {
		t.Errorf("Expected edge 1 to be nearest, got edge %d (distance: %f)", nearest[0].EdgeID, nearest[0].DistanceTo)
	}
}

func TestEuclideanStorageNewStorage(t *testing.T) {
	// Test that NewStorage creates EuclideanStorage for StorageTypeEuclidean
	storage := NewStorage(StorageTypeEuclidean)

	_, ok := storage.(*EuclideanStorage)
	if !ok {
		t.Error("NewStorage(StorageTypeEuclidean) should return *EuclideanStorage")
	}
}
