package horizon

import (
	"fmt"
	"math"

	"github.com/LdDl/viterbi"
	"github.com/golang/geo/s2"
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
func NewMapMatcher(props *HmmProbabilities, graphFileName string) (*MapMatcher, error) {
	mm := &MapMatcher{
		hmmParams: props,
	}
	mapEngine, err := prepareEngine(graphFileName)
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
		return MatcherResult{}, fmt.Errorf("Number of gps measurements need to be 3 atleast")
	}

	stateID := 0
	obsState := make(map[int]*CandidateLayer)
	layers := []RoadPositions{}
	for i := range gpsMeasurements {
		s2point := gpsMeasurements[i].Point
		closest, _ := matcher.engine.s2Storage.NearestNeighborsInRadius(gpsMeasurements[i].Point, statesRadiusMeters, maxStates)
		localStates := make(RoadPositions, len(closest))
		for j := range closest {
			s2polyline := matcher.engine.s2Storage.edges[closest[j].edgeID]
			m := s2polyline.Source
			n := s2polyline.Target
			edge := matcher.engine.edges[m][n]
			proj, fraction := calcProjection(*edge.Polyline, s2point)
			latLng := s2.LatLngFromPoint(proj)
			_ = fraction
			// @todo determine which vertex is better to use. something like below, maybe?
			// choosenID := n
			// if i != 0 {
			// 	if fraction > 0.5 {
			// 		choosenID = m
			// 	} else {
			// 		choosenID = n
			// 	}
			// }
			roadPos := NewRoadPositionFromLonLat(stateID, n, edge, latLng.Lng.Degrees(), latLng.Lat.Degrees(), 4326)
			localStates[j] = roadPos
			stateID++
		}
		layers = append(layers, localStates)
		obsState[gpsMeasurements[i].id] = NewCandidateLayer(gpsMeasurements[i], localStates)
	}
	chRoutes := make(map[int]map[int][]int64)

	routeLengths := make(lengths)

	for i := 1; i < len(layers); i++ {
		prevStates := layers[i-1]
		currentStates := layers[i]
		for m := range prevStates {
			if _, ok := chRoutes[prevStates[m].RoadPositionID]; !ok {
				chRoutes[prevStates[m].RoadPositionID] = make(map[int][]int64)
			}

			one2manyVertices := []int64{}
			one2manyStatesIndices := []int{}
			for n := range currentStates {
				if prevStates[m].GraphVertex == currentStates[n].GraphVertex { // Handle circular movements
					ans := prevStates[m].Projected.DistanceTo(currentStates[n].Projected)
					chRoutes[prevStates[m].RoadPositionID][currentStates[n].RoadPositionID] = []int64{prevStates[m].GraphEdge.Source, prevStates[m].GraphEdge.Target}
					routeLengths.AddRouteLength(prevStates[m], currentStates[n], ans)
					// fmt.Printf("(1) from (%d->%d) to ans (%d->%d) : %f\n", prevStates[m].GraphEdge.Source, prevStates[m].GraphEdge.Target, currentStates[n].GraphEdge.Source, currentStates[n].GraphEdge.Target, ans)
				} else {
					one2manyVertices = append(one2manyVertices, currentStates[n].GraphVertex)
					one2manyStatesIndices = append(one2manyStatesIndices, n)
				}
			}
			answers, pathes := matcher.engine.graph.ShortestPathOneToMany(prevStates[m].GraphVertex, one2manyVertices)
			for i := range answers {
				if answers[i] == -1 {
					answers[i] = math.MaxFloat64
				}
				chRoutes[prevStates[m].RoadPositionID][currentStates[one2manyStatesIndices[i]].RoadPositionID] = pathes[i]
				routeLengths.AddRouteLength(prevStates[m], currentStates[one2manyStatesIndices[i]], answers[i])
				// fmt.Printf("(2) from (%d->%d) to ans (%d->%d) : %f\n", prevStates[m].GraphEdge.Source, prevStates[m].GraphEdge.Target, currentStates[one2manyStatesIndices[i]].GraphEdge.Source, currentStates[one2manyStatesIndices[i]].GraphEdge.Target, anss[i])
			}
		}
	}

	// for i := 1; i < len(layers); i++ {
	// 	prevStates := layers[i-1]
	// 	currentStates := layers[i]
	// 	for m := range prevStates {
	// 		if _, ok := chRoutes[prevStates[m].RoadPositionID]; !ok {
	// 			chRoutes[prevStates[m].RoadPositionID] = make(map[int][]int64)
	// 		}
	// 		for n := range currentStates {
	// 			if prevStates[m].GraphVertex == currentStates[n].GraphVertex {
	// 				ans := prevStates[m].Projected.DistanceTo(currentStates[n].Projected)
	// 				chRoutes[prevStates[m].RoadPositionID][currentStates[n].RoadPositionID] = []int64{prevStates[m].GraphEdge.Source, prevStates[m].GraphEdge.Target}
	// 				routeLengths.AddRouteLength(prevStates[m], currentStates[n], ans)
	// 				// fmt.Printf("(1) from (%d->%d) to ans (%d->%d) : %f\n", prevStates[m].GraphEdge.Source, prevStates[m].GraphEdge.Target, currentStates[n].GraphEdge.Source, currentStates[n].GraphEdge.Target, ans)
	// 			} else {
	// 				ans, path := matcher.engine.graph.ShortestPath(prevStates[m].GraphVertex, currentStates[n].GraphVertex)
	// 				if ans == -1 {
	// 					ans = math.MaxFloat64
	// 				}
	// 				chRoutes[prevStates[m].RoadPositionID][currentStates[n].RoadPositionID] = path
	// 				routeLengths.AddRouteLength(prevStates[m], currentStates[n], ans)
	// 				// fmt.Printf("(2) from (%d->%d) to ans (%d->%d) : %f\n", prevStates[m].GraphEdge.Source, prevStates[m].GraphEdge.Target, currentStates[n].GraphEdge.Source, currentStates[n].GraphEdge.Target, ans)
	// 			}

	// 		}
	// 	}
	// }

	v, err := matcher.PrepareViterbi(obsState, routeLengths, gpsMeasurements)
	if err != nil {
		return MatcherResult{}, err
	}

	vpath := v.EvalPathLogProbabilities()

	if len(vpath.Path) != len(gpsMeasurements) {
		return MatcherResult{}, fmt.Errorf("Number of states in final path != number of observations. Should be unreachable error")
	}

	result := matcher.prepareResult(vpath, gpsMeasurements, chRoutes)

	return result, nil
}

