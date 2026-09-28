package spatial

import (
	"math"
	"math/rand"
	"testing"

	"github.com/golang/geo/s2"
)

type sphericalRadiusQuery struct {
	name string
	call func(*S2Storage, s2.Point, float64) (map[uint64]float64, error)
}

func sphericalRadiusQueries() []sphericalRadiusQuery {
	return []sphericalRadiusQuery{
		{name: "point", call: (*S2Storage).SearchInRadius},
		{name: "interface", call: (*S2Storage).FindInRadius},
		{name: "lon_lat", call: func(storage *S2Storage, pt s2.Point, radius float64) (map[uint64]float64, error) {
			ll := s2.LatLngFromPoint(pt)
			return storage.SearchInRadiusLonLat(ll.Lng.Degrees(), ll.Lat.Degrees(), radius)
		}},
	}
}

func sphericalScanDistance(pt s2.Point, line s2.Polyline) float64 {
	best := math.Inf(1)
	for i := 1; i < len(line); i++ {
		distance := s2.DistanceFromSegment(pt, line[i-1], line[i]).Radians() * EarthRadius
		if distance < best {
			best = distance
		}
	}
	return best
}

func addSphericalDistanceEdge(t *testing.T, storage *S2Storage, id uint64, line s2.Polyline, bound bool) {
	t.Helper()
	edge := &Edge{ID: int64(id), Polyline: &line}
	if bound {
		edge.PrecomputeBound()
	}
	if err := storage.AddEdge(id, edge); err != nil {
		t.Fatal(err)
	}
}

func TestSphericalSearchUsesQueryPoint(t *testing.T) {
	cell := s2.CellFromLatLng(s2.LatLngFromDegrees(55.75, 37.6))
	query := cell.Center()
	near := s2.Polyline{s2.Interpolate(0.1, query, cell.Vertex(0)), s2.Interpolate(0.2, query, cell.Vertex(0))}
	far := s2.Polyline{s2.Interpolate(0.6, query, cell.Vertex(0)), s2.Interpolate(0.7, query, cell.Vertex(0))}
	nearDistance, farDistance := sphericalScanDistance(query, near), sphericalScanDistance(query, far)
	if !(0 < nearDistance && nearDistance < farDistance) {
		t.Fatal("invalid distance fixture")
	}
	storage := NewS2Storage(17, 35)
	addSphericalDistanceEdge(t, storage, 1, far, false)
	addSphericalDistanceEdge(t, storage, 9, near, false)
	for _, radiusQuery := range sphericalRadiusQueries() {
		t.Run(radiusQuery.name, func(t *testing.T) {
			got, err := radiusQuery.call(storage, query, 1)
			if err != nil || len(got) != 2 {
				t.Fatalf("got %v, error=%v", got, err)
			}
			for id, want := range map[uint64]float64{1: farDistance, 9: nearDistance} {
				if math.Abs(got[id]-want) > 1e-8 {
					t.Errorf("edge %d: got %.12g m, want %.12g m from the query point", id, got[id], want)
				}
			}
		})
	}
	for _, withRadius := range []bool{false, true} {
		var got []NearestObject
		var err error
		if withRadius {
			got, err = storage.FindNearestInRadius(query, 1, 1)
		} else {
			got, err = storage.FindNearest(query, 1)
		}
		if err != nil || len(got) != 1 || got[0].EdgeID != 9 || math.Abs(got[0].DistanceTo-nearDistance) > 1e-8 {
			t.Errorf("withRadius=%v: got %v, error=%v, want edge 9 at %.12g m", withRadius, got, err, nearDistance)
		}
	}
}

func TestSphericalRadiusRejectsOutsideGeometry(t *testing.T) {
	query := s2.PointFromLatLng(s2.LatLngFromDegrees(55.75, 37.6))
	// A U-shaped edge has a loose cap containing the query, but every segment is outside 1 m.
	line := *s2.PolylineFromLatLngs([]s2.LatLng{
		s2.LatLngFromDegrees(55.74995, 37.59995),
		s2.LatLngFromDegrees(55.75005, 37.59995),
		s2.LatLngFromDegrees(55.75005, 37.60005),
		s2.LatLngFromDegrees(55.74995, 37.60005),
	})
	distance := sphericalScanDistance(query, line)
	if distance <= 1 {
		t.Fatal("invalid outside-radius fixture")
	}
	for _, bound := range []bool{false, true} {
		storage := NewS2Storage(17, 35)
		addSphericalDistanceEdge(t, storage, 7, line, bound)
		for _, radiusQuery := range sphericalRadiusQueries() {
			got, err := radiusQuery.call(storage, query, 1)
			if err != nil || len(got) != 0 {
				t.Errorf("%s bound=%v: got %v, error=%v; edge is %.9f m away, radius=1 m", radiusQuery.name, bound, got, err, distance)
			}
		}
		got, err := storage.FindNearestInRadius(query, 1, 5)
		if err != nil || len(got) != 0 {
			t.Errorf("nearest bound=%v: got %v, error=%v; radius=1 m", bound, got, err)
		}
	}
}

