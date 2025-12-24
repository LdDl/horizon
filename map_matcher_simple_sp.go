package horizon

import (
	"math"

	"github.com/LdDl/horizon/spatial"
	"github.com/golang/geo/s2"
	"github.com/pkg/errors"
)

const (
	// Default number of candidates to consider when searching for a path
	DEFAULT_CANDIDATES_LIMIT = 10
)

// candidateInfo holds information about a routing candidate
type candidateInfo struct {
	edgeID       uint64
	vertex       int64
	sccComponent int64
	distance     float64
	edge         *spatial.Edge
}

// FindShortestPath finds shortest path between two observations (not necessary GPS points).
// It searches for multiple candidates and selects the best pair that are in the same connected component.
// Priority is given to candidates in the big (main) component.
//
// Parameters:
//   - source, target: GPS measurements to route between
//   - statesRadiusMeters: maximum radius to search nearest edges (use -1 for unlimited)
func (matcher *MapMatcher) FindShortestPath(source, target *GPSMeasurement, statesRadiusMeters float64) (MatcherResult, error) {
	// Get multiple candidates for source
	sourceCandidates, err := matcher.getCandidates(source.Point, statesRadiusMeters, DEFAULT_CANDIDATES_LIMIT)
	if err != nil {
		return MatcherResult{}, errors.Wrap(err, "failed to get source candidates")
	}
	if len(sourceCandidates) == 0 {
		return MatcherResult{}, ErrSourceNotFound
	}

	// Get multiple candidates for target
	targetCandidates, err := matcher.getCandidates(target.Point, statesRadiusMeters, DEFAULT_CANDIDATES_LIMIT)
	if err != nil {
		return MatcherResult{}, errors.Wrap(err, "failed to get target candidates")
	}
	if len(targetCandidates) == 0 {
		return MatcherResult{}, ErrTargetNotFound
	}

	// Find best pair: priority to big SCC, then same SCC, then closest (fallback)
	sourceCandidate, targetCandidate, found := matcher.findBestCandidatePair(sourceCandidates, targetCandidates)
	if !found {
		// Should not happen if we have candidates, but handle defensively
		return MatcherResult{}, errors.Wrapf(ErrCandidatesNotFound, "no routable candidate pair found for source %d and target %d", sourceCandidate.vertex, targetCandidate.vertex)
	}

	// Route between selected candidates
	ans, path := matcher.engine.queryPool.ShortestPath(sourceCandidate.vertex, targetCandidate.vertex)
	if ans == -1.0 {
		return MatcherResult{}, errors.Wrapf(ErrPathNotFound, "no path found between vertices %d and %d", sourceCandidate.vertex, targetCandidate.vertex)
	}
	if len(path) < 2 {
		return MatcherResult{}, errors.Wrapf(ErrSameVertex, "source and target vertices are the same: %d", sourceCandidate.vertex)
	}

	// Build result
	edges := []spatial.Edge{}
	subMatch := SubMatch{
		Observations: make([]ObservationResult, 2),
		Probability:  100.0,
	}

	intermediateEdges := []EdgeResult{}
	for i := 1; i < len(path); i++ {
		s := path[i-1]
		t := path[i]
		edge := matcher.engine.edges[s][t]
		edges = append(edges, *edge)
		edgeGeomCopy := make(s2.Polyline, len(*edge.Polyline))
		copy(edgeGeomCopy, *edge.Polyline)
		intermediateEdges = append(intermediateEdges, EdgeResult{
			Geom:   edgeGeomCopy,
			Weight: edge.Weight,
			ID:     edge.ID,
		})
	}

	subMatch.Observations[0] = ObservationResult{
		Observation: source,
		MatchedEdge: edges[0],
	}
	if len(intermediateEdges) > 1 {
		subMatch.Observations[0].NextEdges = intermediateEdges[1 : len(edges)-1]
	}

	subMatch.Observations[1] = ObservationResult{
		Observation: target,
		MatchedEdge: edges[len(edges)-1],
	}

	return MatcherResult{
		SubMatches: []SubMatch{subMatch},
	}, nil
}

