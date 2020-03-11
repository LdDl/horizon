package main

import "flag"

var (
	addrFlag  = flag.String("h", "0.0.0.0", "Bind address")
	portFlag  = flag.Int("p", 32800, "Port")
	fileFlag  = flag.String("f", "graph.csv", "Filename of *.csv file (you can get one using https://github.com/LdDl/ch/tree/master/cmd/osm2ch#osm2ch)")
	sigmaFlag = flag.Float64("sigma", 50.0, "σ-parameter for evaluating emission probabilities")
	betaFlag  = flag.Float64("beta", 30.0, "β-parameter for evaluationg transition probabilities")
)

func main() {
	flag.Parse()
}
