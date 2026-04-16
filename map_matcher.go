package horizon

import (
	"fmt"
	"math"
	"runtime"
	"sort"
	"sync"

	"github.com/LdDl/ch"
	"github.com/LdDl/horizon/spatial"
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
	viterbiSemaphore - limits concurrent Viterbi computations globally
*/
type MapMatcher struct {
	hmmParams        *HmmProbabilities
	engine           *MapEngine
	viterbiSemaphore chan struct{}
}

// NewMapMatcherDefault Returns pointer to created MapMatcher with default parameters
func NewMapMatcherDefault() *MapMatcher {
	return &MapMatcher{
		hmmParams:        HmmProbabilitiesDefault(),
		viterbiSemaphore: make(chan struct{}, runtime.NumCPU()),
	}
}

// NewMapMatcherFromFiles Returns pointer to created MapMatcher from CSV files
/*
	props - parameters of Hidden Markov Model
	edgesFilename - path to the edges CSV file (e.g., "graph.csv")

	This function expects three CSV files with the same prefix:
	  - {prefix}.csv - edges file (required)
	    Format: from_vertex_id;to_vertex_id;weight;geom;was_one_way;edge_id
	    geom is GeoJSON LineString

	  - {prefix}_vertices.csv - vertices file (required)
	    Format: vertex_id;order_pos;importance;geom
	    geom is GeoJSON Point
	    order_pos and importance are used for contraction hierarchies

	  - {prefix}_shortcuts.csv - shortcuts file (required, can be empty with header only)
	    Format: from_vertex_id;to_vertex_id;weight;via_vertex_id
	    These are precomputed contraction hierarchy shortcuts.
	    If empty, shortcuts will be computed via PrepareContractionHierarchies()

	Example: if edgesFilename is "./data/roads.csv", it will look for:
	  - ./data/roads.csv
	  - ./data/roads_vertices.csv
	  - ./data/roads_shortcuts.csv
*/
func NewMapMatcherFromFiles(props *HmmProbabilities, edgesFilename string) (*MapMatcher, error) {
	mm := &MapMatcher{
		hmmParams:        props,
		viterbiSemaphore: make(chan struct{}, runtime.NumCPU()),
	}
	mapEngine, err := prepareEngine(edgesFilename)
	if err != nil {
		return nil, err
	}
	mm.engine = mapEngine
	return mm, nil
}

// NewMapMatcher returns pointer to created MapMatcher with provided options
func NewMapMatcher(ops ...func(*MapMatcher)) *MapMatcher {
	mm := &MapMatcher{
		hmmParams:        HmmProbabilitiesDefault(),
		viterbiSemaphore: make(chan struct{}, runtime.NumCPU()),
	}
	for _, op := range ops {
		op(mm)
	}
	return mm
}

// WithHmmParameters sets the HMM parameters for the matcher
func WithHmmParameters(params *HmmProbabilities) func(*MapMatcher) {
	return func(matcher *MapMatcher) {
		matcher.hmmParams = params
	}
}

// WithMapEngine sets the map engine for the matcher
func WithMapEngine(engine *MapEngine) func(*MapMatcher) {
	return func(matcher *MapMatcher) {
		matcher.engine = engine
	}
}

// Segment represents a continuous matched segment to process separately (split at break points)
type Segment struct {
	// First observation index in this segment
	start int
	// Last observation index in this segment
	end int
	// Route lengths for this segment only
	routeLengths lengths
}

// cachedRoute is a structure to hold cached RAW shortest path results
type cachedRoute struct {
	cost float64
	path []int64
}

// viterbiResult is for processing each segment (Viterbi) separately using goroutines
type viterbiResult struct {
	vpath viterbi.ViterbiPath
	err   error
}

// unmatchedObs is for tracking unmatched GPS observations
type unmatchedObs struct {
	originalIdx int
	gps         *GPSMeasurement
}

// indexedSubMatch needed to merge matched and unmatched SubMatches in correct order
type indexedSubMatch struct {
	firstObsIdx int
	subMatch    SubMatch
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
	closestSets := [][]spatial.NearestObject{}

	// Array for no candidates found
	unmatchedObservations := []unmatchedObs{}

	// Maps original index to engineGpsMeasurements index (for matched points)
	originalToEngineIdx := make(map[int]int)

	for i := 0; i < len(gpsMeasurements); i++ {
		var closest []spatial.NearestObject
		var err error
		if statesRadiusMeters < 0 {
			closest, err = matcher.engine.storage.FindNearest(gpsMeasurements[i].Point, maxStates)
		} else {
			closest, err = matcher.engine.storage.FindNearestInRadius(gpsMeasurements[i].Point, statesRadiusMeters, maxStates)
		}
		if err != nil {
			return MatcherResult{}, errors.Wrapf(err, "Can't find neighbors for point: '%s' (states radius = %f, max states = %d)", gpsMeasurements[i].Point, statesRadiusMeters, maxStates)
		}
		if len(closest) == 0 {
			// Track unmatched observation instead of just skipping as it done before
			unmatchedObservations = append(unmatchedObservations, unmatchedObs{
				originalIdx: i,
				gps:         gpsMeasurements[i],
			})
			continue
		}
		originalToEngineIdx[i] = len(engineGpsMeasurements)
		engineGpsMeasurements = append(engineGpsMeasurements, gpsMeasurements[i])
		closestSets = append(closestSets, closest)
	}

	// If no matched observations, return all as unmatched with default SubMatches
	if len(engineGpsMeasurements) == 0 {
		allUnmatched := make([]SubMatch, len(unmatchedObservations))
		for i, unmatched := range unmatchedObservations {
			allUnmatched[i] = SubMatch{
				Observations: []ObservationResult{{
					Observation: unmatched.gps,
					IsMatched:   false,
					Code:        CODE_NO_CANDIDATES,
				}},
				Probability: 0,
			}
		}
		return MatcherResult{SubMatches: allUnmatched}, nil
	}

	obsState := make([]*CandidateLayer, len(engineGpsMeasurements))
	for i := 0; i < len(engineGpsMeasurements); i++ {
		s2point := engineGpsMeasurements[i].Point
		srid := engineGpsMeasurements[i].GeoPoint.SRID()
		closest := closestSets[i]
		localStates := make(RoadPositions, len(closest))
		for j := range closest {
			s2polyline := matcher.engine.storage.GetEdge(closest[j].EdgeID)
			m := s2polyline.Source
			n := s2polyline.Target
			edge := matcher.engine.edges.Get(m, n)

			// Use appropriate projection based on SRID
			var proj s2.Point
			var fraction float64
			var next int
			var lon, lat float64

			if srid == 4326 {
				// Spherical geometry (WGS84)
				proj, fraction, next = spatial.CalcProjectionCached(edge, s2point)
				latLng := s2.LatLngFromPoint(proj)
				lon = latLng.Lng.Degrees()
				lat = latLng.Lat.Degrees()
			} else {
				// Euclidean geometry (SRID=0 or other)
				proj, fraction, next = spatial.CalcProjectionEuclidean(*edge.Polyline, s2point)
				lon = proj.Vector.X
				lat = proj.Vector.Y
			}

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
			roadPos := NewRoadPositionFromLonLat(stateID, pickedGraphVertex, routingGraphVertex, edge, lon, lat, srid)
			roadPos.beforeProjection = edge.Weight * fraction
			roadPos.afterProjection = edge.Weight * (1 - fraction)
			roadPos.next = next
			localStates[j] = roadPos
			stateID++
		}
		layers = append(layers, localStates)
		obsState[i] = NewCandidateLayer(engineGpsMeasurements[i], localStates)
	}
	chRoutes := make(map[[2]int][]int64)

	segments := []Segment{}
	segmentStart := 0
	currentRouteLengths := make(lengths)

	// vertex-level path cache to avoid recomputing same routes
	// key: [fromVertex, toVertex] -> {rawCost, rawPath}
	vertexCache := make(map[[2]int64]cachedRoute)

	// @todo: Consider to use ShortestPathOneToMany (need to deal with the order of writing data to to chRoutes and routeLengths)
	//
	// Break-point detection for sub-matches:
	// For each consecutive pair of layers (i-1, i) we try to build a route between every candidate pair.
	// If NO candidate pair produces a valid path, we treat layer i as a break point - the preceding part
	// becomes a finalized Segment (→ its own SubMatch), and a fresh segment starts at layer i.
	// `anyValidRoute` is the fused, in-loop replacement of the old isBreakPoint() helper: it is set
	// to true in every branch that writes a non-empty path into chRoutes[key]; if it stays false after
	// the full M×N scan, we've detected a break point.
	for i := 1; i < len(layers); i++ {
		prevStates := layers[i-1]
		currentStates := layers[i]
		anyValidRoute := false
		for m := range prevStates {
			for n := range currentStates {
				key := [2]int{prevStates[m].RoadPositionID, currentStates[n].RoadPositionID}
				if prevStates[m].RoutingGraphVertex == currentStates[n].RoutingGraphVertex {
					if prevStates[m].GraphEdge.ID == currentStates[n].GraphEdge.ID {
						// Trivial route on the same edge (same routing vertex, same edge) - never a break into submatches
						ans := prevStates[m].Projected.DistanceTo(currentStates[n].Projected)
						chRoutes[key] = []int64{prevStates[m].GraphEdge.Source, prevStates[m].GraphEdge.Target}
						currentRouteLengths.AddRouteLength(prevStates[m], currentStates[n], ans)
						anyValidRoute = true
					} else {
						// We should jump to source vertex of current state, since edges are not the same
						rawCost, rawPath := getCachedPath(matcher.engine.queryPool, vertexCache, matcher.engine.vertexStrongComponent, prevStates[m].RoutingGraphVertex, currentStates[n].GraphEdge.Source)
						routeDistMeters, finalPath := matcher.resolveRoute(rawCost, rawPath, currentStates[n].GraphEdge.Target)
						chRoutes[key] = finalPath
						currentRouteLengths.AddRouteLength(prevStates[m], currentStates[n], routeDistMeters)
						// Only a non-empty CH path counts for break-point detection
						if len(finalPath) > 0 {
							anyValidRoute = true
						}
					}
					continue
				}
				// Same edge but different RoutingGraphVertex: use projected distance only if moving forward
				// (beforeProjection increases => fraction increases => forward movement along the edge).
				// If moving backward, the edge is likely a reverse-direction candidate => fall through to CH routing.
				if prevStates[m].GraphEdge.ID == currentStates[n].GraphEdge.ID && prevStates[m].beforeProjection <= currentStates[n].beforeProjection {
					ans := prevStates[m].Projected.DistanceTo(currentStates[n].Projected)
					chRoutes[key] = []int64{prevStates[m].GraphEdge.Source, prevStates[m].GraphEdge.Target}
					currentRouteLengths.AddRouteLength(prevStates[m], currentStates[n], ans)
					anyValidRoute = true
					continue
				}
				rawCost, rawPath := getCachedPath(matcher.engine.queryPool, vertexCache, matcher.engine.vertexStrongComponent, prevStates[m].RoutingGraphVertex, currentStates[n].RoutingGraphVertex)

				// Since we are doing Edge(target)-Edge(target) Dijkstra's call most of time we could:
				// 1) add penalty for source edge by adding remaining distance to target vertex of source edge
				// 2) add advantage for target edge by subtracting remaining distance to target vertex of target edge
				// @todo: this could lead to negative values. Need to investigate when it happens
				// finalCost = (finalCost + prevStates[m].afterProjection) - currentStates[n].afterProjection
				routeDistMeters, finalPath := matcher.resolveRoute(rawCost, rawPath, currentStates[n].GraphEdge.Target)
				chRoutes[key] = finalPath
				currentRouteLengths.AddRouteLength(prevStates[m], currentStates[n], routeDistMeters)
				// Only a non-empty CH path counts for break-point detection
				if len(finalPath) > 0 {
					anyValidRoute = true
				}
			}
		}

		// Break point (submatches): the entire M×N scan produced no valid route - the trace must be split here.
		// Finalize the current sub-match [segmentStart .. i-1] and restart with fresh routeLengths at i.
		// Each Segment later becomes an independent SubMatch in MatcherResult.SubMatches.
		if !anyValidRoute {
			segments = append(segments, Segment{
				start:        segmentStart,
				end:          i - 1,
				routeLengths: currentRouteLengths,
			})
			segmentStart = i
			currentRouteLengths = make(lengths)
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

	// Make sure the last segment is there with its routeLengths
	segments = append(segments, Segment{
		start:        segmentStart,
		end:          len(layers) - 1,
		routeLengths: currentRouteLengths,
	})

	// Run Viterbi in parallel for each segment with bounded concurrency
	results := make([]viterbiResult, len(segments))
	var wg sync.WaitGroup
	wg.Add(len(segments))

	for i := range segments {
		go func(i int) {
			matcher.viterbiSemaphore <- struct{}{} // Acquire
			defer func() {
				<-matcher.viterbiSemaphore // Release
				wg.Done()
			}()

			seg := &segments[i]
			segmentObsState := obsState[seg.start : seg.end+1]
			segmentGPS := engineGpsMeasurements[seg.start : seg.end+1]

			// Handle single-point segment: no Viterbi needed, just pick best candidate
			if len(segmentObsState) == 1 {
				if len(segmentObsState[0].States) == 0 {
					results[i] = viterbiResult{err: fmt.Errorf("no candidates for single-point segment at index %d", seg.start)}
					return
				}
				// Pick first candidate (already sorted by dist to observation)
				bestCandidate := segmentObsState[0].States[0]
				// Compute emission log probability for single point
				// Note: Viterbi counts emission twice for first observation (start + emission = 2 * emission)
				sigma := matcher.hmmParams.sigma
				if segmentObsState[0].Observation.accuracy > 0 {
					sigma = segmentObsState[0].Observation.accuracy
				}
				distance := bestCandidate.Projected.DistanceTo(segmentObsState[0].Observation.GeoPoint)
				emissionLogProb := LogNormalDistribution(sigma, distance)
				results[i] = viterbiResult{
					vpath: viterbi.ViterbiPath{
						Path:        []viterbi.State{bestCandidate},
						Probability: 2 * emissionLogProb,
					},
				}
				return
			}

			v, err := matcher.PrepareViterbi(segmentObsState, seg.routeLengths, segmentGPS)
			if err != nil {
				results[i] = viterbiResult{err: err}
				return
			}

			vpath, err := v.EvalPathLogProbabilities()
			if err != nil {
				results[i] = viterbiResult{err: errors.Wrapf(err, "Can't evaluate path log probabilities for segment [%d:%d]", seg.start, seg.end)}
				return
			}

			if len(vpath.Path) != len(segmentGPS) {
				results[i] = viterbiResult{err: fmt.Errorf("number of states in final path != number (%d and %d) of observations for segment [%d:%d]", len(vpath.Path), len(segmentGPS), seg.start, seg.end)}
				return
			}

			results[i] = viterbiResult{vpath: vpath}
		}(i)
	}

	wg.Wait()

	// Check for errors and prepare subMatches sequentially
	subMatches := make([]SubMatch, 0, len(segments))
	for i := range segments {
		if results[i].err != nil {
			return MatcherResult{}, results[i].err
		}

		seg := &segments[i]
		segmentLayers := layers[seg.start : seg.end+1]
		segmentGPS := engineGpsMeasurements[seg.start : seg.end+1]

		if ViterbiDebug {
			fmt.Printf("Segment [%d:%d] prob: %f\n", seg.start, seg.end, results[i].vpath.Probability)
			fmt.Println("path:")
			for j := range results[i].vpath.Path {
				fmt.Println("\t", results[i].vpath.Path[j].(*RoadPosition).GraphEdge.ID, results[i].vpath.Path[j].(*RoadPosition).ID())
			}
		}

		subMatch := matcher.prepareSubMatch(results[i].vpath, segmentGPS, segmentLayers, chRoutes)
		subMatches = append(subMatches, subMatch)
	}

	// If no unmatched met, return matched SubMatches directly
	if len(unmatchedObservations) == 0 {
		return MatcherResult{SubMatches: subMatches}, nil
	}

	// Create SubMatches for unmatched observations
	unmatchedSubMatches := make([]SubMatch, len(unmatchedObservations))
	for i, unmatched := range unmatchedObservations {
		unmatchedSubMatches[i] = SubMatch{
			Observations: []ObservationResult{{
				Observation: unmatched.gps,
				IsMatched:   false,
				Code:        CODE_NO_CANDIDATES,
			}},
			Probability: 0,
		}
	}

	// Merge matched and unmatched SubMatches in order of original observation indices
	// Each SubMatch's position is determined by its first observation's ID
	allSubMatches := make([]indexedSubMatch, 0, len(subMatches)+len(unmatchedSubMatches))

	for _, sm := range subMatches {
		if len(sm.Observations) > 0 {
			allSubMatches = append(allSubMatches, indexedSubMatch{
				firstObsIdx: sm.Observations[0].Observation.ID(),
				subMatch:    sm,
			})
		}
	}

	for i, sm := range unmatchedSubMatches {
		allSubMatches = append(allSubMatches, indexedSubMatch{
			firstObsIdx: unmatchedObservations[i].originalIdx,
			subMatch:    sm,
		})
	}

	// Sort by first observation index
	sort.Slice(allSubMatches, func(i, j int) bool {
		return allSubMatches[i].firstObsIdx < allSubMatches[j].firstObsIdx
	})

	// Extract sorted SubMatches
	finalSubMatches := make([]SubMatch, len(allSubMatches))
	for i, ism := range allSubMatches {
		finalSubMatches[i] = ism.subMatch
	}

	return MatcherResult{SubMatches: finalSubMatches}, nil
}

// PrepareViterbi Prepares engine for doing Viterbi's algorithm (see https://github.com/LdDl/viterbi/blob/master/viterbi.go#L25)
/*
	states - set of States
	gpsMeasurements - set of Observations
*/
func (matcher *MapMatcher) PrepareViterbi(obsStates []*CandidateLayer, routeLengths lengths, gpsMeasurements []*GPSMeasurement) (*viterbi.Viterbi, error) {
	v := viterbi.New()

	// statesIndx is used only for human-readable debug output; skip populating in release mode
	var statesIndx map[int]int
	if ViterbiDebug {
		statesIndx = make(map[int]int)
		idx := 0
		for i := range obsStates {
			for j := range obsStates[i].States {
				v.AddState(obsStates[i].States[j])
				statesIndx[obsStates[i].States[j].ID()] = idx
				fmt.Printf(`CustomState{Name: "%d", id: %d}%s`, obsStates[i].States[j].GraphEdge.ID, obsStates[i].States[j].ID(), ",\n")
				idx++
			}
			fmt.Println()
		}
		fmt.Println()
	} else {
		for i := range obsStates {
			for j := range obsStates[i].States {
				v.AddState(obsStates[i].States[j])
			}
		}
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
	// Use observation's accuracy if provided, else default sigma
	sigma := matcher.hmmParams.sigma
	if layer.Observation.accuracy > 0 {
		sigma = layer.Observation.accuracy
	}

	// Loop-invariant: log(1/(sigma*sqrt(2*pi))) and 1/sigma precomputed once per layer
	logConst := math.Log(1.0 / (sqrtTwoPi * sigma))
	invSigma := 1.0 / sigma
	// Pre-allocate to exact size to avoid append-growth copies
	if cap(layer.EmissionLogProbabilities) < len(layer.States) {
		layer.EmissionLogProbabilities = make([]emission, 0, len(layer.States))
	}
	for i := range layer.States {
		distance := layer.States[i].Projected.DistanceTo(layer.Observation.GeoPoint)
		xOverSigma := distance * invSigma
		emissionLogProb := logConst - 0.5*xOverSigma*xOverSigma
		layer.AddEmissionProbability(layer.States[i], emissionLogProb)
	}
}

// computeTransitionLogProbabilities Computes emission probablities between States of current Observation and States of next Observation
/*
	prevLayer - previous Observation
	currentLayer - current Observation
*/
func (matcher *MapMatcher) computeTransitionLogProbabilities(prevLayer, currentLayer *CandidateLayer, routeLengths lengths) error {
	straightDistance := prevLayer.Observation.GeoPoint.DistanceTo(currentLayer.Observation.GeoPoint)
	timeDiff := currentLayer.Observation.dateTime.Sub(prevLayer.Observation.dateTime).Seconds()
	// Pre-allocate to worst-case K*K size to avoid append-growth copies
	maxPairs := len(prevLayer.States) * len(currentLayer.States)
	if cap(currentLayer.TransitionLogProbabilities) < maxPairs {
		currentLayer.TransitionLogProbabilities = make([]transition, 0, maxPairs)
	}
	for i := range prevLayer.States {
		from := prevLayer.States[i]
		for j := range currentLayer.States {
			to := currentLayer.States[j]
			rl := routeLengths[[2]int{from.RoadPositionID, to.RoadPositionID}]
			if rl < 0 {
				continue
			}
			if rl > ROUTE_LENGTH_THRESHOLD {
				// Restrict max route length assuming that too large length is just bad
				currentLayer.AddTransitionProbability(from, to, -ROUTE_LENGTH_THRESHOLD)
				continue
			}
			transitionLogProbability, err := matcher.hmmParams.TransitionLogProbability(rl, straightDistance, timeDiff)
			if err != nil {
				return err
			}
			currentLayer.AddTransitionProbability(from, to, transitionLogProbability)
		}
	}
	return nil
}

// getCachedPath is a helper function to get or compute shortest path with caching
// It uses SCC (Strongly Connected Components) to quickly reject impossible routes
func getCachedPath(queryPool *ch.QueryPool, vertexCache map[[2]int64]cachedRoute, vertexSCC map[int64]int64, fromVertex, toVertex int64) (float64, []int64) {
	// SCC check: if vertices are in different SCCs, no path exists
	fromSCC, fromOK := vertexSCC[fromVertex]
	toSCC, toOK := vertexSCC[toVertex]
	if fromOK && toOK && fromSCC != toSCC {
		// Different SCCs - no path possible, return immediately
		return -1, nil
	}

	key := [2]int64{fromVertex, toVertex}
	// Check cache first
	if cached, ok := vertexCache[key]; ok {
		return cached.cost, cached.path
	}
	// Compute and cache using thread-safe query pool
	rawCost, rawPath := queryPool.ShortestPath(fromVertex, toVertex)
	vertexCache[key] = cachedRoute{cost: rawCost, path: rawPath}
	return rawCost, rawPath
}

// resolveRoute builds final path from cached CH result and computes route distance in meters.
// rawPath comes from cache and must not be mutated, so a copy is made.
func (matcher *MapMatcher) resolveRoute(rawCost float64, rawPath []int64, targetVertex int64) (float64, []int64) {
	if rawCost < 0 {
		return math.MaxFloat64, nil
	}
	finalPath := make([]int64, len(rawPath), len(rawPath)+1)
	copy(finalPath, rawPath)
	finalPath = append(finalPath, targetVertex)
	return matcher.engine.routeDistanceMeters(finalPath), finalPath
}
