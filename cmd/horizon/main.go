package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/LdDl/horizon"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	geojson "github.com/paulmach/go.geojson"
	"github.com/valyala/fasthttp"
)

var (
	addrFlag        = flag.String("h", "0.0.0.0", "Bind address")
	portFlag        = flag.Int("p", 32800, "Port")
	fileFlag        = flag.String("f", "graph.csv", "Filename of *.csv file (you can get one using https://github.com/LdDl/osm2ch#osm2ch)")
	sigmaFlag       = flag.Float64("sigma", 50.0, "σ-parameter for evaluating emission probabilities")
	betaFlag        = flag.Float64("beta", 30.0, "β-parameter for evaluating transition probabilities")
	lonFlag         = flag.Float64("maplon", 0.0, "initial longitude of front-end map")
	latFlag         = flag.Float64("maplat", 0.0, "initial latitude of front-end map")
	zoomFlag        = flag.Float64("mapzoom", 1.0, "initial zoom of front-end map")
	timestampLayout = "2006-01-02T15:04:05"
)

func main() {
	flag.Parse()

	// Init web page
	webPage = fmt.Sprintf(webPage, *lonFlag, *latFlag, *zoomFlag)

	// Init map matcher engine
	hmmParams := horizon.NewHmmProbabilities(*sigmaFlag, *betaFlag)
	matcher, err := horizon.NewMapMatcher(hmmParams, *fileFlag)
	if err != nil {
		log.Panicln(err)
	}

	config := fiber.Config{
		DisableStartupMessage: false,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			log.Println(err)
			return ctx.Status(fasthttp.StatusInternalServerError).JSON(map[string]string{"Error": "undefined"})
		},
		IdleTimeout: 10 * time.Second,
	}
	allCors := cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowHeaders:     "Origin, Authorization, Content-Type, Content-Length, Accept, Accept-Encoding, X-HttpRequest",
		AllowMethods:     "GET, POST, PUT, DELETE",
		ExposeHeaders:    "Content-Length",
		AllowCredentials: true,
		MaxAge:           5600,
	})

	// Init server
	server := fiber.New(config)
	server.Use(allCors)
	server.Get("/", RenderPage())
	api := server.Group("api")
	v010 := api.Group("/v0.1.0")

	v010.Post("/mapmatch", MapMatch(matcher))
	v010.Post("/shortest", FindSP(matcher))
	v010.Post("/isochrones", FindIsochrones(matcher))

	// Start server
	server.Listen(fmt.Sprintf("%s:%d", *addrFlag, *portFlag))
}

// RenderPage Render front-end
func RenderPage() func(*fiber.Ctx) error {
	fn := func(ctx *fiber.Ctx) error {
		ctx.Set("Content-Type", "text/html")
		// ctx.Fasthttp.Request.Header.Set("Content-type", "text/html")
		return ctx.SendString(webPage)
	}
	return fn
}

// H Just alias to map[stirng]string
type H map[string]string

// Request User's request
type Request struct {
	// Set of GPS data
	Data []RequestDatum `json:"gps"`
	// Max number of states for single GPS point (in range [1, 10], default is 5). Field would be ignored for request on '/shortest' service.
	MaxStates *int `json:"maxStates"`
	// Max radius of search for potential candidates (in range [7, 50], default is 25.0)
	StateRadius *float64 `json:"stateRadius"`
}

