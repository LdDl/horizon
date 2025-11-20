package horizon

import (
	"math"
	"testing"
)

var (
	testSigma = 5.0
	testX     = 6.0
)

func TestLogNormalDistribution(t *testing.T) {
	// Test normalized formula: log(1/(sigma*sqrt(2*pi))) - 0.5*(x/sigma)^2
	sqrtTwoPi := math.Sqrt(2.0 * math.Pi)
	expected := math.Log(1.0/(sqrtTwoPi*testSigma)) + (-0.5 * math.Pow(testX/testSigma, 2))
	result := LogNormalDistribution(testSigma, testX)
	eps := 1e-10
	if math.Abs(result-expected) > eps {
		t.Errorf("LogNormalDistribution(sigma=%f, x=%f) = %f, expected %f",
			testSigma, testX, result, expected)
	}
}

func TestLogNormalDistributionUnnormalized(t *testing.T) {
	// Test unnormalized formula: -0.5*(x/sigma)^2
	expected := -0.5 * math.Pow(testX/testSigma, 2)
	result := LogNormalDistributionUnnormalized(testSigma, testX)
	eps := 1e-10
	if math.Abs(result-expected) > eps {
		t.Errorf("LogNormalDistributionUnnormalized(sigma=%f, x=%f) = %f, expected %f",
			testSigma, testX, result, expected)
	}
	// Verify always <= 0
	if result > 0 {
		t.Errorf("LogNormalDistributionUnnormalized should always be <= 0, got %f", result)
	}
}

func TestLogExponentialDistribution(t *testing.T) {
	// Test normalized formula: log(1/beta) - x/beta
	expected := math.Log(1.0/testSigma) - (testX / testSigma)
	result := LogExponentialDistribution(testSigma, testX)
	eps := 1e-10
	if math.Abs(result-expected) > eps {
		t.Errorf("LogExponentialDistribution(beta=%f, x=%f) = %f, expected %f",
			testSigma, testX, result, expected)
	}
}

func TestLogExponentialDistributionUnnormalized(t *testing.T) {
	// Test unnormalized formula: -x/beta
	expected := -testX / testSigma
	result := LogExponentialDistributionUnnormalized(testSigma, testX)
	eps := 1e-10
	if math.Abs(result-expected) > eps {
		t.Errorf("LogExponentialDistributionUnnormalized(beta=%f, x=%f) = %f, expected %f",
			testSigma, testX, result, expected)
	}
	// Verify always <= 0
	if result > 0 {
		t.Errorf("LogExponentialDistributionUnnormalized should always be <= 0, got %f", result)
	}
}
