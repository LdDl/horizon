mapboxgl.accessToken = 'pk.eyJ1IjoiZGltYWhraWluIiwiYSI6ImNqZmNqYWV3bjJxM2IzNG52M3cwNG9sbTEifQ.hBZWN6asfRuTVSKV6Ut1Bw'; // token from Mapbox docs (https://docs.mapbox.com/mapbox-gl-js/example/simple-map/)
var map = new mapboxgl.Map({
    container: 'map',
    style: 'mapbox://styles/mapbox/streets-v11',
    center: [0, 0],
    zoom: 1
});

map.on('load', function() {
    console.log("Map has been loaded");
});