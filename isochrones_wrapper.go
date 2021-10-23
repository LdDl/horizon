package horizon

// IsochronesResult Representation of isochrones algorithm's output
/*
 */
type IsochronesResult struct {
}

// FindIsochrones Find shortest path between two obserations (not necessary GPS points).
/*
	NOTICE: this function snaps point to only one nearest vertex (without multiple candidates for provided point)
	source - source for outcoming isochrones
	maxCost - max cost restriction for single isochrone line
*/
func (matcher *MapMatcher) FindIsochrones(source *GPSMeasurement, maxCost float64) (*IsochronesResult, error) {
	return nil, nil
}
