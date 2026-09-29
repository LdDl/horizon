package spatial

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"testing"

	"github.com/golang/geo/s2"
)

func sphericalNearestReference(storage *S2Storage, query s2.Point, n int) []NearestObject {
	if n <= 0 {
		return nil
	}
	var result []NearestObject
	for id, edge := range storage.edges {
		if edge == nil || edge.Polyline == nil || edge.Polyline.NumEdges() < 1 {
			continue
		}
		result = append(result, NearestObject{EdgeID: id, DistanceTo: sphericalScanDistance(query, *edge.Polyline)})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].DistanceTo != result[j].DistanceTo {
			return result[i].DistanceTo < result[j].DistanceTo
		}
		return result[i].EdgeID < result[j].EdgeID
	})
	if n < len(result) {
		result = result[:n]
	}
	return result
}

func checkSphericalNearest(t *testing.T, got, want []NearestObject) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i].EdgeID != want[i].EdgeID || math.Abs(got[i].DistanceTo-want[i].DistanceTo) > 1e-8 {
			t.Fatalf("candidate %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestSphericalNearestMissedDiagonal(t *testing.T) {
	centerID := s2.CellIDFromToken("46b54a4d64")
	center := s2.CellFromCellID(centerID)
	hidden := s2.CellFromCellID(s2.CellIDFromToken("46b54a52a4"))
	vertex := center.Vertex(0)
	for _, inset := range []float64{0.001, 0.05} {
		t.Run(fmt.Sprint(inset), func(t *testing.T) {
			query := s2.Interpolate(inset, vertex, center.Center())
			storage := NewS2Storage(17, 35)
			addSphericalDistanceEdge(t, storage, 1, s2.Polyline{center.Center(), s2.Interpolate(0.01, center.Center(), vertex)}, true)
			addSphericalDistanceEdge(t, storage, 2, s2.Polyline{s2.Interpolate(inset, vertex, hidden.Center()), s2.Interpolate(2*inset, vertex, hidden.Center())}, true)
			want := sphericalNearestReference(storage, query, 1)
			if want[0].EdgeID != 2 || want[0].DistanceTo >= 5 {
				t.Fatal("invalid diagonal fixture")
			}
			got, err := storage.FindNearest(query, 1)
			if err != nil {
				t.Fatal(err)
			}
			checkSphericalNearest(t, got, want)
		})
	}
}

func TestSphericalNearestSparseAndCount(t *testing.T) {
	storage := NewS2Storage(17, 35)
	query := s2.PointFromLatLng(s2.LatLngFromDegrees(0, 0))
	for _, n := range []int{-1, 0, 1, 5} {
		got, err := storage.FindNearest(query, n)
		if err != nil || len(got) != 0 {
			t.Fatalf("empty storage: got %v, error=%v", got, err)
		}
	}
	for id := uint64(1); id <= 3; id++ {
		lat := float64(id)
		addSphericalDistanceEdge(t, storage, id, *s2.PolylineFromLatLngs([]s2.LatLng{
			s2.LatLngFromDegrees(lat, 1), s2.LatLngFromDegrees(lat+0.0001, 1),
		}), false)
	}
	for _, n := range []int{-1, 0, 1, 2, 3, 5, math.MaxInt} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			got, err := storage.FindNearest(query, n)
			if err != nil {
				t.Fatal(err)
			}
			checkSphericalNearest(t, got, sphericalNearestReference(storage, query, n))
		})
	}
}

type sphericalNearestArea struct {
	name string
	lat  float64
	lon  float64
}

