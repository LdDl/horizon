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
