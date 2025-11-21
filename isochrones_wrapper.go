package horizon

import (
	"fmt"
	"log"

	"github.com/LdDl/horizon/spatial"
	"github.com/pkg/errors"
)

// IsochronesResult Representation of isochrones algorithm's output
/*
 */
type IsochronesResult []*Isochrone

type Isochrone struct {
	Vertex *spatial.Vertex
	Cost   float64
}

// FindIsochrones Find shortest path between two obserations (not necessary GPS points).
/*
	NOTICE: this function snaps point to only one nearest vertex (without multiple candidates for provided point)
	source - source for outcoming isochrones
	maxCost - max cost restriction for single isochrone line
	maxNearestRadius - max radius of search for nearest vertex
*/
func (matcher *MapMatcher) FindIsochrones(source *GPSMeasurement, maxCost float64, maxNearestRadius float64) (IsochronesResult, error) {
	closestSource, _ := matcher.engine.s2Storage.NearestNeighborsInRadius(source.Point, maxNearestRadius, 1)
	if len(closestSource) == 0 {
		// @todo need to handle this case properly...
		return nil, ErrSourceNotFound
	}
	// Find corresponding edge
	s2polylineSource := matcher.engine.s2Storage.GetEdge(closestSource[0].EdgeID)
	// Find vertex for 'source' point
	m, n := s2polylineSource.Source, s2polylineSource.Target
	edgeSource := matcher.engine.edges[m][n]
	if edgeSource == nil {
		return nil, fmt.Errorf("Edge 'source' not found in graph")
	}
	_, fractionSource, _ := spatial.CalcProjection(*edgeSource.Polyline, source.Point)
	choosenSourceVertex := n
	if fractionSource > 0.5 {
		choosenSourceVertex = m
	} else {
		choosenSourceVertex = n
	}

	ans, err := matcher.engine.graph.Isochrones(choosenSourceVertex, maxCost)
	if err != nil {
		return nil, errors.Wrapf(err, "Can't call isochrones for vertex with id '%d'", choosenSourceVertex)
	}
	isochrones := make(IsochronesResult, 0, len(ans))
	for vertexID, cost := range ans {
		vertex, ok := matcher.engine.vertices[vertexID]
		if !ok {
			log.Printf("[WARNING]; No such vertex in storage: %d\n", vertexID)
		}
		isochrones = append(isochrones, &Isochrone{
			Vertex: vertex,
			Cost:   cost,
		})
	}
	return isochrones, nil
}
