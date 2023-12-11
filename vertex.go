package horizon

import (
	"github.com/golang/geo/s2"
)

// Vertex Representation of node on a road (vertex in graph)
/*
	ID - unique identifier (user defined, should be contained in parent graph)
	Point - geometry of vertex, pointer to s2.Point (wrapper)
*/
type Vertex struct {
	*s2.Point
	ID int64
}
