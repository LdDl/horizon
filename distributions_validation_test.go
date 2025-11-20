package horizon

import (
	"math"
	"testing"
)

// TestLogProbabilitiesCorrectness verifies that log probability functions match their regular counterparts
func TestLogProbabilitiesCorrectness(t *testing.T) {
	betas := []float64{
		0.00959442, // Default beta
		0.001,
		0.01,
		0.1,
		0.5,
		1.0,
		2.0,
		10.0,
	}

	xValues := []float64{
		0.0,
		0.001,
		0.1,
		1.0,
		10.0,
		100.0,
		1000.0,
	}

	// Test LogExponentialDistribution
	for _, beta := range betas {
		for _, x := range xValues {
			logProb := LogExponentialDistribution(beta, x)
			regularProb := ExponentialDistribution(beta, x)

			if regularProb > 0 {
				expectedLogProb := math.Log(regularProb)
				eps := 1e-10
				if math.Abs(logProb-expectedLogProb) > eps {
					t.Errorf("LogExponentialDistribution(beta=%f, x=%f) = %f doesn't match log(ExponentialDistribution) = %f",
						beta, x, logProb, expectedLogProb)
				}
			} else {
				// Handle numerical underflow
				if logProb > -100 {
					t.Errorf("LogExponentialDistribution(beta=%f, x=%f) = %f should be very negative when regular prob underflows",
						beta, x, logProb)
				}
			}
		}
	}

	// Test LogNormalDistribution
	sigmas := []float64{
		0.5,
		1.0,
		4.07, // Default sigma
		10.0,
	}

	for _, sigma := range sigmas {
		for _, x := range xValues {
			logProb := LogNormalDistribution(sigma, x)
			regularProb := NormalDistribution(sigma, x)

			if regularProb > 0 {
				expectedLogProb := math.Log(regularProb)
				eps := 1e-10
				if math.Abs(logProb-expectedLogProb) > eps {
					t.Errorf("LogNormalDistribution(sigma=%f, x=%f) = %f doesn't match log(NormalDistribution) = %f",
						sigma, x, logProb, expectedLogProb)
				}
			} else {
				// Handle numerical underflow
				if logProb > -100 {
					t.Errorf("LogNormalDistribution(sigma=%f, x=%f) = %f should be very negative when regular prob underflows",
						sigma, x, logProb)
				}
			}
		}
	}
}

// TestDefaultBetaEdgeCase tests exponential distribution with default beta at x=0
func TestDefaultBetaEdgeCase(t *testing.T) {
	beta := 0.00959442 // Default beta
	x := 0.0

	logProb := LogExponentialDistribution(beta, x)
	expectedLogProb := math.Log(1.0 / beta)

	eps := 1e-10
	if math.Abs(logProb-expectedLogProb) > eps {
		t.Errorf("LogExponentialDistribution(beta=%f, x=0) = %f, expected %f",
			beta, logProb, expectedLogProb)
	}

	// Verify that PDF > 1 is valid for small beta
	if beta < 1.0 && logProb > 0 {
		// This is expected: PDF can exceed 1 for continuous distributions
		t.Logf("LogExponentialDistribution(beta=%f, x=0) = %f (positive, as expected for PDF > 1)", beta, logProb)
	}
}
