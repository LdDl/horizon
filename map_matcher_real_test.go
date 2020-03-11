package horizon

import (
	"testing"
	"time"
)

func TestMapMatcher_BIG(t *testing.T) {

	var (
		graphFileName   = "data/osm2ch_export.csv"
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

		correctStates = MatcherResult{
			Observations: []*ObservationResult{
				&ObservationResult{Observation: gpsMeasurements[0]},
				&ObservationResult{Observation: gpsMeasurements[1]},
				&ObservationResult{Observation: gpsMeasurements[2]},
				&ObservationResult{Observation: gpsMeasurements[3]},
				&ObservationResult{Observation: gpsMeasurements[4]},
				&ObservationResult{Observation: gpsMeasurements[5]},
				&ObservationResult{Observation: gpsMeasurements[5]},
				&ObservationResult{Observation: gpsMeasurements[7]},
				&ObservationResult{Observation: gpsMeasurements[8]},
				&ObservationResult{Observation: gpsMeasurements[9]},
			},
			Probability: -81.741322,
		}
	)

	hmmParams := NewHmmProbabilities(sigma, beta)
	matcher, err := NewMapMatcher(hmmParams, graphFileName)
	if err != nil {
		t.Error(err)
	}

	correctStates.Observations[0].MatchedEdge = *matcher.engine.edges[10099][10100]
	correctStates.Observations[1].MatchedEdge = *matcher.engine.edges[10109][10110]
	correctStates.Observations[2].MatchedEdge = *matcher.engine.edges[10118][10119]
	correctStates.Observations[3].MatchedEdge = *matcher.engine.edges[10120][10121]
	correctStates.Observations[4].MatchedEdge = *matcher.engine.edges[10122][10123]
	correctStates.Observations[5].MatchedEdge = *matcher.engine.edges[10123][10124]
	correctStates.Observations[6].MatchedEdge = *matcher.engine.edges[10124][10125]
	correctStates.Observations[7].MatchedEdge = *matcher.engine.edges[12276][12277]
	correctStates.Observations[8].MatchedEdge = *matcher.engine.edges[12280][12281]
	correctStates.Observations[9].MatchedEdge = *matcher.engine.edges[21762][21763]

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
		if result.Observations[i].MatchedEdge != correctStates.Observations[i].MatchedEdge {
			t.Errorf("Matched edge for observation %d should be %d->%d, but got %d->%d",
				result.Observations[i].Observation.id,
				correctStates.Observations[i].MatchedEdge.Source, correctStates.Observations[i].MatchedEdge.Target,
				result.Observations[i].MatchedEdge.Source, result.Observations[i].MatchedEdge.Target,
			)
		}
	}

}
