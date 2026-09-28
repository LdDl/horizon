package spatial

import (
	"reflect"
	"sync"
	"testing"

	"github.com/golang/geo/s2"
)

type polylineCutTestCase struct {
	name       string
	line       s2.Polyline
	projection s2.Point
	next       int
	prefix     s2.Polyline
	suffix     s2.Polyline
}

func TestPolylineCutsPreserveGeometryAndOwnership(t *testing.T) {
	a := NewEuclideanS2Point(0, 0)
	b := NewEuclideanS2Point(1, 0)
	c := NewEuclideanS2Point(1, 1)
	d := NewEuclideanS2Point(2, 1)
	first := NewEuclideanS2Point(0.5, 0)
	last := NewEuclideanS2Point(1.5, 1)
	line := s2.Polyline{a, b, c, d}
	for _, test := range []polylineCutTestCase{
		{"first_segment", line, first, 1, s2.Polyline{a, first}, s2.Polyline{first, b, c, d}},
		{"after_corner", line, last, 3, s2.Polyline{a, b, c, last}, s2.Polyline{last, d}},
		{"at_vertex_spherical_index", line, b, 2, s2.Polyline{a, b}, s2.Polyline{b, c, d}},
		{"at_vertex_planar_index", line, b, 1, s2.Polyline{a, b}, s2.Polyline{b, b, c, d}},
		{"at_start", line, a, 1, s2.Polyline{a, a}, s2.Polyline{a, b, c, d}},
		{"at_end_spherical_index", line, d, 4, s2.Polyline{a, b, c, d}, s2.Polyline{d, d}},
		{"at_end_planar_index", line, d, 3, s2.Polyline{a, b, c, d}, s2.Polyline{d, d}},
		{"repeated_vertex", s2.Polyline{a, b, b, c}, b, 2, s2.Polyline{a, b}, s2.Polyline{b, b, c}},
		{"one_vertex", s2.Polyline{a}, a, 1, s2.Polyline{a, a}, s2.Polyline{a, a}},
		{"zero_length", s2.Polyline{a, a}, a, 1, s2.Polyline{a, a}, s2.Polyline{a, a}},
	} {
		for _, from := range []bool{false, true} {
			name := test.name + "/up_to"
			if from {
				name = test.name + "/up_from"
			}
			t.Run(name, func(t *testing.T) {
				// Spare source capacity exposes writes through resliced input buffers.
				source := make(s2.Polyline, len(test.line), len(test.line)+8)
				copy(source, test.line)
				var prefix, suffix s2.Polyline
				if from {
					prefix, suffix = ExtractCutUpFrom(source, test.projection, test.next)
				} else {
					suffix, prefix = ExtractCutUpTo(source, test.projection, test.next)
				}
				if !reflect.DeepEqual(source, test.line) {
					t.Fatalf("cut changed its input: %v", source)
				}
				if !reflect.DeepEqual(prefix, test.prefix) || !reflect.DeepEqual(suffix, test.suffix) {
					t.Fatalf("prefix=%v suffix=%v; want %v and %v", prefix, suffix, test.prefix, test.suffix)
				}
				changed := NewEuclideanS2Point(-10, -10)
				// Appending to either result must not overwrite the other result.
				prefix = append(prefix, changed)
				if !reflect.DeepEqual(suffix, test.suffix) {
					t.Fatal("appending to prefix changed suffix")
				}
				prefix = prefix[:len(prefix)-1]
				suffix = append(suffix, changed)
				if !reflect.DeepEqual(prefix, test.prefix) {
					t.Fatal("appending to suffix changed prefix")
				}
				if !reflect.DeepEqual(source, test.line) {
					t.Fatal("appending to a result changed the source")
				}

				// Check element ownership on fresh results before append reallocates them.
				if from {
					prefix, suffix = ExtractCutUpFrom(source, test.projection, test.next)
				} else {
					suffix, prefix = ExtractCutUpTo(source, test.projection, test.next)
				}
				for i := range prefix {
					prefix[i] = changed
				}
				if !reflect.DeepEqual(suffix, test.suffix) || !reflect.DeepEqual(source, test.line) {
					t.Fatal("prefix shares writable coordinates with suffix or source")
				}
				for i := range suffix {
					suffix[i] = changed
				}
				if !reflect.DeepEqual(source, test.line) {
					t.Fatal("suffix shares writable coordinates with source")
				}
			})
		}
	}
}

func TestPolylineCutsUseProjectionIndices(t *testing.T) {
	line := s2.Polyline{
		s2.PointFromLatLng(s2.LatLngFromDegrees(55.7, 37.6)),
		s2.PointFromLatLng(s2.LatLngFromDegrees(55.7, 37.7)),
		s2.PointFromLatLng(s2.LatLngFromDegrees(55.8, 37.7)),
		s2.PointFromLatLng(s2.LatLngFromDegrees(55.8, 37.8)),
	}
	original := append(s2.Polyline(nil), line...)
	edge := &Edge{Polyline: &line}
	edge.PrecomputeCumLen()
	for _, query := range []s2.Point{line[0], line[1], s2.Interpolate(0.5, line[2], line[3]), line[3]} {
		projection, _, next := CalcProjectionCached(edge, query)
		suffix, prefix := ExtractCutUpTo(line, projection, next)
		if len(prefix) < 2 || len(suffix) < 2 || prefix[0] != line[0] || suffix[len(suffix)-1] != line[len(line)-1] {
			t.Fatal("cut lost an endpoint or returned fewer than two coordinates")
		}
		if prefix[len(prefix)-1] != projection || suffix[0] != projection {
			t.Fatal("cut parts do not meet at the projection")
		}
		if !reflect.DeepEqual(line, original) {
			t.Fatal("cut invalidated the edge geometry")
		}
		if got, _, gotNext := CalcProjectionCached(edge, query); got != projection || gotNext != next {
			t.Fatal("cut changed a repeated projection on the same edge")
		}
	}
}

func TestPolylineCutsConcurrentReaders(t *testing.T) {
	line := s2.Polyline{NewEuclideanS2Point(0, 0), NewEuclideanS2Point(1, 0), NewEuclideanS2Point(1, 1), NewEuclideanS2Point(2, 1)}
	original := append(s2.Polyline(nil), line...)
	projection := NewEuclideanS2Point(1.5, 1)
	wantPrefix := s2.Polyline{line[0], line[1], line[2], projection}
	wantSuffix := s2.Polyline{projection, line[3]}
	var wg sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for repeat := 0; repeat < 50; repeat++ {
				prefix, suffix := ExtractCutUpFrom(line, projection, 3)
				if !reflect.DeepEqual(prefix, wantPrefix) || !reflect.DeepEqual(suffix, wantSuffix) {
					t.Error("concurrent cuts changed the geometry")
					return
				}
				// Each caller may modify its returned coordinates independently.
				prefix[0] = projection
				suffix[0] = line[0]
			}
		}()
	}
	wg.Wait()
	if !reflect.DeepEqual(line, original) {
		t.Fatal("concurrent cuts changed the source")
	}
}
