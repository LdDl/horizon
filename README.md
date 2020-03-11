[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://godoc.org/github.com/LdDl/horizon)
[![Build Status](https://travis-ci.com/LdDl/horizon.svg?branch=master)](https://travis-ci.com/LdDl/horizon)
[![Sourcegraph](https://sourcegraph.com/github.com/LdDl/horizon/-/badge.svg)](https://sourcegraph.com/github.com/LdDl/horizon?badge)
[![Go Report Card](https://goreportcard.com/badge/github.com/LdDl/horizon)](https://goreportcard.com/report/github.com/LdDl/horizon)
[![GitHub tag](https://img.shields.io/github/tag/LdDl/horizon.svg)](https://github.com/LdDl/horizon/releases)

# Horizon v0.1.0 (W.I.P)
Horizon is project aimed to do map matching, routing.

## Table of Contents
- [About](#about)
- [Installation](#installation)
- [Usage](#usage)
- [Benchmark](#benchmark)
- [Support](#support)
- [ToDo](#todo)
- [Theory](#theory)
- [Dependencies](#dependencies)
- [License](#license)

## About
Horizon is targeted to make map matching as [OSRM](https://github.com/Project-OSRM/osrm-backend) / [Graphopper](https://github.com/graphhopper/graphhopper) or [Valhala](https://github.com/valhalla/valhalla) have done, but in Go ecosystem.


## Installation
```shell
go get github.com/LdDl/ch
```

## Usage
@todo

## Benchmark
Please follow [link](BENCHMARK.md)

## Support
If you have troubles or questions please [open an issue](https://github.com/LdDl/ch/issues/new).
Feel free to make PR's (we do not have contributing guidelines currently, but we will someday)

## ToDo
Please see [ROADMAP.md](ROADMAP.md)

## Theory
Thanks for approach described in this paper:
**Newson, Paul, and John Krumm. "Hidden Markov map matching through noise and sparseness." Proceedings of the 17th ACM SIGSPATIAL International Conference on Advances in Geographic Information Systems. ACM, 2009**

[Hidden Markov model](https://en.wikipedia.org/wiki/Hidden_Markov_model) is used as backbone for preparing probabities for Viterbi algorithm. Notice that we do not use 'classical' [Normal distribution](https://en.wikipedia.org/wiki/Normal_distribution) for evaluating emission probabilty or [Exponential distribution](https://en.wikipedia.org/wiki/Exponential_distribution) for evaluatuin transition probabilties in HMM. Instead of it we use **Log-normal distribution** for emissions and **Log-exponential distribution** for transitions. Why is that? Because we do not want to get underflow (arithmetic) for small probabilities

[Viterbi algorithm](https://en.wikipedia.org/wiki/Viterbi_algorithm) is used to evaluate the most suitable trace of GPS track.

## Dependencies
* Contraction hierarchies library with bidirectional Dijkstra's algorithm - [ch](https://github.com/LdDl/ch#ch---contraction-hierarchies). License is Apache-2.0
* Viterbi's algorithm implementation - [viterbi](https://github.com/LdDl/viterbi#viterbi). License is Apache-2.0
* S2 (spherical geometry) library - [s2](https://github.com/golang/geo#overview). License is Apache-2.0
* Btree implementation - [btree](https://github.com/google/btree#btree-implementation-for-go). License is Apache-2.0
* GeoJSON stuff - [go.geojson](https://github.com/paulmach/go.geojson#gogeojson). License is MIT

## License
You can check it [here](https://github.com/LdDl/horizon/blob/master/LICENSE)

