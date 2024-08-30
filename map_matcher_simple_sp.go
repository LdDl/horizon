package horizon

import (
	"fmt"

	"github.com/golang/geo/s2"
)

// FindShortestPath Find shortest path between two obserations (not necessary GPS points).
/*
	NOTICE: this function snaps point to nearest edges simply (without multiple 'candidates' for each observation)
	gpsMeasurements - Two observations
	statesRadiusMeters - maximum radius to search nearest polylines
*/
func (matcher *MapMatcher) FindShortestPath(source, target *GPSMeasurement, statesRadiusMeters float64) (MatcherResult, error) {
	closestSource, _ := matcher.engine.s2Storage.NearestNeighborsInRadius(source.Point, statesRadiusMeters, 1)
	if len(closestSource) == 0 {
		// @todo need to handle this case properly...
		return MatcherResult{}, ErrSourceNotFound
	}
	if len(closestSource) > 1 {
		// actually it's impossible if NearestNeighborsInRadius() has been implemented correctly
		return MatcherResult{}, ErrSourceHasMoreEdges
	}

	closestTarget, _ := matcher.engine.s2Storage.NearestNeighborsInRadius(target.Point, statesRadiusMeters, 1)
	if len(closestTarget) == 0 {
		// @todo need to handle this case properly...
		return MatcherResult{}, ErrTargetNotFound
	}
	if len(closestTarget) > 1 {
		// actually it's impossible if NearestNeighborsInRadius() has been implemented correctly
		return MatcherResult{}, ErrTargetHasMoreEdges
	}

	s2polylineSource := matcher.engine.s2Storage.edges[closestSource[0].edgeID]
	s2polylineTarget := matcher.engine.s2Storage.edges[closestTarget[0].edgeID]

	// Find vertex for 'source' point
	m, n := s2polylineSource.Source, s2polylineSource.Target
	edgeSource := matcher.engine.edges[m][n]
	if edgeSource == nil {
		return MatcherResult{}, fmt.Errorf("Edge 'source' not found in graph")
	}
	_, fractionSource, _ := calcProjection(*edgeSource.Polyline, source.Point)
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
		return MatcherResult{}, fmt.Errorf("Edge 'target' not found in graph")
	}
	_, fractionTarget, _ := calcProjection(*edgeTarget.Polyline, target.Point)
	choosenTargetVertex := n
	if fractionTarget > 0.5 {
		choosenTargetVertex = m
	} else {
		choosenTargetVertex = n
	}

	ans, path := matcher.engine.graph.ShortestPath(choosenSourceVertex, choosenTargetVertex)
	if ans == -1.0 {
		return MatcherResult{}, ErrPathNotFound
	}
	if len(path) < 2 {
		return MatcherResult{}, ErrSameVertex
	}
	edges := []Edge{}
	result := MatcherResult{
		Observations: make([]ObservationResult, 2),
		Probability:  100.0,
	}
	for i := 1; i < len(path); i++ {
		s := path[i-1]
		t := path[i]
		edge := matcher.engine.edges[s][t]
		edges = append(edges, *edge)
		edgeGeomCopy := make(s2.Polyline, len(*edge.Polyline))
		copy(edgeGeomCopy, *edge.Polyline)
		result.Observations[i].NextEdges = append(result.Observations[i].NextEdges, EdgeResult{
			Geom:   edgeGeomCopy,
			Weight: edge.Weight,
			ID:     edge.ID,
		})
		// @todo
		// result.Path = append(result.Path, *edge.Polyline...)
	}

	result.Observations[0] = ObservationResult{
		Observation: source,
		MatchedEdge: edges[0],
	}

	result.Observations[1] = ObservationResult{
		Observation: target,
		MatchedEdge: edges[len(edges)-1],
	}

	return result, nil
}
