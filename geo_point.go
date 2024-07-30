package horizon

import (
	"fmt"

	"github.com/golang/geo/r3"
	"github.com/golang/geo/s2"
)

const (
	// EarthRadius Approximate radius of Earth
	EarthRadius = 6370986.884258304
)

// GeoPoint Wrapper around of s2.Point
/*
	Needs additional field "srid" to determine what algorithm has to be used to calculate distance between objects
	SRID = 0, Euclidean distance
	SRID = 4326 (WGS84), Distance on sphere
*/
type GeoPoint struct {
	s2.Point
	srid int
}

// SRID Returns SRID of point
func (gp *GeoPoint) SRID() int {
	return gp.srid
}

// String Pretty print
func (gp *GeoPoint) String() string {
	return fmt.Sprintf("Point{s2: %v, srid: %d}", gp.Point, gp.srid)
}

// SetSRID Sets SRID for point
func (gp *GeoPoint) SetSRID(srid int) {
	gp.srid = srid
}

// newGeoPoint Returns pointer to created GeoPoint
/*
	@NOT EXPORTED

	lon - longitude (X for SRID = 0)
	lat - latitude (Y for SRID = 0)
	srid - SRID
*/
func newGeoPoint(lon, lat float64, srid int) *GeoPoint {
	switch srid {
	case 4326:
		return &GeoPoint{
			srid:  srid,
			Point: s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lon)),
		}
	default:
		return &GeoPoint{
			srid:  srid,
			Point: s2.Point{Vector: r3.Vector{X: lon, Y: lat, Z: 0}},
		}
	}
}

// NewEuclideanPoint Returns pointer to created GeoPoint with SRID = 0
func NewEuclideanPoint(x, y float64) *GeoPoint {
	return newGeoPoint(x, y, 0)
}

// NewWGS84Point Returns pointer to created GeoPoint with SRID = 4326
func NewWGS84Point(lon, lat float64) *GeoPoint {
	return newGeoPoint(lon, lat, 4326)
}

// DistanceTo Compute distance between two points.
/*
	Algorithm of distance calculation depends on SRID.
	SRID = 0, Euclidean distance
	SRID = 4326 (WGS84), Distance on sphere
*/
func (gp *GeoPoint) DistanceTo(gp2 *GeoPoint) float64 {
	if gp.SRID() != gp2.SRID() {
		// SRIDs has to be equal, need to make assert actually. But we are just use Euclidean distance for this case
		return gp.Vector.Distance(gp2.Vector)
	}
	switch gp.SRID() {
	case 4326:
		// Deal with WGS84
		return gp.Distance(gp2.Point).Radians() * EarthRadius
	default:
		// Deal with planar geometry
		return gp.Vector.Distance(gp2.Vector)
	}
}
