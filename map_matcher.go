package horizon

import (
	"fmt"
	"math"

	"github.com/LdDl/viterbi"
	"github.com/golang/geo/s2"
	"github.com/pkg/errors"
)

const (
	ViterbiDebug           = false
	ROUTE_LENGTH_THRESHOLD = 9_999_999_999.0
)

// MapMatcher Engine for solving map matching problem
/*
	hmmParams - parameters of Hidden Markov Model
	engine - wrapper around MapEngine (for KNN and finding shortest path problems)
*/
type MapMatcher struct {
	hmmParams *HmmProbabilities
	engine    *MapEngine
}

// NewMapMatcherDefault Returns pointer to created MapMatcher with default parameters
func NewMapMatcherDefault() *MapMatcher {
	return &MapMatcher{
		hmmParams: HmmProbabilitiesDefault(),
	}
}

// NewMapMatcher Returns pointer to created MapMatcher with provided parameters
/*
	props - parameters of Hidden Markov Model
*/
func NewMapMatcher(props *HmmProbabilities, edgesFilename string) (*MapMatcher, error) {
	mm := &MapMatcher{
		hmmParams: props,
	}
	mapEngine, err := prepareEngine(edgesFilename)
	if err != nil {
		return nil, err
	}
	mm.engine = mapEngine
	return mm, nil
}