// RequestDatum Single row
type RequestDatum struct {
	// Timestamp. Field would be ignored for request on '/shortest' service.
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

// MapMatch Do map match via POST-request
func MapMatch(matcher *horizon.MapMatcher) func(*fiber.Ctx) error {
	fn := func(ctx *fiber.Ctx) error {

		bodyBytes := ctx.Context().PostBody()
		data := Request{}
		err := json.Unmarshal(bodyBytes, &data)
		if err != nil {
			return ctx.Status(400).JSON(H{"Error": err.Error()})
		}

		if len(data.Data) < 3 {
			return ctx.Status(400).JSON(H{"Error": "Please provide 3 GPS points atleast"})
		}

		gpsMeasurements := horizon.GPSMeasurements{}
		for i := range data.Data {
			tm, err := time.Parse(timestampLayout, data.Data[i].Timestamp)
			if err != nil {
				return ctx.Status(400).JSON(H{"Error": "Wrong timestamp layout. Please use YYYY-MM-DDTHH:mm:SS"})
			}
			gpsMeasurement := horizon.NewGPSMeasurement(tm, data.Data[i].LonLat[0], data.Data[i].LonLat[1], 4326)
			gpsMeasurements = append(gpsMeasurements, gpsMeasurement)
		}

		statesRadiusMeters := 25.0
		maxStates := 5
		ans := Response{}

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
			ctx.SendStatus(500)
			ctx.JSON(H{"Error": "Something went wrong on server side"})
		}

		ans.Path = geojson.NewFeatureCollection()
		f := horizon.S2PolylineToGeoJSONFeature(&result.Path)
		ans.Path.AddFeature(f)

		return ctx.Status(200).JSON(ans)
	}
	return fn
}

// FindSP Find shortest path via POST-request
/*
   Actually it can be done just by doing MapMatch for 2 proided points, but this just proof of concept
   Services takes two points, snaps those to nearest vertices and finding path via Dijkstra's algorithm. Output is familiar to MapMatch()
*/
func FindSP(matcher *horizon.MapMatcher) func(*fiber.Ctx) error {
	fn := func(ctx *fiber.Ctx) error {
		bodyBytes := ctx.Context().PostBody()
		data := Request{}
		err := json.Unmarshal(bodyBytes, &data)
		if err != nil {
			return ctx.Status(400).JSON(H{"Error": err.Error()})
		}
		if len(data.Data) != 2 {
			return ctx.Status(400).JSON(H{"Error": "Please provide 2 GPS points only"})
		}

		gpsMeasurements := horizon.GPSMeasurements{}
		ut := time.Now().UTC().Unix()
		for i := range data.Data {
			gpsMeasurement := horizon.NewGPSMeasurementFromID(int(ut), data.Data[i].LonLat[0], data.Data[i].LonLat[1], 4326)
			gpsMeasurements = append(gpsMeasurements, gpsMeasurement)
			ut++
		}

		statesRadiusMeters := 25.0
		ans := Response{}

		if data.StateRadius != nil && *data.StateRadius >= 7 && *data.StateRadius <= 50 {
			statesRadiusMeters = *data.StateRadius
		} else {
			ans.Warnings = append(ans.Warnings, "stateRadius either nil or not in range [7,50]. Using default value: 25.0")
		}

		result, err := matcher.FindShortestPath(gpsMeasurements[0], gpsMeasurements[1], statesRadiusMeters)
		if err != nil {
			log.Println(err)
			ctx.SendStatus(500)
			ctx.JSON(H{"Error": "Something went wrong on server side"})
		}
		ans.Path = geojson.NewFeatureCollection()
		f := horizon.S2PolylineToGeoJSONFeature(&result.Path)
		ans.Path.AddFeature(f)

		return ctx.Status(200).JSON(ans)
	}
	return fn
}

// IsochronesRequest User's request for isochrones
type IsochronesRequest struct {
	// [Longitude, Latitude]
	LonLat [2]float64 `json:"lonLat"`
	// Max cost restrictions for single isochrone. Should be in range [0,+Inf]. Minumim is 0.
	MaxCost *float64 `json:"maxCost"`
	// Max radius of search for nearest vertex (Optional, default is 25.0, should be in range [0,+Inf])
	MaxNearestRadius *float64 `json:"nearestRadius"`
}