func TestSphericalRadiusIncludesComputedBoundary(t *testing.T) {
	query := s2.PointFromLatLng(s2.LatLngFromDegrees(0, 0))
	line := *s2.PolylineFromLatLngs([]s2.LatLng{s2.LatLngFromDegrees(0.0001, 0), s2.LatLngFromDegrees(0.0002, 0)})
	distance := sphericalScanDistance(query, line)
	storage := NewS2Storage(17, 35)
	addSphericalDistanceEdge(t, storage, 7, line, false)
	for _, radiusQuery := range sphericalRadiusQueries() {
		for _, radius := range []float64{math.Nextafter(distance, 0), distance, math.Nextafter(distance, math.Inf(1))} {
			got, err := radiusQuery.call(storage, query, radius)
			_, included := got[7]
			if err != nil || included != (distance <= radius) {
				t.Errorf("%s radius=%.17g distance=%.17g: got %v, error=%v", radiusQuery.name, radius, distance, got, err)
			}
		}
	}
}

func TestSphericalZeroRadius(t *testing.T) {
	query := s2.PointFromLatLng(s2.LatLngFromDegrees(0, 0))
	east := s2.PointFromLatLng(s2.LatLngFromDegrees(0, 0.0001))
	west := s2.PointFromLatLng(s2.LatLngFromDegrees(0, -0.0001))
	north := s2.PointFromLatLng(s2.LatLngFromDegrees(0.00001, 0))
	storage := NewS2Storage(17, 35)
	addSphericalDistanceEdge(t, storage, 1, s2.Polyline{west, query, east}, true)
	addSphericalDistanceEdge(t, storage, 2, s2.Polyline{query, west}, true)
	addSphericalDistanceEdge(t, storage, 3, s2.Polyline{query, query}, true)
	addSphericalDistanceEdge(t, storage, 4, s2.Polyline{north, east}, false)
	for _, radiusQuery := range sphericalRadiusQueries() {
		got, err := radiusQuery.call(storage, query, 0)
		if err != nil || len(got) != 3 {
			t.Fatalf("%s: got %v, error=%v, want three edges through the query", radiusQuery.name, got, err)
		}
		for id := uint64(1); id <= 3; id++ {
			distance, exists := got[id]
			if !exists || distance != 0 {
				t.Errorf("%s edge=%d distance=%g exists=%v, want zero distance", radiusQuery.name, id, distance, exists)
			}
		}
	}
}

func TestSphericalRadiusMatchesSegmentScan(t *testing.T) {
	centers := []s2.LatLng{
		s2.LatLngFromDegrees(55.75, 37.6),
		s2.LatLngFromDegrees(0, 179.99999),
		s2.LatLngFromDegrees(85, 20),
	}
	rng := rand.New(rand.NewSource(42))
	for _, center := range centers {
		storage := NewS2Storage(17, 35)
		query := s2.PointFromLatLng(center)
		lines := make(map[uint64]s2.Polyline)
		for id := uint64(1); id <= 32; id++ {
			line := make(s2.Polyline, 3)
			for i := range line {
				lat := center.Lat.Degrees() + (rng.Float64()-0.5)*0.001
				lon := center.Lng.Degrees() + (rng.Float64()-0.5)*0.001/math.Cos(center.Lat.Radians())
				line[i] = s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lon))
			}
			lines[id] = line
			addSphericalDistanceEdge(t, storage, id, line, id%2 == 0)
		}
		for _, radius := range []float64{1, 10, 30, 100} {
			for _, radiusQuery := range sphericalRadiusQueries() {
				got, err := radiusQuery.call(storage, query, radius)
				if err != nil {
					t.Fatal(err)
				}
				for id, line := range lines {
					want := sphericalScanDistance(query, line)
					distance, included := got[id]
					if included != (want <= radius) || included && math.Abs(distance-want) > 1e-8 {
						t.Fatalf("%s center=%v radius=%g edge=%d: distance=%g included=%v, want distance=%g included=%v", radiusQuery.name, center, radius, id, distance, included, want, want <= radius)
					}
				}
			}
		}
	}
}
