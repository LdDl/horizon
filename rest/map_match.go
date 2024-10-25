package rest

import (
	"encoding/json"
	"log"
	"time"

	"github.com/LdDl/horizon"
	"github.com/gofiber/fiber/v2"
	"github.com/golang/geo/s2"
	geojson "github.com/paulmach/go.geojson"
)

var timestampLayout = "2006-01-02T15:04:05"

// MapMatchRequest User's request for map matching
// swagger:model
type MapMatchRequest struct {
	// Max number of states for single GPS point (in range [1, 10], default is 5). Field would be ignored for request on '/shortest' service.
	MaxStates *int `json:"max_states" example:"5"`
	// Max radius of search for potential candidates (in range [7, 50], default is 25.0)
	StateRadius *float64 `json:"state_radius" example:"7.0"`
	// Set of GPS data
	Data []GPSToMapMatch `json:"gps"`
}

// GPSToMapMatch Representation of GPS data
// swagger:model
type GPSToMapMatch struct {
	// Timestamp. Field would be ignored for request on '/shortest' service.
	Timestamp string `json:"tm" example:"2020-03-11T00:00:00"`
	// [Longitude, Latitude]
	LonLat [2]float64 `json:"lon_lat" example:"37.601249363208915,55.745374309126895"`
}

// MapMatchResponse Server's response for map matching request
// swagger:model
type MapMatchResponse struct {
	// GeoJSON data
	// Set of matched edges for each observation
	Data []ObservationEdgeResponse `json:"data" `
	// Warnings
	Warnings []string `json:"warnings" example:"Warning"`
}

// IntermediateEdgeResponse Edge which is not matched to any observation but helps to form whole travel path
// swagger:model
type IntermediateEdgeResponse struct {
	// Edge geometry as GeoJSON LineString feature
	Geom *geojson.Feature `json:"geom" swaggertype:"object"`
	// Travel cost
	Weight float64 `json:"weight"`
	// Edge identifier
	ID int64 `json:"id" example:"4278"`
}

// Relation between observation and matched edge
type ObservationEdgeResponse struct {
	// Index of an observation. Index correspondes to index in incoming request. If some indices are not presented then it means that they have been trimmed
	ObservationIdx int `json:"obs_idx" example:"0"`
	// Matched edge identifier
	EdgeID int64 `json:"edge_id" example:"3149"`
	// Matched vertex identifier
	VertexID int64 `json:"vertex_id" example:"44014"`
	// Corresponding matched edge as GeoJSON LineString feature
	MatchedEdge *geojson.Feature `json:"matched_edge" swaggertype:"object"`
	// Cut for excess part of the matched edge. Will be null for every observation except the first and the last. Could be null for first/last edge when projection point corresponds to source/target vertices of the edge
	MatchedEdgeCut *geojson.Feature `json:"matched_edge_cut" swaggertype:"object"`
	// Corresponding matched vertex as GeoJSON Point feature
	MatchedVertex *geojson.Feature `json:"matched_vertex" swaggertype:"object"`
	// Corresponding projection on the edge as GeoJSON Point feature
	ProjectedPoint *geojson.Feature `json:"projected_point" swaggertype:"object"`
	// Set of leading edges up to next observation (so these edges is not matched to any observation explicitly). Could be an empty array if observations are very close to each other or if it just last observation
	NextEdges []IntermediateEdgeResponse `json:"next_edges"`
}