// FindIsochrones Find possible isochrones via POST-request
func FindIsochrones(matcher *horizon.MapMatcher) func(*fiber.Ctx) error {
	fn := func(ctx *fiber.Ctx) error {
		bodyBytes := ctx.Context().PostBody()
		data := IsochronesRequest{}
		err := json.Unmarshal(bodyBytes, &data)
		if err != nil {
			return ctx.Status(400).JSON(H{"Error": err.Error()})
		}

		gpsMeasurement := horizon.NewGPSMeasurementFromID(0, data.LonLat[0], data.LonLat[1], 4326)
		maxCost := 0.0
		ans := Response{}
		if data.MaxCost != nil && *data.MaxCost >= 0 {
			maxCost = *data.MaxCost
		} else {
			ans.Warnings = append(ans.Warnings, "maxCost either nil or not in range [0,+Inf]. Using default value: 0.0")
		}

		maxNearestRadius := 25.0
		if data.MaxNearestRadius != nil && *data.MaxNearestRadius >= 0 {
			maxNearestRadius = *data.MaxNearestRadius
		} else {
			ans.Warnings = append(ans.Warnings, "nearestRadius either nil or not in range [0,+Inf]. Using default value: 0.0")
		}

		result, err := matcher.FindIsochrones(gpsMeasurement, maxCost, maxNearestRadius)
		if err != nil {
			log.Println(err)
			ctx.SendStatus(500)
			ctx.JSON(H{"Error": "Something went wrong on server side"})
		}
		_ = result
		return ctx.Status(200).JSON("{\"status\": \"w.i.p\"}")
	}
	return fn
}

