package horizon

import (
	"fmt"
	"math"
	"testing"
	"time"
)

func TestMapMatcher_4326BIG(t *testing.T) {

	var (
		graphFileName   = "./test_data/osm2ch_export.csv"
		sigma           = 50.0
		beta            = 30.0
		gpsMeasurements = GPSMeasurements{
			NewGPSMeasurement(time.Date(1, 1, 1, 1, 1, 1, 0, time.UTC), 37.601249363208915, 55.745374309126895, 4326),
			NewGPSMeasurement(time.Date(1, 1, 1, 1, 1, 2, 0, time.UTC), 37.600552781226014, 55.746223820101498, 4326),
			NewGPSMeasurement(time.Date(1, 1, 1, 1, 1, 3, 0, time.UTC), 37.599959396573908, 55.747450858855984, 4326),
			NewGPSMeasurement(time.Date(1, 1, 1, 1, 1, 4, 0, time.UTC), 37.600526981893317, 55.748017171419498, 4326),
			NewGPSMeasurement(time.Date(1, 1, 1, 1, 1, 5, 0, time.UTC), 37.600655978556816, 55.748728680680564, 4326),
			NewGPSMeasurement(time.Date(1, 1, 1, 1, 1, 6, 0, time.UTC), 37.600372185897115, 55.749454697162832, 4326),
			NewGPSMeasurement(time.Date(1, 1, 1, 1, 1, 7, 0, time.UTC), 37.600694677555865, 55.750521916863391, 4326),
			NewGPSMeasurement(time.Date(1, 1, 1, 1, 1, 8, 0, time.UTC), 37.600965570549214, 55.751371315759044, 4326),
			NewGPSMeasurement(time.Date(1, 1, 1, 1, 1, 9, 0, time.UTC), 37.600926871550165, 55.752634490168425, 4326),
			NewGPSMeasurement(time.Date(1, 1, 1, 1, 1, 10, 0, time.UTC), 37.60001599788666, 55.75607875029978, 4326),
		}

		correctStates = MatcherResult{
			Observations: []*ObservationResult{
				{Observation: gpsMeasurements[0]},
				{Observation: gpsMeasurements[1]},
				{Observation: gpsMeasurements[2]},
				{Observation: gpsMeasurements[3]},
				{Observation: gpsMeasurements[4]},
				{Observation: gpsMeasurements[5]},
				{Observation: gpsMeasurements[5]},
				{Observation: gpsMeasurements[7]},
				{Observation: gpsMeasurements[8]},
				{Observation: gpsMeasurements[9]},
			},
			Probability: -81.716795,
		}
	)

	hmmParams := NewHmmProbabilities(sigma, beta)
	matcher, err := NewMapMatcher(hmmParams, graphFileName)
	if err != nil {
		t.Error(err)
	}

	correctStates.Observations[0].MatchedEdge = *matcher.engine.edges[13640][13641]
	correctStates.Observations[1].MatchedEdge = *matcher.engine.edges[13650][13651]
	correctStates.Observations[2].MatchedEdge = *matcher.engine.edges[13659][13660]
	correctStates.Observations[3].MatchedEdge = *matcher.engine.edges[13661][13662]
	correctStates.Observations[4].MatchedEdge = *matcher.engine.edges[13663][13664]
	correctStates.Observations[5].MatchedEdge = *matcher.engine.edges[13664][13665]
	correctStates.Observations[6].MatchedEdge = *matcher.engine.edges[13665][13666]
	correctStates.Observations[7].MatchedEdge = *matcher.engine.edges[16784][16785]
	correctStates.Observations[8].MatchedEdge = *matcher.engine.edges[16788][16789]
	correctStates.Observations[9].MatchedEdge = *matcher.engine.edges[32639][32640]

	statesRadiusMeters := 7.0
	maxStates := 5
	result, err := matcher.Run(gpsMeasurements, statesRadiusMeters, maxStates)
	if err != nil {
		t.Error(err)
	}

	if len(result.Observations) != len(correctStates.Observations) {
		t.Errorf("Result should contain %d measurements, but got %d", len(correctStates.Observations), len(result.Observations))
	}

	if toFixed(result.Probability, 6) != toFixed(correctStates.Probability, 6) {
		t.Errorf("Path's probability should be %f, but got %f", correctStates.Probability, result.Probability)
	}

	for i := range result.Observations {
		if result.Observations[i].MatchedEdge.Source != correctStates.Observations[i].MatchedEdge.Source &&
			result.Observations[i].MatchedEdge.Source != correctStates.Observations[i].MatchedEdge.Target &&
			result.Observations[i].MatchedEdge.Target != correctStates.Observations[i].MatchedEdge.Source {
			t.Errorf("Matched edge for observation %d should be %d->%d, but got %d->%d",
				result.Observations[i].Observation.id,
				correctStates.Observations[i].MatchedEdge.Source, correctStates.Observations[i].MatchedEdge.Target,
				result.Observations[i].MatchedEdge.Source, result.Observations[i].MatchedEdge.Target,
			)
		}
	}

}

