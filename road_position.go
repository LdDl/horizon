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
	beforeProjection - distance from starting point to projected one
	afterProjection - distance from projected point to last one
	next - index of the next vertex in s2.Polyline after the projected point
*/
type RoadPosition struct {
	Projected          *GeoPoint
	GraphEdge          *Edge
	beforeProjection   float64
	afterProjection    float64
	PickedGraphVertex  int64
	RoutingGraphVertex int64
	RoadPositionID     int
	next               int
}

// NewRoadPositionFromLonLat Returns pointer to created State
/*
	stateID - unique identifier for state
	pickedGraphVertex - indentifier of vertex which is closest to Observation
	routingGraphVertex - indentifier of vertex which will be used in routing (initially it should match pickedGraphVertex in most cases)
	e - pointer to Edge
	lon - longitude (X for SRID = 0)
	lat - latitude (Y for SRID = 0)
	srid - SRID (see https://en.wikipedia.org/wiki/Spatial_reference_system)
*/
func NewRoadPositionFromLonLat(stateID int, pickedGraphVertex, routingGraphVertex int64, e *Edge, lon, lat float64, srid ...int) *RoadPosition {
	state := RoadPosition{
		RoadPositionID:     stateID,
		GraphEdge:          e,
		PickedGraphVertex:  pickedGraphVertex,
		RoutingGraphVertex: routingGraphVertex,
	}
	if len(srid) != 0 {
		switch srid[0] {
		case 0:
			state.Projected = NewEuclideanPoint(lon, lat)
		case 4326:
			state.Projected = NewWGS84Point(lon, lat)
		default:
			state.Projected = NewWGS84Point(lon, lat)
		}
	}
	return &state
}

// NewRoadPositionFromS2LatLng Returns pointer to created State
/*
	stateID - unique identifier for state
	pickedGraphVertex - indentifier of vertex which is closest to Observation
	routingGraphVertex - indentifier of vertex which will be used in routing (initially it should match pickedGraphVertex in most cases)
	e - pointer to Edge
	lon - longitude (X for SRID = 0)
	lat - latitude (Y for SRID = 0)
	srid - SRID (see https://en.wikipedia.org/wiki/Spatial_reference_system)
*/
func NewRoadPositionFromS2LatLng(stateID int, pickedGraphVertex, routingGraphVertex int64, e *Edge, latLng *s2.LatLng, srid ...int) *RoadPosition {
	state := RoadPosition{
		RoadPositionID:     stateID,
		GraphEdge:          e,
		PickedGraphVertex:  pickedGraphVertex,
		RoutingGraphVertex: routingGraphVertex,
	}
	if len(srid) != 0 {
		switch srid[0] {
		case 0:
			state.Projected = NewEuclideanPoint(latLng.Lng.Degrees(), latLng.Lat.Degrees())
		case 4326:
			state.Projected = NewWGS84Point(latLng.Lng.Degrees(), latLng.Lat.Degrees())
		default:
			state.Projected = NewWGS84Point(latLng.Lng.Degrees(), latLng.Lat.Degrees())
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