// getCandidates retrieves candidate edges for a point and converts them to candidateInfo
func (matcher *MapMatcher) getCandidates(pt s2.Point, radiusMeters float64, limit int) ([]candidateInfo, error) {
	var nearestObjects []spatial.NearestObject
	var err error

	if radiusMeters < 0 {
		nearestObjects, err = matcher.engine.storage.FindNearest(pt, limit)
	} else {
		nearestObjects, err = matcher.engine.storage.FindNearestInRadius(pt, radiusMeters, limit)
	}
	if err != nil {
		return nil, err
	}

	candidates := make([]candidateInfo, 0, len(nearestObjects))
	for _, obj := range nearestObjects {
		edgeData := matcher.engine.storage.GetEdge(obj.EdgeID)
		if edgeData == nil {
			continue
		}

		m, n := edgeData.Source, edgeData.Target
		edge := matcher.engine.edges[m][n]
		if edge == nil {
			continue
		}

		// Determine which vertex to use based on projection fraction
		_, fraction, _ := spatial.CalcProjection(*edge.Polyline, pt)
		vertex := n
		if fraction > 0.5 {
			vertex = m
		}

		// Get SCC component for this vertex
		sccComponent, exists := matcher.engine.vertexStrongComponent[vertex]
		if !exists {
			sccComponent = -1
		}

		candidates = append(candidates, candidateInfo{
			edgeID:       obj.EdgeID,
			vertex:       vertex,
			sccComponent: sccComponent,
			distance:     obj.DistanceTo,
			edge:         edge,
		})
	}

	return candidates, nil
}

// findBestCandidatePair finds the best source-target pair with priority to non-tiny SCC.
// Work in the following order:
// 1 both candidates in the same non-tiny SCC (size >= SMALL_COMPONENT_SIZE)
// 2: both candidates in the same SCC (including small ones)
// 3: closest candidates regardless of SCC (fallback, routing may fail)
func (matcher *MapMatcher) findBestCandidatePair(sources, targets []candidateInfo) (candidateInfo, candidateInfo, bool) {
	if len(sources) == 0 || len(targets) == 0 {
		return candidateInfo{}, candidateInfo{}, false
	}

	var bestSource, bestTarget candidateInfo
	bestDistance := math.MaxFloat64
	found := false

	// Priority 1: find best pair where BOTH are in the same non-tiny SCC
	for _, src := range sources {
		if src.sccComponent == -1 {
			continue
		}
		if matcher.engine.isComponentVerySmall[src.sccComponent] {
			continue
		}
		for _, tgt := range targets {
			if tgt.sccComponent != src.sccComponent {
				continue
			}
			totalDist := src.distance + tgt.distance
			if totalDist < bestDistance {
				bestDistance = totalDist
				bestSource = src
				bestTarget = tgt
				found = true
			}
		}
	}

	if found {
		return bestSource, bestTarget, true
	}

	// Priority 2: find best pair in ANY same SCC (including very small ones)
	bestDistance = math.MaxFloat64
	for _, src := range sources {
		if src.sccComponent == -1 {
			continue
		}
		for _, tgt := range targets {
			if tgt.sccComponent == -1 {
				continue
			}
			if src.sccComponent != tgt.sccComponent {
				continue
			}
			totalDist := src.distance + tgt.distance
			if totalDist < bestDistance {
				bestDistance = totalDist
				bestSource = src
				bestTarget = tgt
				found = true
			}
		}
	}

	if found {
		return bestSource, bestTarget, true
	}

	// Priority 3: fallback - try routing pairs sorted by distance until one works
	// @todo: this is huge performance hit, but it is what it is.
	type candidatePair struct {
		src, tgt candidateInfo
		dist     float64
	}
	pairs := make([]candidatePair, 0, len(sources)*len(targets))
	for _, src := range sources {
		for _, tgt := range targets {
			pairs = append(pairs, candidatePair{src, tgt, src.distance + tgt.distance})
		}
	}
	// Sort by total distance
	for i := 0; i < len(pairs)-1; i++ {
		for j := i + 1; j < len(pairs); j++ {
			if pairs[j].dist < pairs[i].dist {
				pairs[i], pairs[j] = pairs[j], pairs[i]
			}
		}
	}
	// Try pairs in order until we find a routable one
	for _, p := range pairs {
		ans, _ := matcher.engine.queryPool.ShortestPath(p.src.vertex, p.tgt.vertex)
		if ans != -1.0 {
			return p.src, p.tgt, true
		}
	}

	return candidateInfo{}, candidateInfo{}, false
}
