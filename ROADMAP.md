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

* Front-end integrated with server-side. Probably via [Mapbox](https://github.com/mapbox/mapbox-gl-js).
* More screenshots in README
* Migrate to Fiber v2
* Migrate to new version of CH (https://github.com/LdDl/ch) v1.7.5

### W.I.P
* gRPC server side
    * generate protobuf structure_W.I.P_
    * Map matching service
    * Isochrones service

* REST server side (and store it in folder cmd/) _W.I.P._
    * w.i.p

* Stabilization of core (need many tests as possible)

### Planned
* Rewrite front-end on [VueJS](https://github.com/vuejs/vue) framework (+ update installation instruction)
* Some kind of wiki
* Cool logo :) PR's are welcome, haha
* Contributing guidelines
* Think about: Add option for REST to provide GPS points without timestamp ???
* Think about: Add sort for provided GPS points to avoid client's mistake ???
* REST server side (and store it in folder cmd/)
    * Service bringing MVT tiles of graph
    * Need to integrate some good heuristics into FindShortestPath() function. Current implementation based on "nearest edge" for choosing source and target vertices.
* Front-end shortest path builder (like current map match, but just different colors? don't know)

### Continuous activity
* README
* Benchmarks and tests
* [Horizon](cmd/horizon) itself
* Roadmap itself
* Front-end improvements
