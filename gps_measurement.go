package horizon

import "time"

// GPSMeasurement Representation of telematic data
/*
	id - unique identifier
	dateTime - timestamp
	GeoPoint - latitude(Y)/longitude(X), pointer to GeoPoint (wrapper)
*/
type GPSMeasurement struct {
	id       int64
	dateTime time.Time
	*GeoPoint
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
		id:       t.Unix(),
	}
	if len(srid) != 0 {
		switch srid[0] {
		case 0:
			gps.GeoPoint = NewEuclideanPoint(lon, lat)
			break
		case 4326:
			gps.GeoPoint = NewWGS84Point(lon, lat)
			break
		default:
			gps.GeoPoint = NewWGS84Point(lon, lat)
			break
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
		id:       dateTime.Unix(),
	}
	if len(srid) != 0 {
		switch srid[0] {
		case 0:
			gps.GeoPoint = NewEuclideanPoint(lon, lat)
			break
		case 4326:
			gps.GeoPoint = NewWGS84Point(lon, lat)
			break
		default:
			gps.GeoPoint = NewWGS84Point(lon, lat)
			break
		}
	}
	return &gps
}