// Run Do magic
/*
	gpsMeasurements - Observations
	statesRadiusMeters - maximum radius to search nearest polylines
	maxStates - maximum of corresponding states
*/
func (matcher *MapMatcher) Run(gpsMeasurements []*GPSMeasurement, statesRadiusMeters float64, maxStates int) (MatcherResult, error) {
	if len(gpsMeasurements) < 3 {
		return MatcherResult{}, ErrMinumimGPSMeasurements
	}

	stateID := 0
	layers := []RoadPositions{}

	engineGpsMeasurements := []*GPSMeasurement{}
	closestSets := [][]NearestObject{}

	for i := 0; i < len(gpsMeasurements); i++ {
		closest, err := matcher.engine.s2Storage.NearestNeighborsInRadius(gpsMeasurements[i].Point, statesRadiusMeters, maxStates)
		if err != nil {
			return MatcherResult{}, errors.Wrapf(err, "Can't find neighbors for point: '%s' (states radius = %f, max states = %d)", gpsMeasurements[i].Point, statesRadiusMeters, maxStates)
		}
		if len(closest) == 0 {
			fmt.Printf("No candidates for %+v at pos %d. Radius: %f. Max sates: %d\n", gpsMeasurements[i].Point, i, statesRadiusMeters, maxStates)
			// @todo need to handle this case properly...
			continue
		}
		engineGpsMeasurements = append(engineGpsMeasurements, gpsMeasurements[i])
		closestSets = append(closestSets, closest)
	}

	if len(engineGpsMeasurements) == 0 {
		return MatcherResult{}, ErrCandidatesNotFound
	}

	obsState := make([]*CandidateLayer, len(engineGpsMeasurements))
	for i := 0; i < len(engineGpsMeasurements); i++ {
		s2point := engineGpsMeasurements[i].Point
		closest := closestSets[i]
		localStates := make(RoadPositions, len(closest))
		for j := range closest {
			s2polyline := matcher.engine.s2Storage.edges[closest[j].edgeID]
			m := s2polyline.Source
			n := s2polyline.Target
			edge := matcher.engine.edges[m][n]
			proj, fraction, next := calcProjection(*edge.Polyline, s2point)
			latLng := s2.LatLngFromPoint(proj)
			pickedGraphVertex := m
			routingGraphVertex := m
			if fraction > 0.5 {
				pickedGraphVertex = n
			} else {
				pickedGraphVertex = m
			}
			// For first candidate layer we should start routing from edge's target vertex
			if i == 0 {
				routingGraphVertex = n
			}
			roadPos := NewRoadPositionFromLonLat(stateID, pickedGraphVertex, routingGraphVertex, edge, latLng.Lng.Degrees(), latLng.Lat.Degrees(), 4326)
			roadPos.beforeProjection = edge.Weight * fraction
			roadPos.afterProjection = edge.Weight * (1 - fraction)
			roadPos.next = next
			localStates[j] = roadPos
			stateID++
		}
		layers = append(layers, localStates)
		obsState[i] = NewCandidateLayer(engineGpsMeasurements[i], localStates)
	}
	chRoutes := make(map[int]map[int][]int64)

	routeLengths := make(lengths)

	// @todo: Consider to use ShortestPathOneToMany (need to deal with the order of writing data to to chRoutes and routeLengths)
	for i := 1; i < len(layers); i++ {
		prevStates := layers[i-1]
		currentStates := layers[i]
		for m := range prevStates {
			if _, ok := chRoutes[prevStates[m].RoadPositionID]; !ok {
				chRoutes[prevStates[m].RoadPositionID] = make(map[int][]int64)
			}
			for n := range currentStates {
				if prevStates[m].RoutingGraphVertex == currentStates[n].RoutingGraphVertex {
					if prevStates[m].GraphEdge.ID == currentStates[n].GraphEdge.ID {
						ans := prevStates[m].Projected.DistanceTo(currentStates[n].Projected)
						chRoutes[prevStates[m].RoadPositionID][currentStates[n].RoadPositionID] = []int64{prevStates[m].GraphEdge.Source, prevStates[m].GraphEdge.Target}
						routeLengths.AddRouteLength(prevStates[m], currentStates[n], ans)
					} else {
						// We should jump to source vertex of current state, since edges are not the same
						ans, path := matcher.engine.graph.ShortestPath(prevStates[m].RoutingGraphVertex, currentStates[n].GraphEdge.Source)
						if ans < 0 {
							ans = math.MaxFloat64
						} else {
							// We should increase travel cost by last edge weight and put last edge's target vertex to the path
							ans += currentStates[n].GraphEdge.Weight
							path = append(path, currentStates[n].GraphEdge.Target)
						}
						chRoutes[prevStates[m].RoadPositionID][currentStates[n].RoadPositionID] = path
						routeLengths.AddRouteLength(prevStates[m], currentStates[n], ans)
					}
					continue
				}
				ans, path := matcher.engine.graph.ShortestPath(prevStates[m].RoutingGraphVertex, currentStates[n].RoutingGraphVertex)
				// ans, path := matcher.engine.graph.ShortestPath(prevStates[m].GraphVertex, currentStates[n].GraphEdge.Source)
				if ans < 0 {
					ans = math.MaxFloat64
				} else {
					// We should increase travel cost by last edge weight and put last edge's target vertex to the path
					ans += currentStates[n].GraphEdge.Weight
					path = append(path, currentStates[n].GraphEdge.Target)
					// Since we are doing Edge(target)-Edge(target) Dijkstra's call most of time we could:
					// 1) add penalty for source edge by adding remaining distance to target vertex of source edge
					// 2) add advantage for target edge by subtracting remaining distance to target vertex of target edge
					// @todo: this could lead to negative values. Need to investigate when it happens
					// ans = (ans + prevStates[m].afterProjection) - currentStates[n].afterProjection
				}
				chRoutes[prevStates[m].RoadPositionID][currentStates[n].RoadPositionID] = path
				routeLengths.AddRouteLength(prevStates[m], currentStates[n], ans)
			}
		}
		// We can skip chaning routing vertices in very last candidates layer
		if i == len(layers)-1 {
			continue
		}
		// After we've built routes between Prev->Current layers we can change source routing vertex to edge's target vertex
		// Let's demonstrate how it should work:
		// In the very first pair of previous and current candidates layers we should search path from edge's target vertex from previous layer to edge's source vertex from current layer: PrevLayer.Edge.Target -> CurrentLayer.Edge.Source
		// For all other pairs we change search vertex of current layer to edge's target vertex: PrevLayer.Edge.Target -> CurrentLayer.Edge.Target
		// It gives us a better handling for cases when a single vertex is indecent to multiple edges (which could lead to mismatch between shortest path edges and actually matched edge for the given candidate)
		for n := range currentStates {
			currentStates[n].RoutingGraphVertex = currentStates[n].GraphEdge.Target
		}
	}

	v, err := matcher.PrepareViterbi(obsState, routeLengths, engineGpsMeasurements)
	if err != nil {
		return MatcherResult{}, err
	}

	vpath := v.EvalPathLogProbabilities()
	if ViterbiDebug {
		fmt.Println("prob:", vpath.Probability)
		fmt.Println("path:")
		for i := range vpath.Path {
			fmt.Println("\t", vpath.Path[i].(*RoadPosition).GraphEdge.ID, vpath.Path[i].(*RoadPosition).ID())
		}
	}

	if len(vpath.Path) != len(engineGpsMeasurements) {
		return MatcherResult{}, fmt.Errorf("number of states in final path != number (%d and %d) of observations. Should be unreachable error", len(vpath.Path), len(engineGpsMeasurements))
	}

	result := matcher.prepareResult(vpath, engineGpsMeasurements, chRoutes)

	return result, nil
}

