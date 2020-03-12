mapboxgl.accessToken = 'pk.eyJ1IjoiZGltYWhraWluIiwiYSI6ImNqZmNqYWV3bjJxM2IzNG52M3cwNG9sbTEifQ.hBZWN6asfRuTVSKV6Ut1Bw'; // token from Mapbox docs (https://docs.mapbox.com/mapbox-gl-js/example/simple-map/)
var map = new mapboxgl.Map({
    container: "map",
    style: "mapbox://styles/mapbox/streets-v11",
    center: [37.60011784074581, 55.74694688386492],
    zoom: 17
});

var draw = new MapboxDraw({
    displayControlsDefault: false,
    userProperties: true,
    controls: {
        point: true
    },
    styles: [
        {
            'id': 'gl-draw-point-point-stroke-inactive',
            'type': 'circle',
            'filter': ['all', ['==', 'active', 'false'],
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
            'filter': ['all', ['==', 'active', 'false'],
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
            'filter': ['all', ['==', '$type', 'Point'],
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
            'filter': ['all', ['==', '$type', 'Point'],
                ['!=', 'meta', 'midpoint'],
                ['==', 'active', 'true']
            ],
            'paint': {
                'circle-radius': 5,
                'circle-color': '#fbb03b'
            }
        }
    ]
});

map.addControl(draw, "top-left");
var timerAnimatedRoute = null;

map.on("load", function() {
    console.log("Map has been loaded");
    map.on("draw.create", updateMapMatch);
    map.on("draw.delete", updateMapMatch);
    map.on("draw.update", updateMapMatch);
});

function updateMapMatch(e) {
    var data = draw.getAll();
    // draw.changeMode("draw_point");
    if (data.features.length < 3) {
        console.log(`You need to provide another ${3-data.features.length} GPS points`);
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
        "stateRadius": 7,
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
        if (this.map.getLayer(layerName)) {
            this.map.removeLayer(layerName);
        }
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
            step = (step + 1) % dashArraySeq.length;
            this.map.setPaintProperty(layerName, "line-dasharray", dashArraySeq[step]);
        }, animationStep);

    });
}

