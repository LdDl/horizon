package horizon

import (
	"fmt"
)

// CostStrategy describes how edge cost is determined from CSV columns.
type CostStrategy int

const (
	// CostStrategyWeight uses the "weight" column directly (backward compatibility).
	CostStrategyWeight CostStrategy = iota
	// CostStrategyTime computes cost as length_meters / (free_speed / 3.6), i.e. travel time in seconds.
	CostStrategyTime
	// CostStrategyDistance uses "length_meters" as cost (no speed data available).
	CostStrategyDistance
)

func (s CostStrategy) String() string {
	switch s {
	case CostStrategyWeight:
		return "weight (backward-compat: raw weight column)"
	case CostStrategyTime:
		return "time (length_meters / free_speed)"
	case CostStrategyDistance:
		return "distance (length_meters, no speed data)"
	default:
		return "unknown"
	}
}

// CostResolver determines edge cost from CSV records using a fallback chain:
//
//  1. "weight" column exists => use it directly (backward compat, logged)
//  2. "length_meters" + "free_speed" exist => cost = length / (speed / 3.6) (travel time in seconds)
//  3. only "length_meters" exists => cost = length_meters (distance-based)
//  4. nothing useful => error
//
// @todo: maybe in future I could implement more complex strategies (e.g. parse geom and eval its length and so on)
type CostResolver struct {
	lookup   *CSVLookup
	strategy CostStrategy
}

// NewCostResolver inspects CSV headers and selects the best cost strategy.
// It logs which strategy was chosen.
func NewCostResolver(lookup *CSVLookup) (*CostResolver, error) {
	r := &CostResolver{lookup: lookup}

	hasWeight := lookup.Has("weight")
	hasLength := lookup.Has("length_meters")
	hasSpeed := lookup.Has("free_speed")

	switch {
	case hasWeight:
		r.strategy = CostStrategyWeight
		if hasLength && hasSpeed {
			fmt.Println("[cost-resolver] using 'weight' column (length_meters + free_speed also available but weight takes priority for backward compatibility)")
		} else {
			fmt.Println("[cost-resolver] using 'weight' column")
		}
	case hasLength && hasSpeed:
		r.strategy = CostStrategyTime
		fmt.Println("[cost-resolver] using time-based cost: length_meters / (free_speed / 3.6)")
	case hasLength:
		r.strategy = CostStrategyDistance
		fmt.Println("[cost-resolver] using distance-based cost: length_meters (no free_speed column found)")
	default:
		return nil, fmt.Errorf("can't determine edge cost: need 'weight' or 'length_meters' column in CSV header")
	}

	return r, nil
}

// Strategy returns the selected cost strategy.
func (r *CostResolver) Strategy() CostStrategy {
	return r.strategy
}

// Resolve computes the cost for a single CSV record.
func (r *CostResolver) Resolve(record []string) (float64, error) {
	switch r.strategy {
	case CostStrategyWeight:
		return r.lookup.Float64(record, "weight")
	case CostStrategyTime:
		length, err := r.lookup.Float64(record, "length_meters")
		if err != nil {
			return 0, err
		}
		speed, err := r.lookup.Float64(record, "free_speed")
		if err != nil {
			return 0, err
		}
		if speed <= 0 {
			return 0, fmt.Errorf("free_speed must be positive, got %f", speed)
		}
		// speed is km/h => m/s = speed / 3.6
		// cost = length_meters / (speed_kmh / 3.6) = travel time in seconds
		return length / (speed / 3.6), nil
	case CostStrategyDistance:
		return r.lookup.Float64(record, "length_meters")
	default:
		return 0, fmt.Errorf("unknown cost strategy: %d", r.strategy)
	}
}
