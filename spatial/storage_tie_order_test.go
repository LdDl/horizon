package spatial

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/geo/s2"
)

type tiedStorageCase struct {
	name   string
	create func() Storage
	line   s2.Polyline
	query  s2.Point
	radius float64
}

func TestStorageNearestTieOrder(t *testing.T) {
	cases := []tiedStorageCase{
		{
			name: "euclidean", create: func() Storage { return NewEuclideanStorage() },
			line:  s2.Polyline{NewEuclideanS2Point(-10, 0), NewEuclideanS2Point(10, 0)},
			query: NewEuclideanS2Point(0, 0), radius: 10,
		},
		{
			name: "spherical", create: func() Storage { return NewS2Storage(17, 35) },
			line:  s2.Polyline{NewWGS84Point(0, 0).Point, NewWGS84Point(0.001, 0).Point},
			query: NewWGS84Point(0.0005, 0).Point, radius: 10,
		},
	}
	want := []NearestObject{{2, 0}, {7, 0}, {9, 0}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			eachNearestPermutation(want, func(order []NearestObject) {
				storage := tc.create()
				for _, object := range order {
					line := append(s2.Polyline(nil), tc.line...)
					// Opposite directions are separate objects even on identical geometry.
					if object.EdgeID == 7 {
						line[0], line[1] = line[1], line[0]
					}
					edge := &Edge{ID: int64(object.EdgeID), Polyline: &line}
					if _, spherical := storage.(*S2Storage); spherical {
						edge.PrecomputeBound()
					}
					if err := storage.AddEdge(object.EdgeID, edge); err != nil {
						t.Fatal(err)
					}
				}
				for _, n := range []int{0, 1, 2, 3} {
					for _, withRadius := range []bool{false, true} {
						for repeat := 0; repeat < 8; repeat++ {
							var got []NearestObject
							var err error
							if withRadius {
								got, err = storage.FindNearestInRadius(tc.query, tc.radius, n)
							} else {
								got, err = storage.FindNearest(tc.query, n)
							}
							if err != nil || len(got) != n || n > 0 && !reflect.DeepEqual(got, want[:n]) {
								t.Fatalf("%s order=%v n=%d radius=%v repeat=%d: got %v error=%v, want %v", tc.name, order, n, withRadius, repeat, got, err, want[:n])
							}
						}
					}
				}
			})
		})
	}
}

func TestEuclideanNearestStrictDistanceBeatsID(t *testing.T) {
	storage := NewEuclideanStorage()
	for _, id := range []uint64{2, 9} {
		y := 2.0
		if id == 9 {
			y = 1
		}
		line := s2.Polyline{NewEuclideanS2Point(-10, y), NewEuclideanS2Point(10, y)}
		if err := storage.AddEdge(id, &Edge{ID: int64(id), Polyline: &line}); err != nil {
			t.Fatal(err)
		}
	}
	query := NewEuclideanS2Point(0, 0)
	for _, withRadius := range []bool{false, true} {
		t.Run(fmt.Sprintf("radius=%v", withRadius), func(t *testing.T) {
			var got []NearestObject
			var err error
			if withRadius {
				got, err = storage.FindNearestInRadius(query, 10, 1)
			} else {
				got, err = storage.FindNearest(query, 1)
			}
			want := []NearestObject{{9, 1}}
			if err != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("got %v error=%v, want %v", got, err, want)
			}
		})
	}
}
