package horizon

import (
	"github.com/LdDl/horizon/spatial"
	"github.com/golang/geo/s2"
)

// ResponseGeometry contains read-only geometry for one observation in an API response.
// Unchanged geometry can share storage with the immutable road graph.
type ResponseGeometry struct {
	Matched   s2.Polyline
	Cut       s2.Polyline
	NextEdges []EdgeResult
}

// ResponseGeometries clips consecutive boundary observations on the same edge.
// A return through other edges starts a separate visit. Nonmonotone visits keep
// their full geometry so clipping does not hide an observed excursion.
func (match SubMatch) ResponseGeometries() []ResponseGeometry {
	observations := match.Observations
	result := make([]ResponseGeometry, len(observations))
	for i, observation := range observations {
		if observation.IsMatched {
			result[i].NextEdges = observation.NextEdges
			if observation.MatchedEdge.Polyline != nil {
				result[i].Matched = *observation.MatchedEdge.Polyline
			}
		}
	}
	if len(observations) == 0 {
		return result
	}
	first, last := observations[0], observations[len(observations)-1]
	leadingEnd := 0
	if hasResponseProjection(first) {
		leadingEnd = 1
		for leadingEnd < len(observations) && len(observations[leadingEnd-1].NextEdges) == 0 &&
			hasResponseProjection(observations[leadingEnd]) && observations[leadingEnd].MatchedEdge.ID == first.MatchedEdge.ID {
			leadingEnd++
		}
		if forwardProjectionRun(observations[:leadingEnd]) {
			suffix, prefix := spatial.ExtractCutUpTo(*first.MatchedEdge.Polyline, first.ProjectedPoint, first.ProjectionPointIdx)
			for i := 0; i < leadingEnd; i++ {
				result[i].Matched = suffix
			}
			result[0].Cut = prefix
		}
	}
	// Preserve the existing first-point convention for a singleton segment.
	if len(observations) == 1 || !hasResponseProjection(last) {
		return result
	}
	trailingStart := len(observations) - 1
	for trailingStart > 0 && len(observations[trailingStart-1].NextEdges) == 0 &&
		hasResponseProjection(observations[trailingStart-1]) && observations[trailingStart-1].MatchedEdge.ID == last.MatchedEdge.ID {
		trailingStart--
	}
	if !forwardProjectionRun(observations[trailingStart:]) {
		return result
	}
	line, next := *last.MatchedEdge.Polyline, last.ProjectionPointIdx
	if leadingEnd == len(observations) && result[0].Cut != nil {
		// The first clipping changes vertex indices in the retained suffix.
		line = result[0].Matched
		next -= first.ProjectionPointIdx - 1
	}
	prefix, suffix := spatial.ExtractCutUpFrom(line, last.ProjectedPoint, next)
	for i := trailingStart; i < len(observations); i++ {
		result[i].Matched = prefix
	}
	result[len(observations)-1].Cut = suffix
	if trailingStart > 0 {
		// The last incoming route edge can duplicate the terminal matched edge.
		incoming := observations[trailingStart-1].NextEdges
		if len(incoming) > 0 && incoming[len(incoming)-1].ID == last.MatchedEdge.ID {
			clipped := append([]EdgeResult(nil), incoming...)
			clipped[len(clipped)-1].Geom = prefix
			result[trailingStart-1].NextEdges = clipped
		}
	}
	return result
}

func hasResponseProjection(observation ObservationResult) bool {
	return observation.IsMatched && observation.MatchedEdge.Polyline != nil &&
		observation.ProjectionPointIdx >= 1 && observation.ProjectionPointIdx <= len(*observation.MatchedEdge.Polyline)
}

func forwardProjectionRun(observations []ObservationResult) bool {
	for i := 1; i < len(observations); i++ {
		previous, current := observations[i-1], observations[i]
		if current.ProjectionPointIdx < previous.ProjectionPointIdx {
			return false
		}
		if current.ProjectionPointIdx != previous.ProjectionPointIdx {
			continue
		}
		origin := (*current.MatchedEdge.Polyline)[current.ProjectionPointIdx-1]
		var previousDistance, currentDistance float64
		if current.Observation != nil && current.Observation.SRID() != 4326 {
			previousDistance = previous.ProjectedPoint.Sub(origin.Vector).Norm2()
			currentDistance = current.ProjectedPoint.Sub(origin.Vector).Norm2()
		} else {
			previousDistance = previous.ProjectedPoint.Distance(origin).Radians()
			currentDistance = current.ProjectedPoint.Distance(origin).Radians()
		}
		if currentDistance < previousDistance {
			return false
		}
	}
	return true
}
