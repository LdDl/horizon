package spatial

import (
	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
)

// Edge Representation of segment of road (edge in graph)
/*
	ID - unique identifier
	Source - identifier of source vertex
	Target - identifier of target vertex
	Weight - cost of moving on edge (usually it is length or time)
	Polyline - geometry of edge, pointer to s2.Polyline (wrapper)
	CumSegLen - cumulative spherical segment distances (radians), precomputed; len == len(*Polyline)-1
	BoundCenter/BoundRadius - precomputed spherical bounding cap for cheap pruning in spatial queries
*/
type Edge struct {
	*s2.Polyline
	CumSegLen    []float64
	BoundCenter  s2.Point
	BoundRadius  s1.Angle
	Weight       float64
	LengthMeters float64
	ID           int64
	Source       int64
	Target       int64
}

// PrecomputeCumLen fills the CumSegLen cache using spherical segment distances.
// Safe to call once at edge construction time; subsequent calls overwrite the cache.
// Caller must ensure Polyline has at least 2 points.
func (e *Edge) PrecomputeCumLen() {
	if e.Polyline == nil {
		return
	}
	pl := *e.Polyline
	if len(pl) < 2 {
		return
	}
	cum := make([]float64, len(pl)-1)
	var accum float64
	for i := 0; i < len(pl)-1; i++ {
		accum += float64(pl[i].Distance(pl[i+1]))
		cum[i] = accum
	}
	e.CumSegLen = cum
}

// PrecomputeBound fills BoundCenter/BoundRadius with the polyline's spherical
// cap bound. Used to cheaply prune spatial queries without scanning polyline segments.
func (e *Edge) PrecomputeBound() {
	if e.Polyline == nil {
		return
	}
	pl := *e.Polyline
	if len(pl) == 0 {
		return
	}
	cap := e.Polyline.CapBound()
	e.BoundCenter = cap.Center()
	e.BoundRadius = cap.Radius()
}
