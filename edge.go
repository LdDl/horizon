package horizon

import (
	"github.com/golang/geo/s2"
)

// Edge Representation of segment of road
/*
	ID - unique identifier
	Source - identifier of source vertex
	Target - identifier of target vertex
	Weight - cost of moving on edge (usually it is length or time)
	Polyline - geometry of edge, pointer to s2.Polyline (wrapper)
*/
type Edge struct {
	ID     int
	Source int
	Target int
	Weight float64
	*s2.Polyline
}
