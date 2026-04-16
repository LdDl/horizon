package horizon

import "math"

// HmmProbabilities Parameters used in evaluating of Normal Distribution and Exponentional Distribution
type HmmProbabilities struct {
	sigma float64
	beta  float64
	// precomputed 1/beta (loop-invariant)
	invBeta float64
	// precomputed log(1/beta) (loop-invariant)
	logInvBeta float64
}

// HmmProbabilitiesDefault Constructor for creating HmmProbabilities with default values
// Sigma - standard deviation of the normal distribution [m] used for modeling the GPS error
// Beta - beta parameter of the exponential distribution used for modeling transition probabilities
func HmmProbabilitiesDefault() *HmmProbabilities {
	return NewHmmProbabilities(4.07, 0.00959442)
}

// NewHmmProbabilities Constructor for creating HmmProbabilities with provided values
func NewHmmProbabilities(sigma, beta float64) *HmmProbabilities {
	invBeta := 1.0 / beta
	return &HmmProbabilities{
		sigma:      sigma,
		beta:       beta,
		invBeta:    invBeta,
		logInvBeta: math.Log(invBeta),
	}
}

// EmissionProbability Evaluate emission probability (normal distribution is used). Absolute distance [m] between GPS measurement and map matching candidate.
func (hp *HmmProbabilities) EmissionProbability(value float64) float64 {
	return NormalDistribution(hp.sigma, value)
}

// EmissionLogProbability Evaluate emission probability (log-normal distribution is used)
func (hp *HmmProbabilities) EmissionLogProbability(value float64) float64 {
	return LogNormalDistribution(hp.sigma, value)
}

// TransitionProbability Evaluate transition probability (exponential distribution is used)
func (hp *HmmProbabilities) TransitionProbability(routeLength, linearDistance, timeDiff float64) (float64, error) {
	transitionMetric, err := hp.normalizedTransitionMetric(routeLength, linearDistance, timeDiff)
	if err != nil {
		return 0, err
	}
	return ExponentialDistribution(hp.beta, transitionMetric), nil
}

// TransitionLogProbability Evaluate transition probability (log-exponential distribution is used)
func (hp *HmmProbabilities) TransitionLogProbability(routeLength, linearDistance, timeDiff float64) (float64, error) {
	transitionMetric, err := hp.normalizedTransitionMetric(routeLength, linearDistance, timeDiff)
	if err != nil {
		return 0, err
	}
	// log(1/beta) - x/beta, with both constants precomputed
	return hp.logInvBeta - transitionMetric*hp.invBeta, nil
}

// normalizedTransitionMetric
func (hp *HmmProbabilities) normalizedTransitionMetric(routeLength, linearDistance, timeDiff float64) (float64, error) {
	if timeDiff < 0.0 {
		return 0.0, ErrTimeDifference
	}
	// return math.Abs(linearDistance-routeLength) / (timeDiff * timeDiff), nil
	return math.Abs(linearDistance - routeLength), nil
}
