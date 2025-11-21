package spatial

import (
	"github.com/golang/geo/s2"
)

// Edge Representation of segment of road (edge in graph)
/*
	ID - unique identifier
	Source - identifier of source vertex
	Target - identifier of target vertex
	Weight - cost of moving on edge (usually it is length or time)
	Polyline - geometry of edge, pointer to s2.Polyline (wrapper)
*/
type Edge struct {
	*s2.Polyline
	Weight float64
	ID     int64
	Source int64
	Target int64
}
