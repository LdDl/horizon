package rest

import (
	"encoding/json"
	"log"
	"time"

	"github.com/LdDl/horizon"
	"github.com/gofiber/fiber/v2"
	geojson "github.com/paulmach/go.geojson"
)

var (
	timestampLayout = "2006-01-02T15:04:05"
)

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
	// GeoJSON Data
	Path *geojson.FeatureCollection `json:"data" swaggerignore:"true"`
	// Warnings
	Warnings []string `json:"warnings" example:"Warning"`
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
		ans.Path = geojson.NewFeatureCollection()
		f := horizon.S2PolylineToGeoJSONFeature(result.Path)
		ans.Path.AddFeature(f)
		return ctx.Status(200).JSON(ans)
	}
	return fn
}
