package horizon

import "time"

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

// GPSTrack Set of telematic data
type GPSTrack []*GPSMeasurement
