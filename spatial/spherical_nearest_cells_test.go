package spatial

import (
	"math"
	"math/rand"
	"testing"

	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
)

func TestNearestCellQueue(t *testing.T) {
	random := rand.New(rand.NewSource(20260928))
	var buffer [64]nearestCell
	queue := nearestCellQueue(buffer[:0])
	const count = 1024
	for i := 1; i <= count; i++ {
		queue.push(nearestCell{id: s2.CellID(i), distance: s1.ChordAngle(random.Intn(25))})
	}
	seen := make(map[s2.CellID]bool)
	previous := s1.ChordAngle(-1)
	for len(queue) > 0 {
		cell := queue.pop()
		if cell.distance < previous || seen[cell.id] {
			t.Fatalf("invalid queue result: %+v, previous distance=%v", cell, previous)
		}
		previous = cell.distance
		seen[cell.id] = true
	}
	if len(seen) != count {
		t.Fatalf("returned %d cells, want %d", len(seen), count)
	}
}

func TestSphericalNearestChangedOccupancy(t *testing.T) {
	storage := NewS2Storage(17, 35)
	query := s2.PointFromLatLng(s2.LatLngFromDegrees(55.75, 37.6))
	for id := uint64(1); id <= 24; id++ {
		// Insert disjoint ranges, then closer edges and exact ties after earlier queries.
		point := s2.PointFromLatLng(s2.LatLngFromDegrees(55.75+float64(24-id)*0.001, 37.6))
		if id >= 20 {
			point = query
		}
		addSphericalDistanceEdge(t, storage, id, s2.Polyline{point, point}, true)
		for _, k := range []int{1, 5, 32} {
			want := sphericalNearestReference(storage, query, k)
			for _, budget := range []int{0, 1, 2, maxNearestCellVisits, math.MaxInt} {
				checkSphericalNearest(t, storage.findNearest(query, k, budget), want)
			}
		}
	}
}