// PrepareViterbi Prepares engine for doing Viterbi's algorithm (see https://github.com/LdDl/viterbi/blob/master/viterbi.go#L25)
/*
	states - set of States
	gpsMeasurements - set of Observations
*/
func (matcher *MapMatcher) PrepareViterbi(obsStates []*CandidateLayer, routeLengths map[int]map[int]float64, gpsMeasurements []*GPSMeasurement) (*viterbi.Viterbi, error) {
	v := &viterbi.Viterbi{}

	statesIndx := make(map[int]int)
	idx := 0
	for i := range obsStates {
		for j := range obsStates[i].States {
			v.AddState(obsStates[i].States[j])
			statesIndx[obsStates[i].States[j].ID()] = idx
			if ViterbiDebug {
				fmt.Printf(`CustomState{Name: "%d", id: %d}%s`, obsStates[i].States[j].GraphEdge.ID, obsStates[i].States[j].ID(), ",\n")
			}
			idx++
		}
		if ViterbiDebug {
			fmt.Println()
		}
	}
	if ViterbiDebug {
		fmt.Println()
	}
	for i := range gpsMeasurements {
		if ViterbiDebug {
			fmt.Printf(`CustomObservation{Name: "%s", id: %d}%s`, obsStates[i].Observation.dateTime.Format("2006-01-02T15:04:05"), i, ",\n")
		}
		v.AddObservation(gpsMeasurements[i])
	}
	if ViterbiDebug {
		fmt.Println()
	}
	layers := make([]*CandidateLayer, len(gpsMeasurements))
	prevLayer := &CandidateLayer{}

	// I guess this is ugly.
	// @todo Refactor data prerapartion for Viterbi's algorithm
	for i := range gpsMeasurements {
		currentLayer := obsStates[i]
		matcher.computeEmissionLogProbabilities(currentLayer)
		// @experimental
		// currentLayer.EmissionLogProbabilities = softmaxEmissions(currentLayer.EmissionLogProbabilities)
		if i == 0 {
			for j := range currentLayer.EmissionLogProbabilities {
				if ViterbiDebug {
					fmt.Printf(`v.PutStartProbability(incStates[%d], %.15f) // Graph edge: %d. Graph vertex: %d%s`, statesIndx[currentLayer.EmissionLogProbabilities[j].rp.ID()], currentLayer.EmissionLogProbabilities[j].prob, currentLayer.EmissionLogProbabilities[j].rp.GraphEdge.ID, currentLayer.EmissionLogProbabilities[j].rp.RoutingGraphVertex, "\n")
				}
				v.PutStartProbability(currentLayer.EmissionLogProbabilities[j].rp, currentLayer.EmissionLogProbabilities[j].prob)
			}
			if ViterbiDebug {
				fmt.Println()
			}
		} else {
			err := matcher.computeTransitionLogProbabilities(prevLayer, currentLayer, routeLengths)
			if err != nil {
				return nil, err
			}
		}
		for j := range currentLayer.EmissionLogProbabilities {
			if ViterbiDebug {
				fmt.Printf(`v.PutEmissionProbability(incStates[%d], observations[%d], %.15f) // Graph edge: %d. Graph vertex: %d%s`, statesIndx[currentLayer.EmissionLogProbabilities[j].rp.ID()], i, currentLayer.EmissionLogProbabilities[j].prob, currentLayer.EmissionLogProbabilities[j].rp.GraphEdge.ID, currentLayer.EmissionLogProbabilities[j].rp.RoutingGraphVertex, "\n")
			}
			v.PutEmissionProbability(currentLayer.EmissionLogProbabilities[j].rp, gpsMeasurements[i], currentLayer.EmissionLogProbabilities[j].prob)
		}
		prevLayer = currentLayer
		layers[i] = currentLayer
		if ViterbiDebug {
			fmt.Println()
		}
	}

	for s := range layers {
		step := layers[s]
		for i := range step.TransitionLogProbabilities {
			if ViterbiDebug {
				fmt.Printf(`v.PutTransitionProbability(incStates[%d], incStates[%d], %.15f) // From graph edge %d (vertex %d) to graph edge %d (vertex %d)%s`, statesIndx[step.TransitionLogProbabilities[i].from.ID()], statesIndx[step.TransitionLogProbabilities[i].to.ID()], step.TransitionLogProbabilities[i].prob,
					step.TransitionLogProbabilities[i].from.GraphEdge.ID, step.TransitionLogProbabilities[i].from.RoutingGraphVertex,
					step.TransitionLogProbabilities[i].to.GraphEdge.ID, step.TransitionLogProbabilities[i].to.RoutingGraphVertex,
					"\n")
			}
			v.PutTransitionProbability(step.TransitionLogProbabilities[i].from, step.TransitionLogProbabilities[i].to, step.TransitionLogProbabilities[i].prob)
		}
		if ViterbiDebug {
			fmt.Println()
		}
	}

	return v, nil
}

