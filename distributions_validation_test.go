package horizon

import (
	"math"
	"testing"
)

// TestLogDistributionsUnnormalized verifies that unnormalized log distributions are always <= 0
func TestLogDistributionsUnnormalized(t *testing.T) {
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

	// Test LogExponentialDistributionUnnormalized - should always be <= 0
	for _, beta := range betas {
		for _, x := range xValues {
			logProb := LogExponentialDistributionUnnormalized(beta, x)
			if logProb > 0 {
				t.Errorf("LogExponentialDistributionUnnormalized(beta=%f, x=%f) = %f should be <= 0",
					beta, x, logProb)
			}
			// Verify formula: -x/beta
			expected := -x / beta
			if math.Abs(logProb-expected) > 1e-10 {
				t.Errorf("LogExponentialDistributionUnnormalized(beta=%f, x=%f) = %f, expected %f",
					beta, x, logProb, expected)
			}
		}
	}

	// Test LogNormalDistributionUnnormalized - should always be <= 0
	sigmas := []float64{
		0.5,
		1.0,
		4.07, // Default sigma
		10.0,
	}

	for _, sigma := range sigmas {
		for _, x := range xValues {
			logProb := LogNormalDistributionUnnormalized(sigma, x)
			if logProb > 0 {
				t.Errorf("LogNormalDistributionUnnormalized(sigma=%f, x=%f) = %f should be <= 0",
					sigma, x, logProb)
			}
			// Verify formula: -0.5*(x/sigma)^2
			expected := -0.5 * math.Pow(x/sigma, 2)
			if math.Abs(logProb-expected) > 1e-10 {
				t.Errorf("LogNormalDistributionUnnormalized(sigma=%f, x=%f) = %f, expected %f",
					sigma, x, logProb, expected)
			}
		}
	}
}

// TestDefaultBetaNoPositiveValues verifies no positive log probabilities with default beta for unnormalized
func TestDefaultBetaNoPositiveValues(t *testing.T) {
	beta := 0.00959442 // Default beta
	x := 0.0

	logProb := LogExponentialDistributionUnnormalized(beta, x)

	// With unnormalized formula, x=0 gives 0, not positive value
	if logProb != 0 {
		t.Errorf("LogExponentialDistributionUnnormalized(beta=%f, x=0) = %f, expected 0", beta, logProb)
	}

	// All values should be <= 0
	if logProb > 0 {
		t.Errorf("LogExponentialDistributionUnnormalized should never return positive values, got %f", logProb)
	}

	t.Logf("LogExponentialDistributionUnnormalized(beta=%f, x=0) = %f (no more positive values!)", beta, logProb)
}
