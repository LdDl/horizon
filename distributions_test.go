package horizon

import (
	"math"
	"testing"
)

var (
	testSigma = 5.0
	testX     = 6.0
	delta     = 8
)

func TestLogNormalDistribution(t *testing.T) {
	v0 := math.Log(NormalDistribution(testSigma, testX))
	v1 := LogNormalDistribution(testSigma, testX)
	if toFixed(v0, delta) != toFixed(v1, delta) {
		t.Error(
			"For sigma", testSigma,
			"and x", testX,
			"expected", v1,
			"got", v0,
		)
	}
}

func TestLogExponentialDistribution(t *testing.T) {
	v0 := math.Log(ExponentialDistribution(testSigma, testX))
	v1 := LogExponentialDistribution(testSigma, testX)
	if v0 != v1 {
		t.Error(
			"For sigma", testSigma,
			"and x", testX,
			"expected", v1,
			"got", v0,
		)
	}
}
