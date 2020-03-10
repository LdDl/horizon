package horizon

type emission struct {
	rp   *RoadPosition
	prob float64
}

type transition struct {
	from *RoadPosition
	to   *RoadPosition
	prob float64
}

type lengths map[int]map[int]float64

func (m lengths) AddRouteLength(from, to *RoadPosition, routeLength float64) {
	if _, ok := m[from.RoadPositionID]; !ok {
		m[from.RoadPositionID] = make(map[int]float64)
		m[from.RoadPositionID][to.RoadPositionID] = routeLength
	} else {
		if _, ok := m[from.RoadPositionID][to.RoadPositionID]; !ok {
			m[from.RoadPositionID][to.RoadPositionID] = routeLength
		}
	}
}

// CandidateLayer Wrapper around Observation
/*
	Observation - observation itself
	States - set of projections on road network
	EmissionLogProbabilities - emission probabilities between Observation and corresponding States
	TransitionLogProbabilities - transition probabilities between States
*/
type CandidateLayer struct {
	Observation                *GPSMeasurement
	States                     RoadPositions
	EmissionLogProbabilities   []emission
	TransitionLogProbabilities []transition
}

// NewCandidateLayer Returns pointer to created CandidateLayer
func NewCandidateLayer(observation *GPSMeasurement, states RoadPositions) *CandidateLayer {
	return &CandidateLayer{
		Observation: observation,
		States:      states,
	}
}

// AddEmissionProbability Append emission probability to slice of emission probablities
func (ts *CandidateLayer) AddEmissionProbability(candidate *RoadPosition, emissionLogProbability float64) {
	ts.EmissionLogProbabilities = append(ts.EmissionLogProbabilities, emission{candidate, emissionLogProbability})
}

// AddTransitionProbability Append transition probability to slice of transition probablities
func (ts *CandidateLayer) AddTransitionProbability(fromPosition, toPosition *RoadPosition, transitionLogProbability float64) {
	ts.TransitionLogProbabilities = append(ts.TransitionLogProbabilities, transition{fromPosition, toPosition, transitionLogProbability})
}
