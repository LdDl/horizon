package rest

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/LdDl/horizon"
	"github.com/gofiber/fiber/v2"
	geojson "github.com/paulmach/go.geojson"
)

// SPRequest User's request for finding shortest path
// swagger:model
type SPRequest struct {
	// Max radius of search for potential candidates (in range [7, 50], default is 25.0)
	StateRadius *float64 `json:"state_radius" example:"10.0"`
	// Set of GPS data
	Data []GPSToShortestPath `json:"gps"`
}

// GPSToShortestPath Representation of GPS data
// swagger:model
type GPSToShortestPath struct {
	// [Longitude, Latitude]
	LonLat [2]float64 `json:"lon_lat" example:"37.601249363208915,55.745374309126895"`
}

// SPResponse Server's response for shortest path request
// swagger:model
type SPResponse struct {
	// Set of matched edges for each path's edge as GeoJSON LineString objects. Each feature contains edge identifier (`id`), travel cost (`weight`) and geometry (`coordinates`)
	Data []*geojson.Feature `json:"data" swaggertype:"object"`
	// Warnings
	Warnings []string `json:"warnings" example:"Warning"`
}

// FindSP Find shortest path via POST-request
/*
   Actually it can be done just by doing MapMatch for 2 proided points, but this just proof of concept
   Services takes two points, snaps those to nearest vertices and finding path via Dijkstra's algorithm. Output is familiar to MapMatch()
*/
// @Summary Find shortest path via POST-request
// @Tags Routing
// @Produce json
// @Param POST-body body rest.SPRequest true "Example of request"
// @Success 200 {object} rest.SPResponse
// @Failure 424 {object} codes.Error424
// @Failure 500 {object} codes.Error500
// @Router /api/v0.1.0/shortest [POST]
func FindSP(matcher *horizon.MapMatcher) func(*fiber.Ctx) error {
	fn := func(ctx *fiber.Ctx) error {
		bodyBytes := ctx.Context().PostBody()
		data := SPRequest{}
		err := json.Unmarshal(bodyBytes, &data)
		if err != nil {
			return ctx.Status(400).JSON(fiber.Map{"Error": err.Error()})
		}
		if len(data.Data) != 2 {
			return ctx.Status(400).JSON(fiber.Map{"Error": fmt.Sprintf("please provide 2 GPS points only. Provided: %d", len(data.Data))})
		}
		gpsMeasurements := horizon.GPSMeasurements{}
		ut := time.Now().UTC().Unix()
		for i := range data.Data {
			gpsMeasurement := horizon.NewGPSMeasurementFromID(int(ut), data.Data[i].LonLat[0], data.Data[i].LonLat[1], 4326)
			gpsMeasurements = append(gpsMeasurements, gpsMeasurement)
			ut++
		}
		statesRadiusMeters := 25.0
		ans := SPResponse{}
		if data.StateRadius != nil && *data.StateRadius >= 7 && *data.StateRadius <= 50 {
			statesRadiusMeters = *data.StateRadius
		} else {
			ans.Warnings = append(ans.Warnings, "stateRadius either nil or not in range [7,50]. Using default value: 25.0")
		}
		result, err := matcher.FindShortestPath(gpsMeasurements[0], gpsMeasurements[1], statesRadiusMeters)
		if err != nil {
			return ctx.Status(500).JSON(fiber.Map{"Error": "Something went wrong on server side"})
		}

		// @todo: For now, we only handle the first sub-match
		// Do we need to handle multiple sub-matches at all? Shortest path should exists...
		subMatch := result.SubMatches[0]
		for i := range subMatch.Observations {
			observationResult := subMatch.Observations[i]
			feature := horizon.S2PolylineToGeoJSONFeature(*observationResult.MatchedEdge.Polyline)
			feature.ID = observationResult.MatchedEdge.ID
			feature.SetProperty("weight", observationResult.MatchedEdge.Weight)
			ans.Data = append(ans.Data, feature)
			for j := range observationResult.NextEdges {
				edgeFeature := horizon.S2PolylineToGeoJSONFeature(observationResult.NextEdges[j].Geom)
				edgeFeature.ID = observationResult.NextEdges[j].ID
				edgeFeature.SetProperty("weight", observationResult.NextEdges[j].Weight)
				ans.Data = append(ans.Data, edgeFeature)
			}
		}
		return ctx.Status(200).JSON(ans)
	}
	return fn
}
