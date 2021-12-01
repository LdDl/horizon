package rest

import (
	"encoding/json"
	"log"
	"time"

	"github.com/LdDl/horizon"
	"github.com/gofiber/fiber/v2"
	geojson "github.com/paulmach/go.geojson"
)

// SPRequest User's request for finding shortest path
type SPRequest struct {
	// Set of GPS data
	Data []RequestDatum `json:"gps"`
	// Max radius of search for potential candidates (in range [7, 50], default is 25.0)
	StateRadius *float64 `json:"stateRadius"`
}

// SPResponse Server's response for shortest path request
type SPResponse struct {
	Path *geojson.FeatureCollection `json:"data"`
	// Warnings
	Warnings []string `json:"warnings"`
}

// FindSP Find shortest path via POST-request
/*
   Actually it can be done just by doing MapMatch for 2 proided points, but this just proof of concept
   Services takes two points, snaps those to nearest vertices and finding path via Dijkstra's algorithm. Output is familiar to MapMatch()
*/
func FindSP(matcher *horizon.MapMatcher) func(*fiber.Ctx) error {
	fn := func(ctx *fiber.Ctx) error {
		bodyBytes := ctx.Context().PostBody()
		data := SPRequest{}
		err := json.Unmarshal(bodyBytes, &data)
		if err != nil {
			return ctx.Status(400).JSON(fiber.Map{"Error": err.Error()})
		}
		if len(data.Data) != 2 {
			return ctx.Status(400).JSON(fiber.Map{"Error": "Please provide 2 GPS points only"})
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
