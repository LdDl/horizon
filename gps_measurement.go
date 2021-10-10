package horizon

import (
	"time"

	geojson "github.com/paulmach/go.geojson"
)

// GPSMeasurements Set of telematic data
type GPSMeasurements []*GPSMeasurement

// GPSMeasurement Representation of telematic data
/*
	id - unique identifier
	dateTime - timestamp
	GeoPoint - latitude(Y)/longitude(X), pointer to GeoPoint (wrapper)
*/
type GPSMeasurement struct {
	id       int
	dateTime time.Time
	*GeoPoint
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
	t - time.Time (will be converted to UnixTimestamp and used for unique identifier)
	lon - longitude (X for SRID = 0)
	lat - latitude (Y for SRID = 0)
	srid - SRID (see https://en.wikipedia.org/wiki/Spatial_reference_system), if not provided then SRID(4326) is used. 0 and 4326 are supported.
*/
func NewGPSMeasurement(t time.Time, lon, lat float64, srid ...int) *GPSMeasurement {
	gps := GPSMeasurement{
		dateTime: t,
		id:       int(t.Unix()),
	}
	if len(srid) != 0 {
		switch srid[0] {
		case 0:
			gps.GeoPoint = NewEuclideanPoint(lon, lat)
		case 4326:
			gps.GeoPoint = NewWGS84Point(lon, lat)
		default:
			gps.GeoPoint = NewWGS84Point(lon, lat)
		}
	}
	return &gps
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
			gps.GeoPoint = NewEuclideanPoint(lon, lat)
		case 4326:
			gps.GeoPoint = NewWGS84Point(lon, lat)
		default:
			gps.GeoPoint = NewWGS84Point(lon, lat)
		}
	}
	return &gps
}

// GeoJSON Returns GeoJSON representation of GeoPoint
func (gp *GeoPoint) GeoJSON() *geojson.Feature {
	return S2PointToGeoJSONFeature(&gp.Point)
}