// computeEmissionLogProbabilities Computes emission probablities between Observation and corresponding States
/*
	layer - wrapper of Observation
*/
func (matcher *MapMatcher) computeEmissionLogProbabilities(layer *CandidateLayer) {
	ems := make([]float64, len(layer.States))
	for i := range layer.States {
		distance := layer.States[i].Projected.DistanceTo(layer.Observation.GeoPoint)
		ems[i] = matcher.hmmParams.EmissionLogProbability(distance)
		layer.AddEmissionProbability(layer.States[i], matcher.hmmParams.EmissionLogProbability(distance))
	}
}

// computeTransitionLogProbabilities Computes emission probablities between States of current Observation and States of next Observation
/*
	prevLayer - previous Observation
	currentLayer - current Observation
*/
func (matcher *MapMatcher) computeTransitionLogProbabilities(prevLayer, currentLayer *CandidateLayer, routeLengths map[int]map[int]float64) error {
	straightDistance := prevLayer.Observation.GeoPoint.DistanceTo(currentLayer.Observation.GeoPoint)
	timeDiff := currentLayer.Observation.dateTime.Sub(prevLayer.Observation.dateTime).Seconds()
	for i := range prevLayer.States {
		from := prevLayer.States[i]
		for j := range currentLayer.States {
			to := currentLayer.States[j]
			if routeLengths[from.RoadPositionID][to.RoadPositionID] < 0 {
				continue
			}
			if routeLengths[from.RoadPositionID][to.RoadPositionID] > ROUTE_LENGTH_THRESHOLD {
				// Restrict max route length assuming that too large length is just bad
				currentLayer.AddTransitionProbability(from, to, -ROUTE_LENGTH_THRESHOLD)
				continue
			}
			transitionLogProbability, err := matcher.hmmParams.TransitionLogProbability(routeLengths[from.RoadPositionID][to.RoadPositionID], straightDistance, timeDiff)
			if err != nil {
				return err
			}
			currentLayer.AddTransitionProbability(from, to, transitionLogProbability)
		}
	}
	return nil
}
