package horizon

import (
	"fmt"
	"math"
	"testing"
)

func TestMapMatcherSRID_0(t *testing.T) {
	matcher := NewMapMatcherDefault()

	gpsMeasurements := GPSMeasurements{
		NewGPSMeasurementFromID(1, 10, 10, 0),
		NewGPSMeasurementFromID(2, 30, 20, 0),
		NewGPSMeasurementFromID(3, 30, 40, 0),
		NewGPSMeasurementFromID(4, 10, 70, 0),
	}

	rp11 := NewRoadPositionFromLonLat(0, 0, &Edge{Source: 0, Target: 1}, 20, 10, 0)
	rp12 := NewRoadPositionFromLonLat(1, 3, &Edge{Source: 0, Target: 1}, 60, 10, 0)

	rp21 := NewRoadPositionFromLonLat(2, 0, &Edge{Source: 0, Target: 1}, 20, 20, 0)
	rp22 := NewRoadPositionFromLonLat(3, 3, &Edge{Source: 0, Target: 1}, 60, 20, 0)

	rp31 := NewRoadPositionFromLonLat(4, 1, &Edge{Source: 0, Target: 1}, 20, 40, 0)
	rp32 := NewRoadPositionFromLonLat(5, 1, &Edge{Source: 0, Target: 1}, 30, 50, 0)
	rp33 := NewRoadPositionFromLonLat(6, 2, &Edge{Source: 0, Target: 1}, 60, 40, 0)

	rp41 := NewRoadPositionFromLonLat(7, 4, &Edge{Source: 0, Target: 1}, 20, 70, 0)
	rp42 := NewRoadPositionFromLonLat(8, 5, &Edge{Source: 0, Target: 1}, 60, 70, 0)

	states := RoadPositions{rp11, rp12, rp21, rp22, rp31, rp32, rp33, rp41, rp42}

	obsState := make(map[int]*CandidateLayer)

	obsState[gpsMeasurements[0].id] = NewCandidateLayer(gpsMeasurements[0], RoadPositions{rp11, rp12})
	obsState[gpsMeasurements[1].id] = NewCandidateLayer(gpsMeasurements[1], RoadPositions{rp21, rp22})
	obsState[gpsMeasurements[2].id] = NewCandidateLayer(gpsMeasurements[2], RoadPositions{rp31, rp32, rp33})
	obsState[gpsMeasurements[3].id] = NewCandidateLayer(gpsMeasurements[3], RoadPositions{rp41, rp42})

	routeLengths := make(lengths)

	routeLengths.AddRouteLength(rp11, rp21, 10)
	routeLengths.AddRouteLength(rp11, rp22, 110)
	routeLengths.AddRouteLength(rp12, rp21, 110)
	routeLengths.AddRouteLength(rp12, rp22, 10)

	routeLengths.AddRouteLength(rp21, rp31, 20)
	routeLengths.AddRouteLength(rp21, rp32, 40)
	routeLengths.AddRouteLength(rp21, rp33, 80)
	routeLengths.AddRouteLength(rp22, rp31, 80)
	routeLengths.AddRouteLength(rp22, rp32, 60)
	routeLengths.AddRouteLength(rp22, rp33, 20)

	routeLengths.AddRouteLength(rp31, rp41, 30)
	routeLengths.AddRouteLength(rp31, rp42, 70)
	routeLengths.AddRouteLength(rp32, rp41, 30)
	routeLengths.AddRouteLength(rp32, rp42, 50)
	routeLengths.AddRouteLength(rp33, rp41, 70)
	routeLengths.AddRouteLength(rp33, rp42, 30)

	v, err := matcher.PrepareViterbi(obsState, routeLengths, gpsMeasurements)
	if err != nil {
		t.Error(err)
	}
	vpath := v.EvalPathLogProbabilities()
	correctProb := -1926.893407386203
	eps := 10e-6
	if math.Abs(vpath.Probability-correctProb) > eps {
		t.Errorf(
			"probability has to be %f, but got %f", correctProb, vpath.Probability,
		)
	}
	if len(vpath.Path) != 4 {
		t.Error(
			"length of found path has to be 4, but got", len(vpath.Path),
		)
	}
	if vpath.Path[0] != states[0] {
		t.Error(
			"First state has to be r11, but got", vpath.Path[0],
		)
	}
	if vpath.Path[1] != states[2] {
		t.Error(
			"Second state has to be r11, but got", vpath.Path[1],
		)
	}
	if vpath.Path[2] != states[4] {
		t.Error(
			"Third state has to be r31, but got", vpath.Path[2],
		)
	}
	if vpath.Path[3] != states[7] {
		t.Error(
			"Fourth state has to be r41, but got", vpath.Path[3],
		)
	}
}

