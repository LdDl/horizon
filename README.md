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
go get -u github.com/LdDl/horizon/...
go install github.com/LdDl/horizon/...
```
Check if **horizon** binary was installed properly:
```shell
horizon -h
```
Output should be:
```shell
  -f string
        Filename of *.csv file (you can get one using https://github.com/LdDl/ch/tree/master/cmd/osm2ch#osm2ch) (default "graph.csv")
  -h string
        Bind address (default "0.0.0.0")
  -p int
        Port (default 32800)
  -sigma float
        σ-parameter for evaluating emission probabilities (default 50)
  -beta float
        β-parameter for evaluationg transition probabilities (default 30)
```

## Usage
Instruction has been made for Linux mainly. For Windows or OSX the way may vary.

0. Installing Prerequisites


    * Install [osm2ch tool](https://github.com/LdDl/ch/tree/master/cmd/osm2ch#osm2ch). It's needed for converting *.osm.pbf file to CSV for proper usage in [contraction hierarchies (ch) library](https://github.com/LdDl/ch#ch---contraction-hierarchies)
        ```shell
        go install github.com/LdDl/ch/...
        ```
    * Check if **osm2ch** binary was installed properly:
        ```shell
        osm2ch -h
        ```

    * Install [osmconvert tool](https://wiki.openstreetmap.org/wiki/Osmconvert). You can follow the [link](https://wiki.openstreetmap.org/wiki/Osmconvert#Binaries) for theirs instruction.
    We advice to use this method (described in [Source](https://wiki.openstreetmap.org/wiki/Osmconvert#Source) paragraph):
        ```shell
        sudo apt install osmctools && wget -O - http://m.m.i24.cc/osmconvert.c | sudo cc -x c - -lz -O3 -o osmconvert
        ```
    * Check if **osmconvert** binary was installed properly:
        ```shell
        osmconvert -h
        ```

1. First of all (except previous step), you need to download road graph (OSM is most popular format, we guess). Notice: you must change bbox for your region.
    ```shell
    wget 'https://overpass-api.de/api/map?bbox=37.5453,55.7237,37.7252,55.7837' -O map.osm
    ```
2. Compress *.osm file via [osmconvert](https://wiki.openstreetmap.org/wiki/Osmconvert). Drop author and version tags also (those not necessary for map matching). 
    ```shell
    osmconvert map.osm --drop-author --drop-version --out-pbf -o=map.osm.pbf
    ```
3. Convert *.osm.pbf to CSV via [osm2ch](https://github.com/LdDl/ch/tree/master/cmd/osm2ch#osm2ch). Notice: osm2ch's default output geometry format is WKT and units is 'km' (kilometers). We are going to change those default values. We are going to extract only edges adapted for cars also .
    ```shell
    osm2ch --file map.osm.pbf --out map.csv --geomf geojson --units m --tags motorway,primary,primary_link,road,secondary,secondary_link,residential,tertiary,tertiary_link,unclassified,trunk,trunk_link
    ```
4. Start **horizon** server. Provide bind address, port, filename for road graph, σ and β parameters of your needs.
    ```shell
    horizon -p 32800 -h 0.0.0.0 -f map.csv -sigma 50.0 -beta 30.0
    ```

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
* Fiber framework (used for server app) - [Fiber](https://github.com/gofiber/fiber). License is MIT

## License
You can check it [here](https://github.com/LdDl/horizon/blob/master/LICENSE)