var (
	webPage = `
	<!DOCTYPE html>
<html>
    <head>
        <meta charset="utf-8" />
        <title>Horizon client</title>
        <meta name="viewport" content="initial-scale=1,maximum-scale=1,user-scalable=no" />
        <script src="https://api.mapbox.com/mapbox-gl-js/v1.8.1/mapbox-gl.js"></script>
        <link href="https://api.mapbox.com/mapbox-gl-js/v1.8.1/mapbox-gl.css" rel="stylesheet" />
        <script src='https://cdnjs.cloudflare.com/ajax/libs/moment.js/2.24.0/moment-with-locales.min.js'></script>
        <style>
            body {
                margin: 0;
                padding: 0;
            }
            #map {
                position: absolute;
                top: 0;
                bottom: 0;
                width: 100%%;
            }
            #icons {
                position: absolute;
                bottom: 0;
                margin: 0 0 8px 140px;
            }
        </style>
    </head>
    <body>
        <script src='https://api.mapbox.com/mapbox-gl-js/plugins/mapbox-gl-draw/v1.1.0/mapbox-gl-draw.js'></script>
        <link rel='stylesheet' href='https://api.mapbox.com/mapbox-gl-js/plugins/mapbox-gl-draw/v1.1.0/mapbox-gl-draw.css' type='text/css' />
        <div id="map"></div>
        <script>
            mapboxgl.accessToken = 'pk.eyJ1IjoiZGltYWhraWluIiwiYSI6ImNqZmNqYWV3bjJxM2IzNG52M3cwNG9sbTEifQ.hBZWN6asfRuTVSKV6Ut1Bw'; // token from Mapbox docs (https://docs.mapbox.com/mapbox-gl-js/example/simple-map/)
            var map = new mapboxgl.Map({
                container: "map",
                style: "mapbox://styles/dimahkiin/ck7q21t6z0ny71imt9v5valra",
                center: [%f, %f],
                zoom: %f
            });

            var textFieldProps = {
                'type': 'identity',
                'property': 'num'
            };

            const theme = [
                {
                    'id': 'gl-draw-polygon-fill-inactive',
                    'type': 'fill',
                    'filter': ['all',
                    ['==', 'active', 'false'],
                    ['==', '$type', 'Polygon'],
                    ['!=', 'mode', 'static']
                    ],
                    'paint': {
                    'fill-color': '#3bb2d0',
                    'fill-outline-color': '#3bb2d0',
                    'fill-opacity': 0.1
                    }
                },
                {
                    'id': 'gl-draw-polygon-fill-active',
                    'type': 'fill',
                    'filter': ['all', ['==', 'active', 'true'], ['==', '$type', 'Polygon']],
                    'paint': {
                    'fill-color': '#fbb03b',
                    'fill-outline-color': '#fbb03b',
                    'fill-opacity': 0.1
                    }
                },
                {
                    'id': 'gl-draw-polygon-midpoint',
                    'type': 'circle',
                    'filter': ['all',
                    ['==', '$type', 'Point'],
                    ['==', 'meta', 'midpoint']],
                    'paint': {
                    'circle-radius': 3,
                    'circle-color': '#fbb03b'
                    }
                },
                {
                    'id': 'gl-draw-polygon-stroke-inactive',
                    'type': 'line',
                    'filter': ['all',
                    ['==', 'active', 'false'],
                    ['==', '$type', 'Polygon'],
                    ['!=', 'mode', 'static']
                    ],
                    'layout': {
                    'line-cap': 'round',
                    'line-join': 'round'
                    },
                    'paint': {
                    'line-color': '#3bb2d0',
                    'line-width': 2
                    }
                },
                {
                    'id': 'gl-draw-polygon-stroke-active',
                    'type': 'line',
                    'filter': ['all', ['==', 'active', 'true'], ['==', '$type', 'Polygon']],
                    'layout': {
                    'line-cap': 'round',
                    'line-join': 'round'
                    },
                    'paint': {
                    'line-color': '#fbb03b',
                    'line-dasharray': [0.2, 2],
                    'line-width': 2
                    }
                },
                {
                    'id': 'gl-draw-line-inactive',
                    'type': 'line',
                    'filter': ['all',
                    ['==', 'active', 'false'],
                    ['==', '$type', 'LineString'],
                    ['!=', 'mode', 'static']
                    ],
                    'layout': {
                    'line-cap': 'round',
                    'line-join': 'round'
                    },
                    'paint': {
                    'line-color': '#3bb2d0',
                    'line-width': 2
                    }
                },
                {
                    'id': 'gl-draw-line-active',
                    'type': 'line',
                    'filter': ['all',
                    ['==', '$type', 'LineString'],
                    ['==', 'active', 'true']
                    ],
                    'layout': {
                    'line-cap': 'round',
                    'line-join': 'round'
                    },
                    'paint': {
                    'line-color': '#fbb03b',
                    'line-dasharray': [0.2, 2],
                    'line-width': 2
                    }
                },
                {
                    'id': 'gl-draw-polygon-and-line-vertex-stroke-inactive',
                    'type': 'circle',
                    'filter': ['all',
                    ['==', 'meta', 'vertex'],
                    ['==', '$type', 'Point'],
                    ['!=', 'mode', 'static']
                    ],
                    'paint': {
                    'circle-radius': 5,
                    'circle-color': '#fff'
                    }
                },
                {
                    'id': 'gl-draw-polygon-and-line-vertex-inactive',
                    'type': 'circle',
                    'filter': ['all',
                    ['==', 'meta', 'vertex'],
                    ['==', '$type', 'Point'],
                    ['!=', 'mode', 'static']
                    ],
                    'paint': {
                    'circle-radius': 3,
                    'circle-color': '#fbb03b'
                    }
                },
                {
                    'id': 'gl-draw-point-point-stroke-inactive',
                    'type': 'circle',
                    'filter': ['all',
                    ['==', 'active', 'false'],
                    ['==', '$type', 'Point'],
                    ['==', 'meta', 'feature'],
                    ['!=', 'mode', 'static']
                    ],
                    'paint': {
                    'circle-radius': 5,
                    'circle-opacity': 1,
                    'circle-color': '#fff'
                    }
                },
                {
                    'id': 'gl-draw-point-inactive',
                    'type': 'circle',
                    'filter': ['all',
                    ['==', 'active', 'false'],
                    ['==', '$type', 'Point'],
                    ['==', 'meta', 'feature'],
                    ['!=', 'mode', 'static']
                    ],
                    'paint': {
                    'circle-radius': 3,
                    'circle-color': '#3bb2d0'
                    }
                },
                {
                    'id': 'gl-draw-point-stroke-active',
                    'type': 'circle',
                    'filter': ['all',
                    ['==', '$type', 'Point'],
                    ['==', 'active', 'true'],
                    ['!=', 'meta', 'midpoint']
                    ],
                    'paint': {
                    'circle-radius': 7,
                    'circle-color': '#fff'
                    }
                },
                {
                    'id': 'gl-draw-point-active',
                    'type': 'circle',
                    'filter': ['all',
                    ['==', '$type', 'Point'],
                    ['!=', 'meta', 'midpoint'],
                    ['==', 'active', 'true']],
                    'paint': {
                    'circle-radius': 5,
                    'circle-color': '#fbb03b'
                    }
                },
                {
                    'id': 'gl-draw-polygon-fill-static',
                    'type': 'fill',
                    'filter': ['all', ['==', 'mode', 'static'], ['==', '$type', 'Polygon']],
                    'paint': {
                    'fill-color': '#404040',
                    'fill-outline-color': '#404040',
                    'fill-opacity': 0.1
                    }
                },
                {
                    'id': 'gl-draw-polygon-stroke-static',
                    'type': 'line',
                    'filter': ['all', ['==', 'mode', 'static'], ['==', '$type', 'Polygon']],
                    'layout': {
                    'line-cap': 'round',
                    'line-join': 'round'
                    },
                    'paint': {
                    'line-color': '#404040',
                    'line-width': 2
                    }
                },
                {
                    'id': 'gl-draw-line-static',
                    'type': 'line',
                    'filter': ['all', ['==', 'mode', 'static'], ['==', '$type', 'LineString']],
                    'layout': {
                    'line-cap': 'round',
                    'line-join': 'round'
                    },
                    'paint': {
                    'line-color': '#404040',
                    'line-width': 2
                    }
                },
                {
                    'id': 'gl-draw-point-static',
                    'type': 'circle',
                    'filter': ['all', ['==', 'mode', 'static'], ['==', '$type', 'Point']],
                    'paint': {
                    'circle-radius': 5,
                    'circle-color': '#404040'
                    }
                }
            ];

            let modifiedStyles = theme.map(function(style) {
                if (style.id === 'gl-draw-point-inactive') {
                    return carCreated(style);
                } else if (style.id === 'gl-draw-point-active') {
                    return carClicked(style);
                } else {
                    return style;
                }
            });

            function carCreated(style) {
                return {
                    id: style.id,
                    filter: style.filter,
                    type: "symbol",
                    layout: {
                        "text-field": ['get', 'id'],
                        "text-variable-anchor": ['top', 'bottom', 'left', 'right'],
                        "text-radial-offset": 1.0,
                        "text-justify": "auto",
                        "icon-image": "loc_marker_placed",
                        "icon-size": 0.5,
                        "icon-allow-overlap": true,
                        "text-allow-overlap": true
                    }
                };
            }

            function carClicked(style) {
                return {
                    id: style.id,
                    filter: style.filter,
                    type: "symbol",
                    layout: {
                        "text-field": ['get', 'id'],
                        "text-variable-anchor": ['top', 'bottom', 'left', 'right'],
                        "text-radial-offset": 1.0,
                        "text-justify": "auto",
                        "icon-image": "loc_marker",
                        "icon-size": 0.5,
                        "icon-allow-overlap": true,
                        "text-allow-overlap": true
                    }
                };
            }

            var draw = new MapboxDraw({
                displayControlsDefault: false,
                userProperties: true,
                controls: {
                    point: true,
                    trash: true
                },
                styles: modifiedStyles
            });

            map.addControl(draw, "top-left");
            var timerAnimatedRoute = null;
            var pointsCounter = 0;

            map.on("load", function() {
                console.log("Map has been loaded");
                map.on("draw.create", updateMapMatch);
                map.on("draw.update", updateMapMatch);
                map.on("draw.delete", updateMapMatch);
            });

            function updateMapMatch(e) {

                if (e.features && e.features.length === 1 && e.type === "draw.create") {
                    pointsCounter++;
                    // Tyring to play with ID
                    // Can't do text-field with properties ['get', 'property_name'] just doesn't work, when I do provide property)
                    draw.delete(e.features[0].id);
                    e.features[0].id = "GPS #" + pointsCounter.toString();
                    draw.add(e.features[0]);
                }

                var data = draw.getAll();
                if (data.features.length < 3) {
                    console.log("You need to provide another " + (3-data.features.length).toString() + " GPS points");
                    if (map.getLayer("layer_matched_route")) { // Clear layer when 'draw.delete' fired
                        map.removeLayer("layer_matched_route");
                    }
                    return
                }

                console.log("Doing map matching");
                let currentTime = new Date();
                let gpsMeasurements = data.features.map(element => {
                    currentTime.setSeconds(currentTime.getSeconds() + 30); // artificial GPS timestamps
                    return {
                        "tm": moment(currentTime).format("YYYY-MM-DDTh:mm:ss"),
                        "lonLat": [element.geometry.coordinates[0], element.geometry.coordinates[1]],
                    };
                });
                doMapMatch(gpsMeasurements)
            }

            function doMapMatch(gpsMeasurements) {
                
                let requestData = {
                    "maxStates": 5,
                    "stateRadius": 50,
                    "gps": gpsMeasurements
                }
                let sourceName = "source_matched_route";
                let layerName = "layer_matched_route";

                fetch("http://localhost:32800/api/v0.1.0/mapmatch", {
                    method: "post",
                    headers: {
                        'Accept': 'application/json',
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify(requestData)
                })
                .then(response => response.json())
                .then(function(jsoned) {

                    clearInterval(timerAnimatedRoute);

                    if (map.getSource(sourceName)) {
                        map.getSource(sourceName).setData(jsoned.data);
                    } else {
                        map.addSource(sourceName, {
                            "type": "geojson",
                            "data": jsoned.data
                        });
                    }
                    if (!map.getLayer(layerName)) {
                        map.addLayer({
                            "id": layerName,
                            "type": "line",
                            "source": sourceName,
                            "layout": {
                                "line-join": "round",
                                "line-cap": "butt"
                            },
                            "paint": {
                                "line-color": "#0000ff",
                                "line-opacity": 0.8 ,
                                "line-dasharray": [0, 4, 3],
                                "line-width": 3
                            }
                        });
                    }

                    // Animation - https://stackoverflow.com/a/45817976/6026885
                    let step = 0;
                    let dashArraySeq = [
                    [0, 4, 3],
                    [1, 4, 2],
                    [2, 4, 1],
                    [3, 4, 0],
                    [0, 1, 3, 3],
                    [0, 2, 3, 2],
                    [0, 3, 3, 1]
                    ];
                    let animationStep = 100;
                    timerAnimatedRoute = setInterval(() => {
                        step = (step + 1) %% dashArraySeq.length;
                        if (map.getLayer(layerName)) {
                            map.setPaintProperty(layerName, "line-dasharray", dashArraySeq[step]);
                        }
                    }, animationStep);

                });
            }
        </script>
        <div id="icons">Icons made by <a href="https://www.flaticon.com/authors/freepik" title="Freepik">Freepik</a> from <a href="https://www.flaticon.com/" title="Flaticon">www.flaticon.com</a></div>
    </body>
</html>
	`
)
