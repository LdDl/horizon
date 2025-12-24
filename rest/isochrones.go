package rest

import (
	"encoding/json"
	"log"

	"github.com/LdDl/horizon"
	"github.com/LdDl/horizon/spatial"
	"github.com/gofiber/fiber/v2"
	geojson "github.com/paulmach/go.geojson"
)

// IsochronesRequest User's request for isochrones
// swagger:model
type IsochronesRequest struct {
	// Max cost restrictions for single isochrone. Should be >= 0.
	MaxCost *float64 `json:"max_cost" example:"2100.0"`
	// Max radius of search for nearest vertex.
	// Use -1 for no limit, 0 for default (100m), or positive value.
	MaxNearestRadius *float64 `json:"nearest_radius" example:"100.0"`
	// [Longitude, Latitude]
	LonLat [2]float64 `json:"lon_lat" example:"37.601249363208915,55.745374309126895"`
}

// IsochronesResponse Server's response for isochrones request
// swagger:model
type IsochronesResponse struct {
	// GeojSON data with properties on each feature: "cost" - travel cost to reach the vertex; "vertex_id" - corresponding vertex
	Isochrones *geojson.FeatureCollection `json:"data" swaggerignore:"true"`
	// Warnings
	Warnings []string `json:"warnings" example:"Warning"`
}

// FindIsochrones Find possible isochrones via POST-request
// @Summary Find possible isochrones via POST-request
// @Tags Isochrones
// @Produce json
// @Param POST-body body rest.IsochronesRequest true "Example of request"
// @Success 200 {object} rest.IsochronesResponse
// @Failure 424 {object} codes.Error424
// @Failure 500 {object} codes.Error500
// @Router /api/v0.1.0/isochrones [POST]
func FindIsochrones(matcher *horizon.MapMatcher) func(*fiber.Ctx) error {
	fn := func(ctx *fiber.Ctx) error {
		bodyBytes := ctx.Context().PostBody()
		data := IsochronesRequest{}
		err := json.Unmarshal(bodyBytes, &data)
		if err != nil {
			return ctx.Status(400).JSON(fiber.Map{"Error": err.Error()})
		}
		gpsMeasurement := horizon.NewGPSMeasurementFromID(0, data.LonLat[0], data.LonLat[1], 4326)
		maxCost := 0.0
		ans := IsochronesResponse{}
		if data.MaxCost != nil && *data.MaxCost >= 0 {
			maxCost = *data.MaxCost
		} else if data.MaxCost != nil {
			ans.Warnings = append(ans.Warnings, "max_cost should be >= 0. Using default value: 0.0")
		}
		maxNearestRadius := horizon.ResolveRadius(data.MaxNearestRadius, horizon.DEFAULT_SP_RADIUS)
		result, err := matcher.FindIsochrones(gpsMeasurement, maxCost, maxNearestRadius)
		if err != nil {
			log.Println(err)
			return ctx.Status(500).JSON(fiber.Map{"Error": "Something went wrong on server side"})
		}
		ans.Isochrones = geojson.NewFeatureCollection()
		for i := range result {
			isochrone := result[i]
			if isochrone.Vertex == nil {
				log.Println("Empty vertex")
				ctx.Status(500).JSON("Empty vertex")
			}
			f := spatial.S2PointToGeoJSONFeature(isochrone.Vertex.Point)
			f.ID = i
			f.Properties["cost"] = isochrone.Cost
			f.Properties["vertex_id"] = isochrone.Vertex.ID
			ans.Isochrones.AddFeature(f)
		}
		return ctx.Status(200).JSON(ans)
	}
	return fn
}
