package horizon

import "math"

// round Round float64
func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

// toFixed Round float64 to N decimal places
func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}
