package horizon

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/LdDl/ch"
	"github.com/golang/geo/s2"
	geojson "github.com/paulmach/go.geojson"
)

// MapEngine Engine for solving finding shortest path and KNN problems
/*
	edges - set of edges (map[from_vertex]map[to_vertex]Edge)
	s2Storage - datastore for B-tree. It is used for solving KNN problem
	graph - Graph(E,V). It wraps ch.Graph (see https://github.com/LdDl/ch/blob/master/graph.go#L17). It used for solving finding shortest path problem
*/
type MapEngine struct {
	edges     map[int64]map[int64]*Edge
	s2Storage *S2Storage
	graph     ch.Graph
}

// NewMapEngineDefault Returns pointer to created MapEngine with default parameters
func NewMapEngineDefault() *MapEngine {
	index := NewS2Storage(17, 35)
	return &MapEngine{
		edges:     make(map[int64]map[int64]*Edge),
		s2Storage: index,
	}
}

// NewMapEngine Returns pointer to created MapEngine with provided parameters
/*
	storageLevel - level for S2
	degree - degree of b-tree
*/
func NewMapEngine(storageLevel int, degree int) *MapEngine {
	index := NewS2Storage(storageLevel, degree)
	return &MapEngine{
		edges:     make(map[int64]map[int64]*Edge),
		s2Storage: index,
	}
}

// prepareGraph Insertes vertices and edges into MapEngine
/*
	edges - set of edges (map[from_vertex]map[to_vertex]Edge)
*/
func (engine *MapEngine) prepareGraph(edges map[int64]map[int64]*Edge) {
	engine.edges = edges
	for i := range edges {
		engine.graph.CreateVertex(i)
		for j := range edges[i] {
			engine.graph.CreateVertex(j)
			engine.graph.AddEdge(i, j, edges[i][j].Weight)
			engine.s2Storage.AddEdge(uint64(edges[i][j].ID), edges[i][j])
		}
	}
}

func prepareEngine(graphFileName string) (*MapEngine, error) {
	engine := NewMapEngineDefault()
	fmt.Printf("Extractiong edges from '%s' file... ", graphFileName)
	st := time.Now()
	edges, err := extractEdgesFromCSV(graphFileName)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Done in %v\n", time.Since(st))
	fmt.Printf("Preparing graph... ")
	st = time.Now()
	engine.prepareGraph(edges)
	fmt.Printf("Done in %v\n", time.Since(st))
	fmt.Printf("Preparing contracts... ")
	st = time.Now()
	engine.graph.PrepareContracts()
	fmt.Printf("Done in %v\n", time.Since(st))
	return engine, nil
}

func extractEdgesFromCSV(fname string) (map[int64]map[int64]*Edge, error) {
	file, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader := csv.NewReader(bufio.NewReader(file))
	reader.Comma = ';'
	reader.LazyQuotes = true
	_, err = reader.Read()
	if err != nil {
		return nil, err
	}

	ans := make(map[int64]map[int64]*Edge)
	edgeID := int64(0)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}

		source, err := strconv.ParseInt(record[0], 10, 64)
		if err != nil {
			return nil, err
		}
		target, err := strconv.ParseInt(record[1], 10, 64)
		if err != nil {
			return nil, err
		}
		if source == target {
			continue
		}

		oneway := record[2]
		weight, err := strconv.ParseFloat(record[3], 64)
		if err != nil {
			return nil, err
		}

		coordinates := record[4]
		bytesCoordinates := []byte(coordinates)
		geojsonPolyline, err := geojson.UnmarshalGeometry(bytesCoordinates)
		if err != nil {
			return nil, err
		}

		s2Polyline, err := GeoJSONToS2PolylineFeature(geojsonPolyline)
		if err != nil {
			return nil, err
		}

		if _, ok := ans[source]; !ok {
			ans[source] = make(map[int64]*Edge)
			ans[source][target] = &Edge{
				ID:       edgeID,
				Source:   source,
				Target:   target,
				Weight:   weight,
				Polyline: s2Polyline,
			}
		} else {
			ans[source][target] = &Edge{
				ID:       edgeID,
				Source:   source,
				Target:   target,
				Weight:   weight,
				Polyline: s2Polyline,
			}
		}

		if oneway == "B" {
			reverseS2Polyline := make(s2.Polyline, len(*s2Polyline))
			copy(reverseS2Polyline, *s2Polyline)
			reverseS2Polyline.Reverse()

			edgeID++
			if _, ok := ans[target]; !ok {
				ans[target] = make(map[int64]*Edge)
				ans[target][source] = &Edge{
					ID:       edgeID,
					Source:   target,
					Target:   source,
					Weight:   weight,
					Polyline: &reverseS2Polyline,
				}
			} else {
				ans[target][source] = &Edge{
					ID:       edgeID,
					Source:   target,
					Target:   source,
					Weight:   weight,
					Polyline: &reverseS2Polyline,
				}
			}
		}
		edgeID++
	}
	return ans, nil
}
