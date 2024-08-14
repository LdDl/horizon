package horizon

import (
	"github.com/LdDl/viterbi"
	"github.com/golang/geo/s2"
)

// ObservationResult Representation of gps measurement matched to G(v,e)
/*
	Observation - gps measurement itself
	MatchedEdge - edge in G(v,e) corresponding to current gps measurement
*/
type ObservationResult struct {
	Observation     *GPSMeasurement
	MatchedEdge     Edge
	MatchedVertex   Vertex
	ProjectedPoint  s2.Point
	MatchedVertexID int64
}

// MatcherResult Representation of map matching algorithm's output
/*
	Observations - set of ObservationResult
	Probability - probability got from Viterbi's algotithm
	Path - final path as s2.Polyline
	VerticesPath - IDs of graph vertices corresponding to traveled path
*/
type MatcherResult struct {
	Observations []ObservationResult
	Path         s2.Polyline
	VerticesPath []int64
	Probability  float64
}

// prepareResult Return MatcherResult for corresponding ViterbiPath, set of gps measurements and calculated routes' lengths
func (matcher *MapMatcher) prepareResult(vpath viterbi.ViterbiPath, gpsMeasurements GPSMeasurements, chRoutes map[int]map[int][]int64) MatcherResult {
	result := MatcherResult{
		Observations: make([]ObservationResult, len(gpsMeasurements)),
		Probability:  vpath.Probability,
		VerticesPath: []int64{},
	}

	rpPath := make(RoadPositions, len(vpath.Path))
	for i := range vpath.Path {
		rpPath[i] = vpath.Path[i].(*RoadPosition)
	}

	result.Observations[0] = ObservationResult{
		Observation:     gpsMeasurements[0],
		MatchedEdge:     *rpPath[0].GraphEdge,
		MatchedVertex:   *matcher.engine.vertices[rpPath[0].GraphVertex],
		ProjectedPoint:  rpPath[0].Projected.Point,
		MatchedVertexID: rpPath[0].GraphVertex,
	}
	result.VerticesPath = append(result.VerticesPath, rpPath[0].GraphEdge.Source, rpPath[0].GraphEdge.Target)
	// Cut first graph edge [next vertex to projected point : last_vertex]
	// And then prepend projected point to given slice
	// result.Path = append(result.Path, gpsMeasurements[0].Point, rpPath[0].Projected.Point)
	// result.Path = append(result.Path, *rpPath[0].GraphEdge.Polyline...)
	result.Path = append(result.Path, append(s2.Polyline{rpPath[0].Projected.Point}, (*rpPath[0].GraphEdge.Polyline)[rpPath[0].next:]...)...)
	// Iterate other states
	for i := 1; i < len(rpPath); i++ {
		previousState := rpPath[i-1]
		currentState := rpPath[i]
		result.Observations[i] = ObservationResult{
			Observation:     gpsMeasurements[i],
			MatchedEdge:     *currentState.GraphEdge,
			MatchedVertex:   *matcher.engine.vertices[currentState.GraphVertex],
			ProjectedPoint:  currentState.Projected.Point,
			MatchedVertexID: currentState.GraphVertex,
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
			result.Path = append(result.Path, (*edge.Polyline)...)
		}
		// if previousState.GraphEdge.ID == currentState.GraphEdge.ID {
		// 	result.Observations[i] = ObservationResult{
		// 		Observation:     gpsMeasurements[i],
		// 		MatchedEdge:     *previousState.GraphEdge,
		// 		ProjectedPoint:  previousState.Projected.Point,
		// 		MatchedVertexID: previousState.GraphVertex,
		// 	}
		// 	continue
		// }
		// path := chRoutes[previousState.RoadPositionID][currentState.RoadPositionID]
		// for e := 1; e < len(path); e++ {
		// 	sourceVertex := path[e-1]
		// 	targetVertex := path[e]
		// 	edge := matcher.engine.edges[sourceVertex][targetVertex]
		// 	if len(*edge.Polyline) < 2 {
		// 		fmt.Printf("[WARNING]: Edge %d have less than 2 points\n", edge.ID)
		// 	}
		// 	result.Path = append(result.Path, (*edge.Polyline)[1:]...)
		// 	// result.Path = append(result.Path, *edge.Polyline...) // Cause duplicates in geometry
		// 	result.VerticesPath = append(result.VerticesPath, targetVertex)
		// 	if e == len(path)-1 {
		// 		result.Observations[i] = ObservationResult{
		// 			Observation: gpsMeasurements[i],
		// 			MatchedEdge: *edge,
		// 		}
		// 	}
		// }
	}
	// result.Path = append(result.Path, rpPath[len(rpPath)-1].Projected.Point, gpsMeasurements[len(gpsMeasurements)-1].Point)

	// for i := 1; i < len(rpPath); i++ {
	// 	previousState := rpPath[i-1]
	// 	currentState := rpPath[i]
	// 	// fmt.Println("path len", i, len(chRoutes[previousState.RoadPositionID][currentState.RoadPositionID]))
	// 	if previousState.GraphEdge.ID == currentState.GraphEdge.ID {
	// 		result.Observations[i] = ObservationResult{
	// 			Observation:     gpsMeasurements[i],
	// 			MatchedEdge:     *previousState.GraphEdge,
	// 			ProjectedPoint:  previousState.Projected.Point,
	// 			MatchedVertexID: previousState.GraphVertex,
	// 		}
	// 		continue
	// 	}
	// 	path := chRoutes[previousState.RoadPositionID][currentState.RoadPositionID]
	// 	for e := 1; e < len(path); e++ {
	// 		sourceVertex := path[e-1]
	// 		targetVertex := path[e]
	// 		edge := matcher.engine.edges[sourceVertex][targetVertex]
	// 		if len(*edge.Polyline) < 2 {
	// 			fmt.Printf("[WARNING]: Edge %d have less than 2 points\n", edge.ID)
	// 		}
	// 		result.Path = append(result.Path, (*edge.Polyline)[1:]...)
	// 		// result.Path = append(result.Path, *edge.Polyline...) // Cause duplicates in geometry
	// 		result.VerticesPath = append(result.VerticesPath, targetVertex)
	// 		if e == len(path)-1 {
	// 			result.Observations[i] = ObservationResult{
	// 				Observation: gpsMeasurements[i],
	// 				MatchedEdge: *edge,
	// 			}
	// 		}
	// 	}
	// }
	// Cut whole geometry [first vertex : previous vertex to projected point]
	// And then append projected point to given slice

	// @todo: I believe there is a better way to handle this case since projected point has been calculated in Run() function
	// _, _, next := calcProjection(result.Path, rpPath[len(rpPath)-1].Projected.Point)
	// result.Path = append(result.Path[:(next-1)], rpPath[len(rpPath)-1].Projected.Point)
	return result
}