func TestSphericalNearestMatchesFullScan(t *testing.T) {
	areas := []sphericalNearestArea{
		{name: "moscow", lat: 55.75, lon: 37.6},
		{name: "dateline", lat: 0, lon: 180},
		{name: "north_pole", lat: 89.9999, lon: 0},
		{name: "south_pole", lat: -89.9999, lon: 0},
		{name: "face_boundary", lat: 0, lon: 45},
	}
	for _, area := range areas {
		t.Run(area.name, func(t *testing.T) {
			random := rand.New(rand.NewSource(42))
			center := s2.CellFromLatLng(s2.LatLngFromDegrees(area.lat, area.lon))
			storage := NewS2Storage(17, 35)
			for id := uint64(1); id <= 40; id++ {
				var line s2.Polyline
				for i := 0; i < 3; i++ {
					line = append(line, s2.PointFromLatLng(s2.LatLngFromDegrees(area.lat+(random.Float64()-0.5)*0.006, area.lon+(random.Float64()-0.5)*0.006)))
				}
				addSphericalDistanceEdge(t, storage, id, line, id%2 == 0)
			}
			cell := s2.CellFromCellID(center.ID().Parent(17))
			queries := []s2.Point{center.Center(), cell.Center(), cell.Vertex(0), cell.Vertex(1), cell.Vertex(2), cell.Vertex(3)}
			for i := 0; i < 10; i++ {
				queries = append(queries, s2.PointFromLatLng(s2.LatLngFromDegrees(area.lat+(random.Float64()-0.5)*0.01, area.lon+(random.Float64()-0.5)*0.01)))
			}
			for i, query := range queries {
				for _, n := range []int{1, 5, 20, 41} {
					t.Run(fmt.Sprintf("q%d/k%d", i, n), func(t *testing.T) {
						got, err := storage.FindNearest(query, n)
						if err != nil {
							t.Fatal(err)
						}
						checkSphericalNearest(t, got, sphericalNearestReference(storage, query, n))
					})
				}
			}
		})
	}
}

func TestSphericalNearestGlobalAndTies(t *testing.T) {
	for _, level := range []int{0, 3, 17, 30} {
		t.Run(fmt.Sprint(level), func(t *testing.T) {
			for _, reverse := range []bool{false, true} {
				storage := NewS2Storage(level, 35)
				for i := 0; i < 6; i++ {
					face := i
					if reverse {
						face = 5 - i
					}
					point := s2.CellFromCellID(s2.CellIDFromFace(face)).Center()
					for _, id := range []uint64{uint64(face)*2 + 2, uint64(face)*2 + 1} {
						addSphericalDistanceEdge(t, storage, id, s2.Polyline{point, point}, id%2 == 0)
					}
				}
				for face := 0; face < 6; face++ {
					cell := s2.CellFromCellID(s2.CellIDFromFace(face))
					for _, query := range []s2.Point{cell.Center(), cell.Vertex(0)} {
						for _, n := range []int{1, 3, 11, 15} {
							got, err := storage.FindNearest(query, n)
							if err != nil {
								t.Fatal(err)
							}
							checkSphericalNearest(t, got, sphericalNearestReference(storage, query, n))
						}
					}
				}
			}
		})
	}
}

func TestSphericalNearestBudgetFallback(t *testing.T) {
	storage := NewS2Storage(17, 35)
	for id := uint64(1); id <= 40; id++ {
		lat := 55.75 + float64(id)*0.00001
		addSphericalDistanceEdge(t, storage, id, *s2.PolylineFromLatLngs([]s2.LatLng{
			s2.LatLngFromDegrees(lat, 37.6), s2.LatLngFromDegrees(lat+0.0005, 37.605),
		}), true)
	}
	for _, ll := range []s2.LatLng{s2.LatLngFromDegrees(55.75, 37.6), s2.LatLngFromDegrees(0, 0)} {
		query := s2.PointFromLatLng(ll)
		want := sphericalNearestReference(storage, query, 5)
		for _, budget := range []int{0, 1, 2, maxNearestCellVisits, math.MaxInt} {
			t.Run(fmt.Sprintf("%v/budget%d", ll, budget), func(t *testing.T) {
				// Even an interrupted first covering must return the same complete result.
				checkSphericalNearest(t, storage.findNearest(query, 5, budget), want)
			})
		}
	}
}
