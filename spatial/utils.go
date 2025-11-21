package spatial

import (
	"fmt"
	"math"

	"github.com/golang/geo/s2"
	geojson "github.com/paulmach/go.geojson"
)

// CalcProjection Returns projection on line and fraction for point (spherical geometry)
/*
	line - s2.Polyline
	point - s2.Point

	projected - projection of point on line
	fraction - number in [0;1], describes how far projected point from first point of polyline
	next - index of the next vertex after the projected point
*/
func CalcProjection(line s2.Polyline, point s2.Point) (projected s2.Point, fraction float64, next int) {
	pr, next := line.Project(point)
	subs := s2.Polyline{}
	for i := 0; i < next; i++ {
		subs = append(subs, line[i])
	}
	subs = append(subs, pr)
	return pr, (subs.Length() / line.Length()).Radians(), next
}

// CalcProjectionEuclidean Returns projection on line and fraction for point (Euclidean/planar geometry)
/*
	line - s2.Polyline (using Vector.X/Y as Euclidean coordinates)
	point - s2.Point (using Vector.X/Y as Euclidean coordinates)

	projected - projection of point on line
	fraction - number in [0;1], describes how far projected point from first point of polyline
	next - index of the next vertex after the projected point
*/
func CalcProjectionEuclidean(line s2.Polyline, point s2.Point) (projected s2.Point, fraction float64, next int) {
	if len(line) < 2 {
		return point, 0, 0
	}

	px, py := point.Vector.X, point.Vector.Y
	minDist := -1.0
	var bestProjection s2.Point
	bestNext := 1

	// Find the closest segment
	for i := 0; i < len(line)-1; i++ {
		x1, y1 := line[i].Vector.X, line[i].Vector.Y
		x2, y2 := line[i+1].Vector.X, line[i+1].Vector.Y

		dx := x2 - x1
		dy := y2 - y1

		var projX, projY float64
		if dx == 0 && dy == 0 {
			projX, projY = x1, y1
		} else {
			t := ((px-x1)*dx + (py-y1)*dy) / (dx*dx + dy*dy)
			if t < 0 {
				projX, projY = x1, y1
			} else if t > 1 {
				projX, projY = x2, y2
			} else {
				projX = x1 + t*dx
				projY = y1 + t*dy
			}
		}

		distSq := (px-projX)*(px-projX) + (py-projY)*(py-projY)
		if minDist < 0 || distSq < minDist {
			minDist = distSq
			bestProjection = NewEuclideanS2Point(projX, projY)
			bestNext = i + 1
		}
	}

	// Calculate total line length and length up to projection
	totalLength := 0.0
	lengthToProj := 0.0

	for i := 0; i < len(line)-1; i++ {
		dx := line[i+1].Vector.X - line[i].Vector.X
		dy := line[i+1].Vector.Y - line[i].Vector.Y
		segLen := math.Sqrt(dx*dx + dy*dy)
		totalLength += segLen

		if i < bestNext-1 {
			lengthToProj += segLen
		} else if i == bestNext-1 {
			// Add partial segment
			dx := bestProjection.Vector.X - line[i].Vector.X
			dy := bestProjection.Vector.Y - line[i].Vector.Y
			lengthToProj += math.Sqrt(dx*dx + dy*dy)
		}
	}

	if totalLength == 0 {
		return bestProjection, 0, bestNext
	}

	return bestProjection, lengthToProj / totalLength, bestNext
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
func S2PolylineToGeoJSONFeature(pts s2.Polyline) *geojson.Feature {
	coordinates := make([][]float64, len(pts))
	for i := range pts {
		latLng := s2.LatLngFromPoint(pts[i])
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

// ExtractCutUpTo cuts geometry between very first point and neighbor of the projected point index in the polyline
func ExtractCutUpTo(polyline s2.Polyline, projected s2.Point, projectedIdx int) (s2.Polyline, s2.Polyline) {
	polyCopy := polyline
	polyCopyCut := polyline

	// Cut segment from the start of the polyline up to projection poit
	polyCopy = append(s2.Polyline{projected}, polyCopy[projectedIdx:]...)

	// Cut segment from projection point up to the end of the polyline
	part := polyCopyCut[:projectedIdx-1]
	if len(part) == 0 {
		polyCopyCut = s2.Polyline{polyCopyCut[0], projected}
	} else {
		polyCopyCut = append(polyCopyCut[:projectedIdx-1], projected)
	}
	return polyCopy, polyCopyCut
}

// ExtractCutUpFrom cuts geometry between neighbor of the projected point index in the polyline and last point
func ExtractCutUpFrom(polyline s2.Polyline, projected s2.Point, projectedIdx int) (s2.Polyline, s2.Polyline) {
	polyCopy := polyline
	polyCopyCut := polyline

	// Cut segment from the projection poit up to the end of the polyline
	part := polyCopy[:projectedIdx-1]
	if len(part) == 0 {
		polyCopy = s2.Polyline{polyCopy[0], projected}
	} else {
		polyCopy = append(polyCopy[:projectedIdx-1], projected)
	}

	// Cut segment from the start of the polyline up to projection poit
	polyCopyCut = append(s2.Polyline{projected}, polyCopyCut[projectedIdx:]...)
	return polyCopy, polyCopyCut
}
