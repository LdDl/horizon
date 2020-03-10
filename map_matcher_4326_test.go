package horizon

import "testing"

func TestMapMatcherSRID_4326(t *testing.T) {

	var (
		graphFileName   = "data/matcher_4326_test.csv"
		sigma           = 50.0
		beta            = 2.0
		gpsMeasurements = GPSMeasurements{
			NewGPSMeasurementFromID(1, 37.662745994981435, 55.77323867786974, 4326),
			NewGPSMeasurementFromID(2, 37.66373679411533, 55.77352528537278, 4326),
			NewGPSMeasurementFromID(3, 37.6634658408828, 55.77408712095024, 4326),
			NewGPSMeasurementFromID(4, 37.66271768643477, 55.77491052526131, 4326),
		}

		correctStates = MatcherResult{
			Observations: []*ObservationResult{
				&ObservationResult{Observation: gpsMeasurements[0]},
				&ObservationResult{Observation: gpsMeasurements[1]},
				&ObservationResult{Observation: gpsMeasurements[2]},
				&ObservationResult{Observation: gpsMeasurements[3]},
			},
			Probability: -95.148024,
		}
	)

	var err error
	matcher := NewMapMatcher(NewHmmProbabilities(sigma, beta))
	matcher.engine, err = prepareEngine(graphFileName)
	if err != nil {
		t.Error(err)
	}

	correctStates.Observations[0].MatchedEdge = *matcher.engine.edges[101][102]
	correctStates.Observations[1].MatchedEdge = *matcher.engine.edges[101][102]
	correctStates.Observations[2].MatchedEdge = *matcher.engine.edges[101][102]
	correctStates.Observations[3].MatchedEdge = *matcher.engine.edges[102][105]

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
