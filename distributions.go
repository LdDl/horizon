package horizon

import "math"

const (
	twoPi = 2.0 * math.Pi
)

var (
	sqrtTwoPi = math.Sqrt(twoPi)
)

// NormalDistribution https://en.wikipedia.org/wiki/Normal_distribution
func NormalDistribution(sigma, x float64) float64 {
	return 1.0 / (sqrtTwoPi * sigma) * math.Exp(-0.5*math.Pow(x/sigma, 2))
}

// LogNormalDistribution computes log of normal distribution PDF (normalized)
// log(f(x)) = log(1/(sigma*sqrt(2*pi))) - 0.5*(x/sigma)^2
// Note: Can return positive values when density > 1 (valid for PDFs)
func LogNormalDistribution(sigma, x float64) float64 {
	return math.Log(1.0/(sqrtTwoPi*sigma)) + (-0.5 * math.Pow(x/sigma, 2))
}

// LogNormalDistributionUnnormalized computes unnormalized log of normal distribution PDF
// For Viterbi/HMM we only need relative probabilities, so we drop constant terms
// log(f(x)) ~ -0.5*(x/sigma)^2
// Reference: GraphHopper/OSRM implementations
// Always returns values <= 0
func LogNormalDistributionUnnormalized(sigma, x float64) float64 {
	return -0.5 * math.Pow(x/sigma, 2)
}

// ExponentialDistribution computes (1/beta) * exp(-x/beta), where beta = 1/lambda
func ExponentialDistribution(beta, x float64) float64 {
	return (1.0 / beta) * math.Exp(-x/beta)
}

// LogExponentialDistribution computes log of exponential distribution PDF (normalized)
// log(f(x)) = log(1/beta) - x/beta
// Note: Can return positive values when beta < 1 (valid for PDFs)
func LogExponentialDistribution(beta, x float64) float64 {
	return math.Log(1.0/beta) - (x / beta)
}

// LogExponentialDistributionUnnormalized computes unnormalized log of exponential distribution PDF
// For Viterbi/HMM we only need relative probabilities, so we drop constant terms
// log(f(x)) ~ -x/beta
// Reference: GraphHopper/OSRM implementations
// Always returns values <= 0
func LogExponentialDistributionUnnormalized(beta, x float64) float64 {
	return -x / beta
}
