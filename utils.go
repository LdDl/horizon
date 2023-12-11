package horizon

import (
	"fmt"
	"math"

	"github.com/golang/geo/s2"
	geojson "github.com/paulmach/go.geojson"
)

// calcProjection Returns projection on line and fraction for point
/*
	line - s2.Polyline
	point - s2.Point

	projected - projection of point on line
	fraction - number in [0;1], describes how far projected point from first point of polyline
	next - index of the next vertex after the projected point
*/
func calcProjection(line s2.Polyline, point s2.Point) (projected s2.Point, fraction float64, next int) {
	pr, next := line.Project(point)
	subs := s2.Polyline{}
	for i := 0; i < next; i++ {
		subs = append(subs, line[i])
	}
	subs = append(subs, pr)
	return pr, (subs.Length() / line.Length()).Radians(), next
}

// round Round float64
func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

// toFixed Round float64 to N decimal places
func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

// GeoJSONToS2PolylineFeature Returns *s2.Polyline representation of *geojson.Geometry (of LineString type)
func GeoJSONToS2PolylineFeature(pts *geojson.Geometry) (*s2.Polyline, error) {
	latLngs := []s2.LatLng{}
	if pts.Type == "LineString" {
		for i := range pts.LineString {
			latLng := s2.LatLngFromDegrees(pts.LineString[i][1], pts.LineString[i][0])
			latLngs = append(latLngs, latLng)
		}
	} else {
		return nil, fmt.Errorf("type of geometry is: %s. Expected: 'LineString'", pts.Type)
	}
	return s2.PolylineFromLatLngs(latLngs), nil
}

// S2PolylineToGeoJSONFeature Returns GeoJSON representation of *s2.Polyline
func S2PolylineToGeoJSONFeature(pts *s2.Polyline) *geojson.Feature {
	coordinates := make([][]float64, len(*pts))
	for i := range *pts {
		latLng := s2.LatLngFromPoint((*pts)[i])
		coordinates[i] = []float64{latLng.Lng.Degrees(), latLng.Lat.Degrees()}
	}
	return geojson.NewLineStringFeature(coordinates)
}

// GeoJSONToS2PointFeature Returns s2.Point representation of *geojson.Geometry (of Point type)
func GeoJSONToS2PointFeature(pts *geojson.Geometry) (s2.Point, error) {
	var latLng s2.LatLng
	if pts.Type == "Point" {
		latLng = s2.LatLngFromDegrees(pts.Point[1], pts.Point[0])
	} else {
		return s2.Point{}, fmt.Errorf("type of geometry is: %s. Expected: 'Point'", pts.Type)
	}
	return s2.PointFromLatLng(latLng), nil
}

// S2PointToGeoJSONFeature Returns GeoJSON representation of *s2.Point
func S2PointToGeoJSONFeature(pt *s2.Point) *geojson.Feature {
	latLng := s2.LatLngFromPoint(*pt)
	return geojson.NewPointFeature([]float64{latLng.Lng.Degrees(), latLng.Lat.Degrees()})
}
