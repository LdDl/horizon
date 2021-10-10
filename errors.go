package horizon

import (
	"fmt"
)

var (
	ErrMinumimGPSMeasurements = fmt.Errorf("number of gps measurements need to be 3 atleast")
	ErrCandidatesNotFound     = fmt.Errorf("there is no a single GPS point having candidates")
	ErrTimeDifference         = fmt.Errorf("time difference between subsequent location measurements must be >= 0")
	ErrSourceNotFound         = fmt.Errorf("can't find closest edge for 'source' point")
	ErrSourceHasMoreEdges     = fmt.Errorf("more than 1 edge for 'source' point")
	ErrTargetNotFound         = fmt.Errorf("can't find closest edge for 'target' point")
	ErrTargetHasMoreEdges     = fmt.Errorf("more than 1 edge for 'target' point")
	ErrPathNotFound           = fmt.Errorf("path not found")
)
