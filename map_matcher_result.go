package horizon

import (
	"fmt"

	"github.com/LdDl/horizon/spatial"
	"github.com/LdDl/viterbi"
	"github.com/golang/geo/s2"
)

// ObservationResult Representation of gps measurement matched to G(v,e)
/*
	Observation - gps measurement itself
	IsMatched - true if observation was successfully matched to a road, false if no candidates were found
	MatchedEdge - edge in G(v,e) corresponding to current gps measurement (empty if IsMatched is false)
	MatchedVertex - stands for closest vertex to the observation (empty if IsMatched is false)
	ProjectedPoint - projection onto the matched edge (empty if IsMatched is false)
	ProjectedPointIdx - index of the point in polyline which follows projection point
	NextEdges - set of leading edges up to next observation. Could be an empty array if observations are very close to each other or if it just last observation
*/
type ObservationResult struct {
	Observation        *GPSMeasurement
	IsMatched          bool
	MatchedEdge        spatial.Edge
	MatchedVertex      spatial.Vertex
	ProjectedPoint     s2.Point
	ProjectionPointIdx int
	NextEdges          []EdgeResult
}

type EdgeResult struct {
	Geom   s2.Polyline
	Weight float64
	ID     int64
}

// SubMatch Representation of a single continuous matched segment
/*
	Observations - set of ObservationResult for this segment
	Probability - probability got from Viterbi's algorithm for this segment
*/
type SubMatch struct {
	Observations []ObservationResult
	Probability  float64
}

// MatcherResult Representation of map matching algorithm's output
/*
	SubMatches - set of SubMatch segments (split when route cannot be computed between consecutive points)
*/
type MatcherResult struct {
	SubMatches []SubMatch
}

// prepareSubMatch returns SubMatch for corresponding ViterbiPath, set of gps measurements and calculated routes' lengths
func (matcher *MapMatcher) prepareSubMatch(vpath viterbi.ViterbiPath, gpsMeasurements GPSMeasurements, layers []RoadPositions, chRoutes map[int]map[int][]int64) SubMatch {
	subMatch := SubMatch{
		Observations: make([]ObservationResult, len(gpsMeasurements)),
		Probability:  vpath.Probability,
	}

	rpPath := make(RoadPositions, len(vpath.Path))
	for i := range vpath.Path {
		rpPath[i] = vpath.Path[i].(*RoadPosition)
	}

	subMatch.Observations[0] = ObservationResult{
		Observation:        gpsMeasurements[0],
		IsMatched:          true,
		MatchedEdge:        *rpPath[0].GraphEdge,
		MatchedVertex:      *matcher.engine.vertices[rpPath[0].PickedGraphVertex],
		ProjectedPoint:     rpPath[0].Projected.Point,
		ProjectionPointIdx: rpPath[0].next,
	}

	// Iterate other states
	lastEdgeID := int64(-1)
	for i := 1; i < len(rpPath); i++ {
		previousState := rpPath[i-1]
		currentState := rpPath[i]
		subMatch.Observations[i] = ObservationResult{
			Observation:        gpsMeasurements[i],
			IsMatched:          true,
			MatchedEdge:        *currentState.GraphEdge,
			MatchedVertex:      *matcher.engine.vertices[currentState.PickedGraphVertex],
			ProjectedPoint:     currentState.Projected.Point,
			ProjectionPointIdx: currentState.next,
		}
		if previousState.GraphEdge.ID == currentState.GraphEdge.ID {
			continue
		}
		path := chRoutes[previousState.RoadPositionID][currentState.RoadPositionID]
		if len(path) < 2 {
			continue
		}
		for j := 1; j < len(path); j++ {
			sourceVertex := path[j-1]
			targetVertex := path[j]
			edge := matcher.engine.edges[sourceVertex][targetVertex]
			if len(*edge.Polyline) < 2 {
				fmt.Printf("[WARNING]: Edge %d have less than 2 points\n", edge.ID)
			}
			if i == len(rpPath)-1 && j == len(path)-1 {
				continue
			}
			lastEdgeID = edge.ID
			edgeGeomCopy := make(s2.Polyline, len(*edge.Polyline))
			copy(edgeGeomCopy, *edge.Polyline)
			subMatch.Observations[i-1].NextEdges = append(subMatch.Observations[i-1].NextEdges, EdgeResult{
				Geom:   edgeGeomCopy,
				Weight: edge.Weight,
				ID:     edge.ID,
			})
		}
	}
	if rpPath[len(rpPath)-1].GraphEdge.ID == lastEdgeID {
		// @todo:
		// Last edge is the same as matched
	}

	return subMatch
}
