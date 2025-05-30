## ROADMAP
New ideas, thought about needed features will be store in this file.

### Done
* Initial core
    * Work on distributions and HMM ([Hidden Markov model](https://en.wikipedia.org/wiki/Hidden_Markov_model))
    * Integration with [s2](https://github.com/golang/geo#overview) library
    * Integration with [btree](https://github.com/google/btree#btree-implementation-for-go) library
    * Integration with [viterbi](https://github.com/LdDl/viterbi#viterbi) library
    * Integration with [ch](https://github.com/LdDl/ch) library
    * Integration with [go.geojson](https://github.com/paulmach/go.geojson#gogeojson) library

* REST server side (and store it in folder cmd/)
    * Main server application via [fasthttp](https://github.com/valyala/fasthttp#fasthttp-----)-based framework called [Fiber](https://github.com/gofiber/fiber)
    * Map matching service
    * Shortest path finder (we are trying to avoid word "routing") service
    * Isochrones service

* gRPC server side
    * generate protobuf structure
    * Map matching service
    * Isochrones service
    * gRPC docs (autogen)

* ~~Front-end integrated with server-side. Probably via [Mapbox](https://github.com/mapbox/mapbox-gl-js).~~
Replaced with [Maplibre](https://maplibre.org/) and [Maptiler](https://www.maptiler.com/) due Mapbox [changed license](https://github.com/mapbox/mapbox-gl-js/releases/tag/v2.0.0)

* More screenshots in README
* ~~Migrate to Fiber v2~~
* ~~Migrate to new version of CH (https://github.com/LdDl/ch) v1.7.5~~
* ~~Swagger docs (autogen)~~ - https://github.com/LdDl/horizon/pull/10
* ~~Snake case for JSON's~~

### W.I.P

* Stabilization of core (need many tests as possible)

* Rewrite front-end in [VueJS](https://github.com/vuejs/vue) or [Svelte](https://svelte.dev/) framework (+ update installation instruction) [AS SEPARATE REPOSITORY + design Figma]

### Planned
* Some kind of wiki
* Cool logo :) PR's are welcome, haha
* Contributing guidelines
* Think about: Add option for REST to provide GPS points without timestamp ???
* Think about: Add sort for provided GPS points to avoid client's mistake ???
* REST server side (and store it in folder cmd/)
    * Service bringing MVT tiles of graph
    * Need to integrate some good heuristics into FindShortestPath() function. Current implementation based on "nearest edge" for choosing source and target vertices.
* Front-end shortest path builder (like current map match, but just different colors? don't know)
* Front-end isochrones

### Continuous activity
* README
* Benchmarks and tests
* [Horizon](cmd/horizon) itself
* Roadmap itself
* Front-end improvements
