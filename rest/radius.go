package rest

import "math"

// Default radius values (in meters)
const (
	// Default search radius for map matching
	DEFAULT_STATE_RADIUS = 50.0
	// Default search radius for shortest path (more permissive)
	DEFAULT_SP_RADIUS = 100.0
)

// ResolveRadius resolves the radius value based on the API design:
//   - < 0: no limit (returns -1 to signal unlimited search)
//   - 0 or nil: use default
//   - > 0: use provided value
func ResolveRadius(radius *float64, defaultValue float64) float64 {
	if radius == nil {
		return defaultValue
	}
	if math.Abs(*radius) < 1e-9 {
		return defaultValue
	}
	return *radius
}
