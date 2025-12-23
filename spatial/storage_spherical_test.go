package spatial

import (
	"testing"

	"github.com/golang/geo/s2"
)

func TestS2StorageSearchInRadius(t *testing.T) {
	storage := NewS2Storage(17, 35)

	storage.AddEdge(1, &Edge{Polyline: s2.PolylineFromLatLngs(
		[]s2.LatLng{
			s2.LatLngFromDegrees(55.852908303860076, 37.35355854034424),
			s2.LatLngFromDegrees(55.8523241360071, 37.353633642196655),
			s2.LatLngFromDegrees(55.85179416169811, 37.35381603240967),
		})},
	)

	storage.AddEdge(2, &Edge{Polyline: s2.PolylineFromLatLngs(
		[]s2.LatLng{
			s2.LatLngFromDegrees(55.85179867855513, 37.35383212566376),
			s2.LatLngFromDegrees(55.85162553199324, 37.3538938164711),
		})},
	)

	/*
		Point 1 - Line 1 = 6.151 meters
		Point 1 - Line 2 = 28.463 meters

		Point 2 - Line 1 = 728 meters
		Point 2 - Line 2 = 723.737 meters

		Point 3 - Line 1 = 7706.241 meters
		Point 3 - Line 2 = 7687.042 meters
	*/

	found, err := storage.SearchInRadiusLonLat(37.35382676124573, 55.85205463292928, 7)
	if err != nil {
		t.Error(err)
	}
	if len(found) != 1 {
		t.Errorf("Should be 1 element found, but got %d", len(found))
	}

	found2, err := storage.SearchInRadiusLonLat(37.34351634979247, 55.84872257736021, 690)
	if err != nil {
		t.Error(err)
	}
	if len(found2) != 1 {
		t.Errorf("Should be 1 element found, but got %d", len(found2))
	}

	found3, err := storage.SearchInRadiusLonLat(37.40020751953125, 55.787577714316704, 7900)
	if err != nil {
		t.Error(err)
	}
	if len(found3) != 2 {
		t.Errorf("Should be 2 elements found, but got %d", len(found3))
	}
}

func TestS2StorageNewStorage(t *testing.T) {
	// Test that NewStorage creates S2Storage for StorageTypeSpherical
	storage := NewStorage(StorageTypeSpherical)

	_, ok := storage.(*S2Storage)
	if !ok {
		t.Error("NewStorage(StorageTypeSpherical) should return *S2Storage")
	}
}

func TestS2StorageFindNearest(t *testing.T) {
	// Use the same test data as TestS2StorageSearchInRadius
	storage := NewS2Storage(17, 35)

	storage.AddEdge(1, &Edge{Polyline: s2.PolylineFromLatLngs(
		[]s2.LatLng{
			s2.LatLngFromDegrees(55.852908303860076, 37.35355854034424),
			s2.LatLngFromDegrees(55.8523241360071, 37.353633642196655),
			s2.LatLngFromDegrees(55.85179416169811, 37.35381603240967),
		})},
	)

	storage.AddEdge(2, &Edge{Polyline: s2.PolylineFromLatLngs(
		[]s2.LatLng{
			s2.LatLngFromDegrees(55.85179867855513, 37.35383212566376),
			s2.LatLngFromDegrees(55.85162553199324, 37.3538938164711),
		})},
	)

	// Test point (same as in SearchInRadius test)
	// Point 1 - Line 1 = 6.151 meters
	// Point 1 - Line 2 = 28.463 meters
	pt := s2.PointFromLatLng(s2.LatLngFromDegrees(55.85205463292928, 37.35382676124573))

	// Compare FindNearest with FindNearestInRadius (large radius)
	nearestWithRadius, err := storage.FindNearestInRadius(pt, 1000, 2)
	if err != nil {
		t.Errorf("FindNearestInRadius failed: %v", err)
	}

	nearestWithoutRadius, err := storage.FindNearest(pt, 2)
	if err != nil {
		t.Errorf("FindNearest failed: %v", err)
	}

	// Both should return same number of results
	if len(nearestWithRadius) != len(nearestWithoutRadius) {
		t.Errorf("Result count mismatch: FindNearestInRadius=%d, FindNearest=%d",
			len(nearestWithRadius), len(nearestWithoutRadius))
	}

	// Both should return same edges in same order
	for i := range nearestWithRadius {
		if i >= len(nearestWithoutRadius) {
			break
		}
		if nearestWithRadius[i].EdgeID != nearestWithoutRadius[i].EdgeID {
			t.Errorf("Edge mismatch at %d: FindNearestInRadius=%d, FindNearest=%d",
				i, nearestWithRadius[i].EdgeID, nearestWithoutRadius[i].EdgeID)
		}
	}

	// Test with n=0
	nearest0, err := storage.FindNearest(pt, 0)
	if err != nil {
		t.Errorf("FindNearest with n=0 failed: %v", err)
	}
	if nearest0 != nil {
		t.Errorf("Expected empty result for n=0, got %d", len(nearest0))
	}
}
