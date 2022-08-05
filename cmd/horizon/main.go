package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/LdDl/horizon"
	"github.com/LdDl/horizon/rest"
	"github.com/LdDl/horizon/rest/docs"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

var (
	addrFlag   = flag.String("h", "0.0.0.0", "Bind address")
	portFlag   = flag.Int("p", 32800, "Port")
	fileFlag   = flag.String("f", "graph.csv", "Filename of *.csv file (you can get one using https://github.com/LdDl/osm2ch#osm2ch)")
	sigmaFlag  = flag.Float64("sigma", 50.0, "σ-parameter for evaluating emission probabilities")
	betaFlag   = flag.Float64("beta", 30.0, "β-parameter for evaluating transition probabilities")
	lonFlag    = flag.Float64("maplon", 0.0, "initial longitude of front-end map")
	latFlag    = flag.Float64("maplat", 0.0, "initial latitude of front-end map")
	zoomFlag   = flag.Float64("mapzoom", 1.0, "initial zoom of front-end map")
	apiPath    = "api"
	apiVersion = "0.1.0"
)

// @title API for working with Horizon
// @version 0.1.0

// @contact.name API support
// @contact.url https://github.com/LdDl/horizon#table-of-contents
// @contact.email sexykdi@gmail.com

// @BasePath /

// @schemes http https
func main() {
	flag.Parse()

	// Init web page
	webPage = fmt.Sprintf(webPage, *lonFlag, *latFlag, *zoomFlag)

	// Init map matcher engine
	hmmParams := horizon.NewHmmProbabilities(*sigmaFlag, *betaFlag)
	matcher, err := horizon.NewMapMatcher(hmmParams, *fileFlag)
	if err != nil {
		fmt.Println(err)
		return
	}

	config := fiber.Config{
		DisableStartupMessage: false,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			log.Println("error:", err)
			return ctx.Status(500).JSON(map[string]string{"Error": "undefined"})
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
	server.Get("/", rest.RenderPage(webPage))
	apiGroup := server.Group(apiPath)
	apiVersionGroup := apiGroup.Group(fmt.Sprintf("/v%s", apiVersion))

	apiVersionGroup.Post("/mapmatch", rest.MapMatch(matcher))
	apiVersionGroup.Post("/shortest", rest.FindSP(matcher))
	apiVersionGroup.Post("/isochrones", rest.FindIsochrones(matcher))

	docsStaticGroup := apiVersionGroup.Group("/docs-static")
	docsStaticGroup.Use("/", docs.PrepareStaticAssets())

	docsGroup := apiVersionGroup.Group("/docs")
	docsGroup.Use("/", docs.PrepareStaticPage())

	// Start server
	if err := server.Listen(fmt.Sprintf("%s:%d", *addrFlag, *portFlag)); err != nil {
		fmt.Println(err)
		return
	}
}

var (
	webPage = `
<!DOCTYPE html>
<html>
    <head>
        <meta charset="utf-8" />
        <title>Horizon client</title>
        <meta name="viewport" content="initial-scale=1,maximum-scale=1,user-scalable=no" />
        <script src='https://unpkg.com/maplibre-gl@2.1.9/dist/maplibre-gl.js'></script>
        <link href='https://unpkg.com/maplibre-gl@2.1.9/dist/maplibre-gl.css' rel='stylesheet' />
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
            maplibregl.setRTLTextPlugin('https://cdn.maptiler.com/mapbox-gl-js/plugins/mapbox-gl-rtl-text/v0.2.3/mapbox-gl-rtl-text.js');
            var map = window.map = new maplibregl.Map({
                container: 'map',
                center: [%f, %f],
                zoom: %f,
                style: 'https://api.maptiler.com/maps/bff50186-2623-47cc-9108-f6d014566cbf/style.json?key=dznzK4GQ1Lj5U7XsI22j'
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
                        "icon-size": 1.0,
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
                        "icon-size": 1.0,
                        "icon-allow-overlap": true,
                        "text-allow-overlap": true
                    }
                };
            }

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

            var draw = new MapboxDraw({
                displayControlsDefault: false,
                userProperties: true,
                controls: {
                    point: true,
                    trash: true
                },
                styles: modifiedStyles
            });
            
            map.addControl(new maplibregl.NavigationControl());
            map.addControl(draw, "top-left");
            var timerAnimatedRoute = null;
            var pointsCounter = 0;

            map.on("load", function() {
                console.log("Map has been loaded");
                // let imgMarker = new Image(64, 64);
                // imgMarker.src = 'loc_marker.svg';
                // map.addImage('loc_marker',  imgMarker, {pixelRatio: window.devicePixelRatio});
                // let imgMarkerPlaced = new Image(64, 64);
                // imgMarkerPlaced.src = 'loc_marker_placed.svg';
                // map.addImage('loc_marker_placed',  imgMarkerPlaced, {pixelRatio: window.devicePixelRatio});
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