// PrepareViterbi Prepares engine for doing Viterbi's algorithm (see https://github.com/LdDl/viterbi/blob/master/viterbi.go#L25)
/*
	states - set of States
	gpsMeasurements - set of Observations
*/
func (matcher *MapMatcher) PrepareViterbi(obsStates map[int]*CandidateLayer, routeLengths map[int]map[int]float64, gpsMeasurements []*GPSMeasurement) (*viterbi.Viterbi, error) {
	v := &viterbi.Viterbi{}
	for i := range obsStates {
		for _, st := range obsStates[i].States {
			v.AddState(st)
		}
	}
	for i := range gpsMeasurements {
		v.AddObservation(gpsMeasurements[i])
	}

	layers := make([]*CandidateLayer, len(gpsMeasurements))
	prevLayer := &CandidateLayer{}

	// I guess this is ugly.
	// @todo Refactor data prerapartion for Viterbi's algorithm
	for i := range gpsMeasurements {
		currentLayer := obsStates[gpsMeasurements[i].id]
		matcher.computeEmissionLogProbabilities(currentLayer)
		// @experimental
		// currentLayer.EmissionLogProbabilities = softmaxEmissions(currentLayer.EmissionLogProbabilities)
		if i == 0 {
			for j := range currentLayer.EmissionLogProbabilities {
				v.PutStartProbability(currentLayer.EmissionLogProbabilities[j].rp, currentLayer.EmissionLogProbabilities[j].prob)
			}
		} else {
			err := matcher.computeTransitionLogProbabilities(prevLayer, currentLayer, routeLengths)
			if err != nil {
				return nil, err
			}
		}
		for j := range currentLayer.EmissionLogProbabilities {
			v.PutEmissionProbability(currentLayer.EmissionLogProbabilities[j].rp, gpsMeasurements[i], currentLayer.EmissionLogProbabilities[j].prob)
		}
		prevLayer = currentLayer
		layers[i] = currentLayer
	}

	for s := range layers {
		step := layers[s]
		// @experimental
		// step.TransitionLogProbabilities = softmaxTransitions(step.TransitionLogProbabilities)
		for i := range step.TransitionLogProbabilities {
			v.PutTransitionProbability(step.TransitionLogProbabilities[i].from, step.TransitionLogProbabilities[i].to, step.TransitionLogProbabilities[i].prob)
		}
	}

	return v, nil
}

// @experimental Play with normalization of emission probabilities.
func softmaxEmissions(a []emission) []emission {
	sum := 0.0
	output := make([]emission, len(a))
	for i := range a {
		output[i] = emission{a[i].rp, math.Exp(a[i].prob)}
		sum += output[i].prob
	}
	for i := range output {
		output[i].prob = output[i].prob / sum
	}
	return output
}

// @experimental Play with normalization of transition probabilities.
func softmaxTransitions(a []transition) []transition {
	sum := 0.0
	output := make([]transition, len(a))
	for i := range a {
		output[i] = transition{a[i].from, a[i].to, math.Exp(a[i].prob)}
		sum += output[i].prob
	}
	for i := range output {
		output[i].prob = output[i].prob / sum
	}
	return output
}

// computeEmissionLogProbabilities Computes emission probablities between Observation and corresponding States
/*
	layer - wrapper of Observation
*/
func (matcher *MapMatcher) computeEmissionLogProbabilities(layer *CandidateLayer) {
	for i := range layer.States {
		distance := layer.States[i].Projected.DistanceTo(layer.Observation.GeoPoint)
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
			transitionLogProbability, err := matcher.hmmParams.TransitionLogProbability(routeLengths[from.RoadPositionID][to.RoadPositionID], straightDistance, timeDiff)
			if err != nil {
				return err
			}
			currentLayer.AddTransitionProbability(from, to, transitionLogProbability)
		}
	}
	return nil
}
