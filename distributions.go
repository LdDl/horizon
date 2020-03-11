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

// LogNormalDistribution https://en.wikipedia.org/wiki/Log-normal_distribution
func LogNormalDistribution(sigma, x float64) float64 {
	return math.Log(1.0/(sqrtTwoPi*sigma)) + (-0.5 * math.Pow(x/sigma, 2))
}

// ExponentialDistribution 1 / (β*exp(-x/β)), beta = 1/λ
func ExponentialDistribution(beta, x float64) float64 {
	return 1.0 / beta * math.Exp(-x/beta)
}

// LogExponentialDistribution ln(1/β) - (x/β), beta = 1/λ
func LogExponentialDistribution(beta, x float64) float64 {
	return math.Log(1.0/beta) - (x / beta)
}
