package horizon

import (
	"fmt"

	"github.com/LdDl/horizon/spatial"
	"github.com/golang/geo/s2"
	"github.com/pkg/errors"
)

// FindShortestPath Find shortest path between two obserations (not necessary GPS points).
/*
	NOTICE: this function snaps point to nearest edges simply (without multiple 'candidates' for each observation)
	gpsMeasurements - Two observations
	statesRadiusMeters - maximum radius to search nearest polylines
*/
func (matcher *MapMatcher) FindShortestPath(source, target *GPSMeasurement, statesRadiusMeters float64) (MatcherResult, error) {
	var closestSource []spatial.NearestObject
	var err error
	if statesRadiusMeters < 0 {
		closestSource, err = matcher.engine.storage.FindNearest(source.Point, 1)
		if err != nil {
			return MatcherResult{}, errors.Wrapf(err, "FindNearest failed for source point %v", source.Point)
		}
	} else {
		closestSource, err = matcher.engine.storage.FindNearestInRadius(source.Point, statesRadiusMeters, 1)
		if err != nil {
			return MatcherResult{}, errors.Wrapf(err, "FindNearestInRadius failed for source point %v with radius %f", source.Point, statesRadiusMeters)
		}
	}
	// @todo need to handle error also
	if len(closestSource) == 0 {
		// @todo need to handle this case properly...
		return MatcherResult{}, ErrSourceNotFound
	}
	if len(closestSource) > 1 {
		// actually it's impossible if FindNearestInRadius() has been implemented correctly
		return MatcherResult{}, ErrSourceHasMoreEdges
	}

	var closestTarget []spatial.NearestObject
	if statesRadiusMeters < 0 {
		closestTarget, err = matcher.engine.storage.FindNearest(target.Point, 1)
		if err != nil {
			return MatcherResult{}, errors.Wrapf(err, "FindNearest failed for target point %v", target.Point)
		}
	} else {
		closestTarget, err = matcher.engine.storage.FindNearestInRadius(target.Point, statesRadiusMeters, 1)
		if err != nil {
			return MatcherResult{}, errors.Wrapf(err, "FindNearestInRadius failed for target point %v with radius %f", target.Point, statesRadiusMeters)
		}
	}
	if len(closestTarget) == 0 {
		// @todo need to handle this case properly...
		return MatcherResult{}, ErrTargetNotFound
	}
	if len(closestTarget) > 1 {
		// actually it's impossible if FindNearestInRadius() has been implemented correctly
		return MatcherResult{}, ErrTargetHasMoreEdges
	}

	s2polylineSource := matcher.engine.storage.GetEdge(closestSource[0].EdgeID)
	s2polylineTarget := matcher.engine.storage.GetEdge(closestTarget[0].EdgeID)

	// Find vertex for 'source' point
	m, n := s2polylineSource.Source, s2polylineSource.Target
	edgeSource := matcher.engine.edges[m][n]
	if edgeSource == nil {
		return MatcherResult{}, fmt.Errorf("Edge 'source' not found in graph for edgeID=%d", closestSource[0].EdgeID)
	}
	_, fractionSource, _ := spatial.CalcProjection(*edgeSource.Polyline, source.Point)
	choosenSourceVertex := n
	if fractionSource > 0.5 {
		choosenSourceVertex = m
	} else {
		choosenSourceVertex = n
	}

	// Find vertex for 'target' point
	m, n = s2polylineTarget.Source, s2polylineTarget.Target
	edgeTarget := matcher.engine.edges[m][n]
	if edgeTarget == nil {
		return MatcherResult{}, fmt.Errorf("Edge 'target' not found in graph for edgeID=%d", closestTarget[0].EdgeID)
	}
	_, fractionTarget, _ := spatial.CalcProjection(*edgeTarget.Polyline, target.Point)
	choosenTargetVertex := n
	if fractionTarget > 0.5 {
		choosenTargetVertex = m
	} else {
		choosenTargetVertex = n
	}

	// Check if source and target vertices are in the same connected component
	sourceComp, sourceExists := matcher.engine.vertexComponent[choosenSourceVertex]
	targetComp, targetExists := matcher.engine.vertexComponent[choosenTargetVertex]
	if sourceExists && targetExists && sourceComp != targetComp {
		return MatcherResult{}, errors.Wrapf(ErrDifferentComponents, "vertices %d (component %d) and %d (component %d) are in different components", choosenSourceVertex, sourceComp, choosenTargetVertex, targetComp)
	}

	ans, path := matcher.engine.queryPool.ShortestPath(choosenSourceVertex, choosenTargetVertex)
	if ans == -1.0 {
		return MatcherResult{}, errors.Wrapf(ErrPathNotFound, "no path found between vertices %d and %d", choosenSourceVertex, choosenTargetVertex)
	}
	if len(path) < 2 {
		return MatcherResult{}, errors.Wrapf(ErrSameVertex, "source and target vertices are the same: %d", choosenSourceVertex)
	}
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
