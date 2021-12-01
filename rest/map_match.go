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
type MapMatchRequest struct {
	// Set of GPS data
	Data []GPSToMapMatch `json:"gps"`
	// Max number of states for single GPS point (in range [1, 10], default is 5). Field would be ignored for request on '/shortest' service.
	MaxStates *int `json:"maxStates"`
	// Max radius of search for potential candidates (in range [7, 50], default is 25.0)
	StateRadius *float64 `json:"stateRadius"`
}

// GPSToMapMatch Representation of GPS data
type GPSToMapMatch struct {
	// Timestamp. Field would be ignored for request on '/shortest' service.
	Timestamp string `json:"tm"`
	// [Longitude, Latitude]
	LonLat [2]float64 `json:"lonLat"`
}

// MapMatchResponse Server's response for map matching request
type MapMatchResponse struct {
	Path *geojson.FeatureCollection `json:"data"`
	// Warnings
	Warnings []string `json:"warnings"`
}

// MapMatch Do map match via POST-request
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
			gpsMeasurement := horizon.NewGPSMeasurement(tm, data.Data[i].LonLat[0], data.Data[i].LonLat[1], 4326)
			gpsMeasurements = append(gpsMeasurements, gpsMeasurement)
		}

		statesRadiusMeters := 25.0
		maxStates := 5
		ans := MapMatchResponse{}

		if data.MaxStates != nil && *data.MaxStates > 0 && *data.MaxStates < 10 {
			maxStates = *data.MaxStates
		} else {
			ans.Warnings = append(ans.Warnings, "maxStates either nil or not in range [1,10]. Using default value: 5")
		}
		if data.StateRadius != nil && *data.StateRadius >= 7 && *data.StateRadius <= 50 {
			statesRadiusMeters = *data.StateRadius
		} else {
			ans.Warnings = append(ans.Warnings, "stateRadius either nil or not in range [7,50]. Using default value: 25.0")
		}

		result, err := matcher.Run(gpsMeasurements, statesRadiusMeters, maxStates)
		if err != nil {
			log.Println(err)
			return ctx.Status(500).JSON(fiber.Map{"Error": "Something went wrong on server side"})
		}

		ans.Path = geojson.NewFeatureCollection()
		f := horizon.S2PolylineToGeoJSONFeature(&result.Path)
		ans.Path.AddFeature(f)

		return ctx.Status(200).JSON(ans)
	}
	return fn
}