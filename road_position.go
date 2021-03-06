package horizon

import (
	"fmt"

	"github.com/golang/geo/s2"
)

// RoadPositions Set of states
type RoadPositions []*RoadPosition

// RoadPosition Representation of state (in terms of Hidden Markov Model)
/*
	ID - unique identifier of state
	GraphEdge - pointer to closest edge in graph
	GraphVertex  - indentifier of closest vertex
	Projected - point (Observation) project onto edge, pointer to GeoPoint
*/
type RoadPosition struct {
	RoadPositionID int
	GraphEdge      *Edge
	GraphVertex    int64
	Projected      *GeoPoint
}

// NewRoadPositionFromLonLat Returns pointer to created State
/*
	stateID - unique identifier for state
	graphVertex - indentifier of vertex which is closest to Observation
	e - pointer to Edge
	lon - longitude (X for SRID = 0)
	lat - latitude (Y for SRID = 0)
	srid - SRID (see https://en.wikipedia.org/wiki/Spatial_reference_system)
*/
func NewRoadPositionFromLonLat(stateID int, graphVertex int64, e *Edge, lon, lat float64, srid ...int) *RoadPosition {
	state := RoadPosition{
		RoadPositionID: stateID,
		GraphEdge:      e,
		GraphVertex:    graphVertex,
	}
	if len(srid) != 0 {
		switch srid[0] {
		case 0:
			state.Projected = NewEuclideanPoint(lon, lat)
			break
		case 4326:
			state.Projected = NewWGS84Point(lon, lat)
			break
		default:
			state.Projected = NewWGS84Point(lon, lat)
			break
		}
	}
	return &state
}

// NewRoadPositionFromS2LatLng Returns pointer to created State
/*
	stateID - unique identifier for state
	graphVertex - indentifier of vertex which is closest to Observation
	e - pointer to Edge
	lon - longitude (X for SRID = 0)
	lat - latitude (Y for SRID = 0)
	srid - SRID (see https://en.wikipedia.org/wiki/Spatial_reference_system)
*/
func NewRoadPositionFromS2LatLng(stateID int, graphVertex int64, e *Edge, latLng *s2.LatLng, srid ...int) *RoadPosition {
	state := RoadPosition{
		RoadPositionID: stateID,
		GraphEdge:      e,
		GraphVertex:    graphVertex,
	}
	if len(srid) != 0 {
		switch srid[0] {
		case 0:
			state.Projected = NewEuclideanPoint(latLng.Lng.Degrees(), latLng.Lat.Degrees())
			break
		case 4326:
			state.Projected = NewWGS84Point(latLng.Lng.Degrees(), latLng.Lat.Degrees())
			break
		default:
			state.Projected = NewWGS84Point(latLng.Lng.Degrees(), latLng.Lat.Degrees())
			break
		}
	}
	return &state
}

// ID Method to fit interface State (see https://github.com/LdDl/viterbi/blob/master/viterbi.go#L9)
func (state RoadPosition) ID() int {
	return state.RoadPositionID
}

// String Pretty format for State
func (state RoadPosition) String() string {
	latlng := s2.LatLngFromPoint(state.Projected.Point)
	return fmt.Sprintf(
		"State is:\n\tSourceVertexID => %v\n\tTargetVertexID => %v\n\tSRID: %d\n\tCoords => %v",
		state.GraphEdge.Source, state.GraphEdge.Target, state.Projected.srid, latlng.String(),
	)
}
