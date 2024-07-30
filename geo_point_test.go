package horizon

import (
	"math"
	"testing"
)

func TestGeoPointDistance(t *testing.T) {
	// SRID = 0
	euPt1_0 := NewEuclideanPoint(9, 8)
	euPt2_0 := NewEuclideanPoint(6, 12)

	d0 := euPt1_0.DistanceTo(euPt2_0)
	actualDistance0 := 5.0
	if d0 != actualDistance0 {
		t.Errorf("SRID = 0, Has to be %f, but got %f", actualDistance0, d0)
	}

	// SRID = 4326
	euPt1_4326 := NewWGS84Point(37.35382676124573, 55.85205463292928)
	euPt2_4326 := NewWGS84Point(37.34351634979247, 55.84872257736021)

	d4326 := euPt1_4326.DistanceTo(euPt2_4326)
	actualDistance4326 := 742.605185
	eps := 10e-6
	if math.Abs(d4326-actualDistance4326) > eps {
		t.Errorf("SRID = 4326, Has to be %f, but got %0.10f", actualDistance4326, d4326)
	}
}
