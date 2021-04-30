package horizon

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	geojson "github.com/paulmach/go.geojson"
	"github.com/pkg/errors"

	"github.com/LdDl/ch"
)

// MapEngine Engine for solving finding shortest path and KNN problems
// edges - set of edges (map[from_vertex]map[to_vertex]Edge)
// s2Storage - datastore for B-tree. It is used for solving KNN problem
// graph - Graph(E,V). It wraps ch.Graph (see https://github.com/LdDl/ch/blob/master/graph.go#L17). It used for solving finding shortest path problem.
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

func prepareEngine(edgesFilename string) (*MapEngine, error) {
	engine := NewMapEngineDefault()

	/* Prepare filenames (output of 'osm2ch' CLI tool) */
	fnamePart := strings.Split(edgesFilename, ".csv")
	edgesFilename = fnamePart[0] + ".csv"
	verticesFilename := fnamePart[0] + "_vertices.csv"
	shortcutsFilename := fnamePart[0] + "_shortcuts.csv"
	fmt.Printf("Extractiong edges from '%s' file...\n", edgesFilename)
	st := time.Now()
	err := engine.extractDataFromCSVs(edgesFilename, verticesFilename, shortcutsFilename)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Done in %v\n", time.Since(st))
	fmt.Printf("Loading graph and preparing engine... ")
	st = time.Now()
	fmt.Printf("Done in %v\n", time.Since(st))
	return engine, nil
}

func (engine *MapEngine) extractDataFromCSVs(edgesFname, verticesFname, shortcutsFname string) error {
	// Allocate memory for edges
	engine.edges = make(map[int64]map[int64]*Edge)

	// Read edges first
	fileEdges, err := os.Open(edgesFname)
	if err != nil {
		return err
	}
	defer fileEdges.Close()
	readerEdges := csv.NewReader(fileEdges)
	readerEdges.Comma = ';'

	// Fill graph with edges informations
	// Skip header of CSV-file
	_, err = readerEdges.Read()
	if err != nil {
		return err
	}
	// Read file line by line
	edgeID := int64(0)
	for {
		record, err := readerEdges.Read()
		if err == io.EOF {
			break
		}
		sourceVertex, err := strconv.ParseInt(record[0], 10, 64)
		if err != nil {
			return err
		}
		targetVertex, err := strconv.ParseInt(record[1], 10, 64)
		if err != nil {
			return err
		}
		weight, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			return err
		}
		err = engine.graph.CreateVertex(sourceVertex)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Can't add source vertex with from_vertex_id = '%d'", sourceVertex))
		}
		err = engine.graph.CreateVertex(targetVertex)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Can't add target vertex with to_vertex_id = '%d'", targetVertex))
		}
		err = engine.graph.AddEdge(sourceVertex, targetVertex, weight)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Can't add edge: from_vertex_id = '%d' | to_vertex_id = '%d'", sourceVertex, targetVertex))
		}

		coordinates := record[3]
		bytesCoordinates := []byte(coordinates)
		geojsonPolyline, err := geojson.UnmarshalGeometry(bytesCoordinates)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Can't parse GeoJSON geometry of the edge: from_vertex_id = '%d' | to_vertex_id = '%d' | geom = '%s'", sourceVertex, targetVertex, coordinates))
		}
		s2Polyline, err := GeoJSONToS2PolylineFeature(geojsonPolyline)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Can't prepare s2-polyline edge: from_vertex_id = '%d' | to_vertex_id = '%d' | geom = '%s'", sourceVertex, targetVertex, coordinates))
		}
		if _, ok := engine.edges[sourceVertex]; !ok {
			engine.edges[sourceVertex] = make(map[int64]*Edge)
		}
		edge := Edge{
			ID:       edgeID,
			Source:   sourceVertex,
			Target:   targetVertex,
			Weight:   weight,
			Polyline: s2Polyline,
		}
		engine.edges[sourceVertex][targetVertex] = &edge

		err = engine.s2Storage.AddEdge(uint64(edgeID), &edge)
		if err != nil {
			return err
		}
		edgeID++
	}

	/* Now prepare order position and importance of each vertex */
	/* This helps to avade graph.PrepareContractionHierarchies() call */
	// Read vertices
	fileVertices, err := os.Open(verticesFname)
	if err != nil {
		return err
	}
	defer fileVertices.Close()
	readerVertices := csv.NewReader(fileVertices)
	readerVertices.Comma = ';'

	// Skip header of CSV-file
	_, err = readerVertices.Read()
	if err != nil {
		return err
	}
	// Read file line by line
	for {
		record, err := readerVertices.Read()
		if err == io.EOF {
			break
		}
		vertexExternal, err := strconv.ParseInt(record[0], 10, 64)
		if err != nil {
			return err
		}
		vertexOrderPos, err := strconv.Atoi(record[1])
		if err != nil {
			return err
		}
		vertexImportance, err := strconv.Atoi(record[2])
		if err != nil {
			return err
		}
		vertexInternal, vertexFound := engine.graph.FindVertex(vertexExternal)
		if !vertexFound {
			return fmt.Errorf("Vertex with Label = %d is not found in graph", vertexExternal)
		}
		engine.graph.Vertices[vertexInternal].SetOrderPos(vertexOrderPos)
		engine.graph.Vertices[vertexInternal].SetImportance(vertexImportance)
	}

	/* After hierarchies prepared add shortcuts to graph */
	// Read contractions
	fileShortcuts, err := os.Open(shortcutsFname)
	if err != nil {
		return err
	}
	defer fileShortcuts.Close()
	readerShortcuts := csv.NewReader(fileShortcuts)
	readerShortcuts.Comma = ';'
	// Skip header of CSV-file
	_, err = readerShortcuts.Read()
	if err != nil {
		return err
	}
	// Read file line by line
	for {
		record, err := readerShortcuts.Read()
		if err == io.EOF {
			break
		}
		sourceExternal, err := strconv.ParseInt(record[0], 10, 64)
		if err != nil {
			return err
		}
		targetExternal, err := strconv.ParseInt(record[1], 10, 64)
		if err != nil {
			return err
		}
		weight, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			return err
		}
		contractionExternal, err := strconv.ParseInt(record[3], 10, 64)
		if err != nil {
			return err
		}
		err = engine.graph.AddEdge(sourceExternal, targetExternal, weight)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Can't add shortcut with source_internal_ID = '%d' and target_internal_ID = '%d'", sourceExternal, targetExternal))
		}
		err = engine.graph.AddShortcut(sourceExternal, targetExternal, contractionExternal, weight)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Can't add shortcut with source_internal_ID = '%d' and target_internal_ID = '%d' to internal map", sourceExternal, targetExternal))
		}
	}
	return nil
}