func BenchmarkMapMatcherSRID_4326BIG(b *testing.B) {
	b.Log("Please wait until initial data is loaded (SRID 4326, average graph)")
	var (
		graphFileName   = "./test_data/osm2ch_export.csv"
		sigma           = 50.0
		beta            = 30.0
		gpsMeasurements = GPSMeasurements{
			NewGPSMeasurement(time.Date(1, 1, 1, 1, 1, 1, 0, time.UTC), 37.601249363208915, 55.745374309126895, 4326),
			NewGPSMeasurement(time.Date(1, 1, 1, 1, 1, 2, 0, time.UTC), 37.600552781226014, 55.746223820101498, 4326),
			NewGPSMeasurement(time.Date(1, 1, 1, 1, 1, 3, 0, time.UTC), 37.599959396573908, 55.747450858855984, 4326),
			NewGPSMeasurement(time.Date(1, 1, 1, 1, 1, 4, 0, time.UTC), 37.600526981893317, 55.748017171419498, 4326),
			NewGPSMeasurement(time.Date(1, 1, 1, 1, 1, 5, 0, time.UTC), 37.600655978556816, 55.748728680680564, 4326),
			NewGPSMeasurement(time.Date(1, 1, 1, 1, 1, 6, 0, time.UTC), 37.600372185897115, 55.749454697162832, 4326),
			NewGPSMeasurement(time.Date(1, 1, 1, 1, 1, 7, 0, time.UTC), 37.600694677555865, 55.750521916863391, 4326),
			NewGPSMeasurement(time.Date(1, 1, 1, 1, 1, 8, 0, time.UTC), 37.600965570549214, 55.751371315759044, 4326),
			NewGPSMeasurement(time.Date(1, 1, 1, 1, 1, 9, 0, time.UTC), 37.600926871550165, 55.752634490168425, 4326),
			NewGPSMeasurement(time.Date(1, 1, 1, 1, 1, 10, 0, time.UTC), 37.600385085563467, 55.755596255965337, 4326),
		}
	)

	hmmParams := NewHmmProbabilities(sigma, beta)
	matcher, err := NewMapMatcher(hmmParams, graphFileName)
	if err != nil {
		b.Error(err)
	}

	statesRadiusMeters := 7.0
	maxStates := 5

	b.Log("BenchmarkMapMatcherSRID_4326 is starting...")
	b.ResetTimer()

	for k := 0.; k <= 12; k++ {
		n := int(math.Pow(2, k))
		b.Run(fmt.Sprintf("%s/%d/pts-%d", "Map match for WGS84 points (average graph)", n, len(gpsMeasurements)), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				result, err := matcher.Run(gpsMeasurements, statesRadiusMeters, maxStates)
				if err != nil {
					b.Error(err)
				}
				_ = result
			}
		})
	}
}