func BenchmarkMapMatcherSRID_0(b *testing.B) {
	b.Log("Please wait until initial data is loaded (SRID 0, small graph)")
	matcher := NewMapMatcherDefault()
	gpsMeasurements := GPSMeasurements{
		NewGPSMeasurementFromID(1, 10, 10, 0),
		NewGPSMeasurementFromID(2, 30, 20, 0),
		NewGPSMeasurementFromID(3, 30, 40, 0),
		NewGPSMeasurementFromID(4, 10, 70, 0),
	}
	rp11 := NewRoadPositionFromLonLat(0, 0, &Edge{Source: 0, Target: 1}, 20, 10, 0)
	rp12 := NewRoadPositionFromLonLat(1, 3, &Edge{Source: 0, Target: 1}, 60, 10, 0)
	rp21 := NewRoadPositionFromLonLat(2, 0, &Edge{Source: 0, Target: 1}, 20, 20, 0)
	rp22 := NewRoadPositionFromLonLat(3, 3, &Edge{Source: 0, Target: 1}, 60, 20, 0)
	rp31 := NewRoadPositionFromLonLat(4, 1, &Edge{Source: 0, Target: 1}, 20, 40, 0)
	rp32 := NewRoadPositionFromLonLat(5, 1, &Edge{Source: 0, Target: 1}, 30, 50, 0)
	rp33 := NewRoadPositionFromLonLat(6, 2, &Edge{Source: 0, Target: 1}, 60, 40, 0)
	rp41 := NewRoadPositionFromLonLat(7, 4, &Edge{Source: 0, Target: 1}, 20, 70, 0)
	rp42 := NewRoadPositionFromLonLat(8, 5, &Edge{Source: 0, Target: 1}, 60, 70, 0)

	obsState := make(map[int]*CandidateLayer)
	obsState[gpsMeasurements[0].id] = NewCandidateLayer(gpsMeasurements[0], RoadPositions{rp11, rp12})
	obsState[gpsMeasurements[1].id] = NewCandidateLayer(gpsMeasurements[1], RoadPositions{rp21, rp22})
	obsState[gpsMeasurements[2].id] = NewCandidateLayer(gpsMeasurements[2], RoadPositions{rp31, rp32, rp33})
	obsState[gpsMeasurements[3].id] = NewCandidateLayer(gpsMeasurements[3], RoadPositions{rp41, rp42})

	routeLengths := make(lengths)

	routeLengths.AddRouteLength(rp11, rp21, 10)
	routeLengths.AddRouteLength(rp11, rp22, 110)
	routeLengths.AddRouteLength(rp12, rp21, 110)
	routeLengths.AddRouteLength(rp12, rp22, 10)

	routeLengths.AddRouteLength(rp21, rp31, 20)
	routeLengths.AddRouteLength(rp21, rp32, 40)
	routeLengths.AddRouteLength(rp21, rp33, 80)
	routeLengths.AddRouteLength(rp22, rp31, 80)
	routeLengths.AddRouteLength(rp22, rp32, 60)
	routeLengths.AddRouteLength(rp22, rp33, 20)

	routeLengths.AddRouteLength(rp31, rp41, 30)
	routeLengths.AddRouteLength(rp31, rp42, 70)
	routeLengths.AddRouteLength(rp32, rp41, 30)
	routeLengths.AddRouteLength(rp32, rp42, 50)
	routeLengths.AddRouteLength(rp33, rp41, 70)
	routeLengths.AddRouteLength(rp33, rp42, 30)

	b.Log("BenchmarkMapMatcherSRID_0 is starting...")
	b.ResetTimer()

	for k := 0.; k <= 12; k++ {
		n := int(math.Pow(2, k))
		b.Run(fmt.Sprintf("%s/%d/pts-%d", "Map match for Euclidean points", n, len(gpsMeasurements)), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				v, err := matcher.PrepareViterbi(obsState, routeLengths, gpsMeasurements)
				if err != nil {
					b.Error(err)
				}
				vpath := v.EvalPathLogProbabilities()
				_ = vpath
			}
		})
	}
}
