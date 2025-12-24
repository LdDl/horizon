package horizon

type MatcherCode uint32

const (
	// Everything is fine
	CODE_OK MatcherCode = iota + 900
	// Simply no candidates
	CODE_NO_CANDIDATES
	// Observation is alone (no route to previous or next observation)
	CODE_ALONE_OBSERVATION
)
