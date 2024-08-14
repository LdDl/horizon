package horizon

import (
	"fmt"
	"math"
	"testing"
	"time"
)

func TestMapMatcherSRID_4326(t *testing.T) {

	var (
		currentTime     = time.Now()
		graphFileName   = "./test_data/matcher_4326_test.csv"
		sigma           = 50.0
		beta            = 2.0
		gpsMeasurements = GPSMeasurements{
			NewGPSMeasurement(1, 37.662745994981435, 55.77323867786974, 4326, WithGPSTime(currentTime.Add(1*time.Second))),
			NewGPSMeasurement(2, 37.66373679411533, 55.77352528537278, 4326, WithGPSTime(currentTime.Add(2*time.Second))),
			NewGPSMeasurement(3, 37.6634658408828, 55.77408712095024, 4326, WithGPSTime(currentTime.Add(3*time.Second))),
			NewGPSMeasurement(4, 37.66271768643477, 55.77491052526131, 4326, WithGPSTime(currentTime.Add(4*time.Second))),
		}

		correctStates = MatcherResult{
			Observations: []ObservationResult{
				{Observation: gpsMeasurements[0]},
				{Observation: gpsMeasurements[1]},
				{Observation: gpsMeasurements[2]},
				{Observation: gpsMeasurements[3]},
			},
			Probability: -52.195440,
		}
	)

	hmmParams := NewHmmProbabilities(sigma, beta)
	matcher, err := NewMapMatcher(hmmParams, graphFileName)
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

	eps := 10e-6
	if math.Abs(result.Probability-correctStates.Probability) > eps {
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

func BenchmarkMapMatcherSRID_4326(b *testing.B) {
	b.Log("Please wait until initial data is loaded (SRID 4326, small graph)")
	var (
		graphFileName   = "./test_data/matcher_4326_test.csv"
		sigma           = 50.0
		beta            = 2.0
		gpsMeasurements = GPSMeasurements{
			NewGPSMeasurementFromID(1, 37.662745994981435, 55.77323867786974, 4326),
			NewGPSMeasurementFromID(2, 37.66373679411533, 55.77352528537278, 4326),
			NewGPSMeasurementFromID(3, 37.6634658408828, 55.77408712095024, 4326),
			NewGPSMeasurementFromID(4, 37.66271768643477, 55.77491052526131, 4326),
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
		b.Run(fmt.Sprintf("%s/%d/pts-%d", "Map match for WGS84 points (small graph)", n, len(gpsMeasurements)), func(b *testing.B) {
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
