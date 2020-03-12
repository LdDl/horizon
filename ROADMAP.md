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

### W.I.P
* REST server side (and store it in folder cmd/)
    * Shortest path finder (we are trying to avoid word "routing") service

### Planned
* Stable core (need many tests as possible)
* Front-end integrated with server-side. Probably via [Mapbox](https://github.com/mapbox/mapbox-gl-js).
* Rewrite front-end on [VueJS](https://github.com/vuejs/vue) framework (+ update installation instruction)
* gRPC server side (and store it in folder cmd/) with same features as REST
* Some kind of wiki
* Cool logo :) PR's are welcome, haha
* Contributing guidelines
* Think about: Add option for REST to provide GPS points without timestamp ???
* Think about: Add sort for provided GPS points to avoid client's mistake ???
* REST server side (and store it in folder cmd/)
    * Service bringing MVT tiles of graph
    * Isochrones
    
### Continuous activity
* README
* Benchmarks and tests
* [Horizon](cmd/horizon) itself
* Roadmap itself