// MapMatch Do map match via POST-request
// @Summary Do map match via POST-request
// @Tags Map matching
// @Produce json
// @Param POST-body body rest.MapMatchRequest true "Example of request"
// @Success 200 {object} rest.MapMatchResponse
// @Failure 424 {object} codes.Error424
// @Failure 500 {object} codes.Error500
// @Router /api/v0.1.0/mapmatch [POST]
func MapMatch(matcher *horizon.MapMatcher) func(*fiber.Ctx) error {
	fn := func(ctx *fiber.Ctx) error {
		bodyBytes := ctx.Context().PostBody()
		data := MapMatchRequest{}
		err := json.Unmarshal(bodyBytes, &data)
		if err != nil {
			return ctx.Status(400).JSON(fiber.Map{"Error": err.Error()})
		}
		if len(data.Data) < 3 {
			return ctx.Status(400).JSON(fiber.Map{"Error": "Please provide 3 GPS points atleast"})
		}
		gpsMeasurements := horizon.GPSMeasurements{}
		for i := range data.Data {
			tm, err := time.Parse(timestampLayout, data.Data[i].Timestamp)
			if err != nil {
				return ctx.Status(400).JSON(fiber.Map{"Error": "Wrong timestamp layout. Please use YYYY-MM-DDTHH:mm:SS"})
			}
			// Use index of measurement as ID
			gpsMeasurement := horizon.NewGPSMeasurement(i, data.Data[i].LonLat[0], data.Data[i].LonLat[1], 4326, horizon.WithGPSTime(tm))
			gpsMeasurements = append(gpsMeasurements, gpsMeasurement)
		}
		statesRadiusMeters := 25.0
		maxStates := 5
		ans := MapMatchResponse{}
		if data.MaxStates != nil && *data.MaxStates > 0 && *data.MaxStates < 10 {
			maxStates = *data.MaxStates
		} else {
			ans.Warnings = append(ans.Warnings, "max_states either nil or not in range [1,10]. Using default value: 5")
		}
		if data.StateRadius != nil && *data.StateRadius >= 7 && *data.StateRadius <= 50 {
			statesRadiusMeters = *data.StateRadius
		} else {
			ans.Warnings = append(ans.Warnings, "state_radius either nil or not in range [7,50]. Using default value: 25.0")
		}
		result, err := matcher.Run(gpsMeasurements, statesRadiusMeters, maxStates)
		if err != nil {
			log.Println(err)
			return ctx.Status(500).JSON(fiber.Map{"Error": "Something went wrong on server side"})
		}
		ans.Data = make([]ObservationEdgeResponse, len(result.Observations))
		for i := range result.Observations {
			observationResult := result.Observations[i]
			matchedEdgePolyline := *observationResult.MatchedEdge.Polyline
			var matchedEdgeCut s2.Polyline
			if i == 0 {
				matchedEdgePolyline, matchedEdgeCut = extractCutUpTo(*observationResult.MatchedEdge.Polyline, observationResult.ProjectedPoint, observationResult.ProjectionPointIdx)
			} else if i == len(result.Observations)-1 {
				matchedEdgePolyline, matchedEdgeCut = extractCutUpFrom(*observationResult.MatchedEdge.Polyline, observationResult.ProjectedPoint, observationResult.ProjectionPointIdx)
			}
			ans.Data[i] = ObservationEdgeResponse{
				ObservationIdx: observationResult.Observation.ID(),
				EdgeID:         observationResult.MatchedEdge.ID,
				MatchedEdge:    horizon.S2PolylineToGeoJSONFeature(matchedEdgePolyline),
				MatchedVertex:  horizon.S2PointToGeoJSONFeature(observationResult.MatchedVertex.Point),
				ProjectedPoint: horizon.S2PointToGeoJSONFeature(&observationResult.ProjectedPoint),
				NextEdges:      make([]IntermediateEdgeResponse, len(observationResult.NextEdges)),
			}
			if len(matchedEdgeCut) > 0 {
				ans.Data[i].MatchedEdgeCut = horizon.S2PolylineToGeoJSONFeature(matchedEdgeCut)
			}
			for j := range observationResult.NextEdges {
				ans.Data[i].NextEdges[j] = IntermediateEdgeResponse{
					Geom:   horizon.S2PolylineToGeoJSONFeature(observationResult.NextEdges[j].Geom),
					Weight: observationResult.NextEdges[j].Weight,
					ID:     observationResult.NextEdges[j].ID,
				}
			}
		}
		return ctx.Status(200).JSON(ans)
	}
	return fn
}

// Cuts geometry between very first point and neighbor of the projected point index in the polyline
func extractCutUpTo(polyline s2.Polyline, projected s2.Point, projectedIdx int) (s2.Polyline, s2.Polyline) {
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

// Cuts geometry between neighbor of the projected point index in the polyline and last point
func extractCutUpFrom(polyline s2.Polyline, projected s2.Point, projectedIdx int) (s2.Polyline, s2.Polyline) {
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
