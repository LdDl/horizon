package horizon

import (
	"time"

	"github.com/LdDl/horizon/spatial"
)

// GPSMeasurements Set of telematic data
type GPSMeasurements []*GPSMeasurement

// GPSMeasurement Representation of telematic data
/*
	id - unique identifier
	dateTime - timestamp
	GeoPoint - latitude(Y)/longitude(X), pointer to GeoPoint (wrapper)
	accuracy - GPS measurement accuracy in meters (<=0 means use default sigma)
*/
type GPSMeasurement struct {
	*spatial.GeoPoint
	dateTime time.Time
	id       int
	accuracy float64
}

// Accuracy Returns GPS measurement accuracy in meters (0 means use default sigma)
func (gps *GPSMeasurement) Accuracy() float64 {
	return gps.accuracy
}

// ID Returns generated identifier for GPS-point
func (gps *GPSMeasurement) ID() int {
	return gps.id
}

// TM Returns generated (or provided) timestamp for GPS-point
func (gps *GPSMeasurement) TM() time.Time {
	return gps.dateTime
}

// GPSTrack Set of telematic data
type GPSTrack []*GPSMeasurement

// NewGPSMeasurement Returns pointer to created GPSMeasurement
/*
	id - unique identifier
	lon - longitude (X for SRID = 0)
	lat - latitude (Y for SRID = 0)
	srid - SRID (see https://en.wikipedia.org/wiki/Spatial_reference_system), if not provided then SRID(4326) is used. 0 and 4326 are supported.
*/
func NewGPSMeasurement(id int, lon, lat float64, srid int, options ...func(*GPSMeasurement)) *GPSMeasurement {
	gps := &GPSMeasurement{
		dateTime: time.Now(),
		id:       id,
	}
	switch srid {
	case 0:
		gps.GeoPoint = spatial.NewEuclideanPoint(lon, lat)
	case 4326:
		gps.GeoPoint = spatial.NewWGS84Point(lon, lat)
	default:
		gps.GeoPoint = spatial.NewWGS84Point(lon, lat)
	}
	for _, o := range options {
		o(gps)
	}
	return gps
}

// WithGPSTime sets user defined time for GPS measurement
func WithGPSTime(t time.Time) func(*GPSMeasurement) {
	return func(gps *GPSMeasurement) {
		gps.dateTime = t
	}
}

// WithGPSAccuracy sets user defined accuracy for GPS measurement in meters
// Value of <=0 means use default sigma from HmmProbabilities
func WithGPSAccuracy(accuracy float64) func(*GPSMeasurement) {
	return func(gps *GPSMeasurement) {
		gps.accuracy = accuracy
	}
}

// NewGPSMeasurementFromID Returns pointer to created GPSMeasurement
/*
	id - unique identifier (will be converted to time.Time also)
	lon - longitude (X for SRID = 0)
	lat - latitude (Y for SRID = 0)
	srid - SRID (see https://en.wikipedia.org/wiki/Spatial_reference_system), if not provided then SRID(4326) is used. 0 and 4326 are supported.
*/
func NewGPSMeasurementFromID(id int, lon, lat float64, srid ...int) *GPSMeasurement {
	dateTime := time.Now().Add(time.Duration(id) * time.Second)
	gps := GPSMeasurement{
		dateTime: dateTime,
		id:       int(dateTime.Unix()),
	}
	if len(srid) != 0 {
		switch srid[0] {
		case 0:
			gps.GeoPoint = spatial.NewEuclideanPoint(lon, lat)
		case 4326:
			gps.GeoPoint = spatial.NewWGS84Point(lon, lat)
		default:
			gps.GeoPoint = spatial.NewWGS84Point(lon, lat)
		}
	}
	return &gps
}
