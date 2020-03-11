package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/LdDl/horizon"
	"github.com/gofiber/fiber"
	geojson "github.com/paulmach/go.geojson"
)

var (
	addrFlag        = flag.String("h", "0.0.0.0", "Bind address")
	portFlag        = flag.Int("p", 32800, "Port")
	fileFlag        = flag.String("f", "graph.csv", "Filename of *.csv file (you can get one using https://github.com/LdDl/ch/tree/master/cmd/osm2ch#osm2ch)")
	sigmaFlag       = flag.Float64("sigma", 50.0, "σ-parameter for evaluating emission probabilities")
	betaFlag        = flag.Float64("beta", 30.0, "β-parameter for evaluationg transition probabilities")
	timestampLayout = "2006-01-02T15:04:05"
)

func main() {
	flag.Parse()

	// Init map matcher engine
	hmmParams := horizon.NewHmmProbabilities(*sigmaFlag, *betaFlag)
	matcher, err := horizon.NewMapMatcher(hmmParams, *fileFlag)
	if err != nil {
		log.Panicln(err)
	}

	// Init server
	server := fiber.New()
	api := server.Group("api")
	v010 := api.Group("/v0.1.0")
	v010.Post("/mapmatch", MapMatch(matcher))

	// Start server
	server.Listen(fmt.Sprintf("%s:%d", *addrFlag, *portFlag))
}

// H Just alias to map[stirng]string
type H map[string]string

// Request User's request
type Request struct {
	// Set of GPS data
	Data []RequestDatum `json:"gps"`
	// Max number of states for single GPS point (in range [1, 10], default is 5)
	MaxStates *int `json:"maxStates"`
	// Max radius of search for potential candidates (in range [7, 15], default is 7.0)
	StateRadius *float64 `json:"stateRadius"`
}

// RequestDatum Single row
type RequestDatum struct {
	// Timestamp
	Timestamp string `json:"tm"`
	// [Longitude, Latitude]
	LonLat [2]float64 `json:"lonLat"`
}

// Response Server's response
type Response struct {
	Path *geojson.FeatureCollection `json:"data"`
	// Warnings
	Warnings []string `json:"warnings"`
}

// MapMatch Do map match via GET-request
func MapMatch(matcher *horizon.MapMatcher) func(*fiber.Ctx) {
	fn := func(ctx *fiber.Ctx) {

		bodyBytes := ctx.Fasthttp.PostBody()
		data := Request{}
		err := json.Unmarshal(bodyBytes, &data)
		if err != nil {
			ctx.SendStatus(400)
			ctx.JSON(H{"Error": err.Error()})
			return
		}

		if len(data.Data) < 3 {
			ctx.SendStatus(400)
			ctx.JSON(H{"Error": "Please provide 3 GPS points atleast"})
			return
		}

		gpsMeasurements := horizon.GPSMeasurements{}
		for i := range data.Data {
			tm, err := time.Parse(timestampLayout, data.Data[i].Timestamp)
			if err != nil {
				ctx.SendStatus(400)
				ctx.JSON(H{"Error": "Wrong timestamp layout. Please use YYYY-MM-DDTHH:mm:SS"})
				return
			}
			gpsMeasurement := horizon.NewGPSMeasurement(tm, data.Data[i].LonLat[0], data.Data[i].LonLat[1], 4326)
			gpsMeasurements = append(gpsMeasurements, gpsMeasurement)
		}

		statesRadiusMeters := 7.0
		maxStates := 5
		ans := Response{}

		if data.MaxStates != nil && *data.MaxStates > 0 && *data.MaxStates < 10 {
			maxStates = *data.MaxStates
		} else {
			ans.Warnings = append(ans.Warnings, "maxStates either nil or not in range [1,10]. Using default value: 5")
		}
		if data.MaxStates != nil && *data.StateRadius >= 7 && *data.StateRadius <= 15 {
			statesRadiusMeters = *data.StateRadius
		} else {
			ans.Warnings = append(ans.Warnings, "stateRadius either nil or not in range [7,16]. Using default value: 7.0")
		}

		result, err := matcher.Run(gpsMeasurements, statesRadiusMeters, maxStates)
		if err != nil {
			log.Println(err)
			ctx.SendStatus(500)
			ctx.JSON(H{"Error": "Something went wrong on server side"})
		}

		ans.Path = geojson.NewFeatureCollection()
		f := horizon.S2PolylineToGeoJSONFeature(&result.Path)
		ans.Path.AddFeature(f)

		ctx.JSON(ans)
	}
	return fn
}
