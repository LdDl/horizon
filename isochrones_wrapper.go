package horizon

import "fmt"

// IsochronesResult Representation of isochrones algorithm's output
/*
 */
type IsochronesResult struct {
}

// FindIsochrones Find shortest path between two obserations (not necessary GPS points).
/*
	NOTICE: this function snaps point to only one nearest vertex (without multiple candidates for provided point)
	source - source for outcoming isochrones
	maxCost - max cost restriction for single isochrone line
	maxNearestRadius - max radius of search for nearest vertex
*/
func (matcher *MapMatcher) FindIsochrones(source *GPSMeasurement, maxCost float64, maxNearestRadius float64) (*IsochronesResult, error) {
	closestSource, _ := matcher.engine.s2Storage.NearestNeighborsInRadius(source.Point, maxNearestRadius, 1)
	if len(closestSource) == 0 {
		// @todo need to handle this case properly...
		return &IsochronesResult{}, ErrSourceNotFound
	}
	// Find corresponding edge
	s2polylineSource := matcher.engine.s2Storage.edges[closestSource[0].edgeID]
	// Find vertex for 'source' point
	m, n := s2polylineSource.Source, s2polylineSource.Target
	edgeSource := matcher.engine.edges[m][n]
	if edgeSource == nil {
		return &IsochronesResult{}, fmt.Errorf("Edge 'source' not found in graph")
	}
	_, fractionSource := calcProjection(*edgeSource.Polyline, source.Point)
	choosenSourceVertex := n
	if fractionSource > 0.5 {
		choosenSourceVertex = m
	} else {
		choosenSourceVertex = n
	}
	_ = choosenSourceVertex
	return nil, nil
}
