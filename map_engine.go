package horizon

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/LdDl/ch"
	"github.com/LdDl/horizon/spatial"
	"github.com/pkg/errors"
)

// MapEngine Engine for solving finding shortest path and KNN problems
// edges - set of edges (map[from_vertex]map[to_vertex]Edge)
// storage - spatial storage for solving KNN problem (implements spatial.Storage interface)
// vertices - datastore for graph vertices (with geometry property)
// graph - Graph(E,V). It wraps ch.Graph (see https://github.com/LdDl/ch/blob/master/graph.go#L17). It used for solving finding shortest path problem.
// queryPool - thread-safe query pool for concurrent shortest path queries (ch v1.10.0+)
// vertexComponent - matches vertex ID to its weakly connected component ID
// bigComponentID - ID of the largest weakly connected component. -1 if no components found
type MapEngine struct {
	edges     map[int64]map[int64]*spatial.Edge
	storage   spatial.Storage
	vertices  map[int64]*spatial.Vertex
	graph     ch.Graph
	queryPool *ch.QueryPool
	// Weak connected components
	vertexComponent map[int64]int64
	bigComponentID  int64
	// Strong connected components (SCC, Tarjan's algorithm)
	vertexStrongComponent map[int64]int64
	bigStrongComponentID  int64
	isComponentVerySmall  map[int64]bool
}

// NewMapEngineDefault Returns pointer to created MapEngine with default parameters
func NewMapEngineDefault() *MapEngine {
	storage := spatial.NewStorage(spatial.StorageTypeSpherical)
	return &MapEngine{
		edges:    make(map[int64]map[int64]*spatial.Edge),
		vertices: make(map[int64]*spatial.Vertex),
		storage:  storage,
	}
}

// NewMapEngine Returns pointer to created MapEngine with provided parameters
func NewMapEngine(opts ...func(*MapEngine)) *MapEngine {
	engine := &MapEngine{
		edges:    make(map[int64]map[int64]*spatial.Edge),
		vertices: make(map[int64]*spatial.Vertex),
		storage:  nil,
	}
	for _, opt := range opts {
		opt(engine)
	}
	return engine
}

// WithGraph is an option which sets graph for MapEngine
func WithGraph(graph ch.Graph) func(*MapEngine) {
	return func(engine *MapEngine) {
		shortcutsNum := graph.GetShortcutsNum()
		if shortcutsNum == 0 {
			// Prepare shortcuts to speed up shortest path calculations
			graph.PrepareContractionHierarchies()
		}
		engine.graph = graph
		// Initialize thread-safe query pool for concurrent shortest path queries
		engine.queryPool = graph.NewQueryPool()
	}
}

// WithStorage is an option which sets storage for MapEngine
func WithStorage(storage spatial.Storage) func(*MapEngine) {
	return func(engine *MapEngine) {
		engine.storage = storage
	}
}

// WithS2Storage is an option which sets s2Storage for MapEngine (backward compatibility)
// Deprecated: Use WithStorage instead
func WithS2Storage(storage *spatial.S2Storage) func(*MapEngine) {
	return func(engine *MapEngine) {
		engine.storage = storage
	}
}

// WithEdges is an option which sets edges for MapEngine and populate existing spatial index
// This option requires that storage is already set in MapEngine which you can do via WithStorage option
// If storage is nil then edges are just set without populating spatial index!!!
func WithEdges(edges []*spatial.Edge) func(*MapEngine) {
	return func(engine *MapEngine) {
		for _, edge := range edges {
			if engine.edges[edge.Source] == nil {
				engine.edges[edge.Source] = make(map[int64]*spatial.Edge)
			}
			engine.edges[edge.Source][edge.Target] = edge
			if engine.storage != nil {
				engine.storage.AddEdge(uint64(edge.ID), edge)
			}
		}
	}
}

// WithVertices is an option which sets vertices for MapEngine
func WithVertices(vertices []*spatial.Vertex) func(*MapEngine) {
	return func(engine *MapEngine) {
		for _, vertex := range vertices {
			engine.vertices[vertex.ID] = vertex
		}
	}
}

func prepareEngine(edgesFilename string) (*MapEngine, error) {
	engine := NewMapEngineDefault()

	/* Prepare filenames (output of 'osm2ch' CLI tool) */
	fnamePart := strings.Split(edgesFilename, ".csv")
	edgesFilename = fnamePart[0] + ".csv"
	verticesFilename := fnamePart[0] + "_vertices.csv"
	shortcutsFilename := fnamePart[0] + "_shortcuts.csv"
	fmt.Printf("Extracting edges from '%s' file...\n", edgesFilename)
	st := time.Now()
	err := engine.extractDataFromCSVs(edgesFilename, verticesFilename, shortcutsFilename)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Done in %v\n", time.Since(st))
	fmt.Printf("Loading graph and preparing engine...\n")
	st = time.Now()
	fmt.Printf("Done in %v\n", time.Since(st))
	return engine, nil
}

// routeDistanceMeters sums LengthMeters of edges along a path (sequence of vertex IDs).
// Used for transition probability which needs distance in meters, not time-based cost.
func (engine *MapEngine) routeDistanceMeters(path []int64) float64 {
	if len(path) < 2 {
		return 0
	}
	totalMeters := 0.0
	for i := 0; i < len(path)-1; i++ {
		from := path[i]
		to := path[i+1]
		if targets, ok := engine.edges[from]; ok {
			if edge, ok := targets[to]; ok {
				totalMeters += edge.LengthMeters
			}
		}
	}
	return totalMeters
}

func (engine *MapEngine) extractDataFromCSVs(edgesFname, verticesFname, shortcutsFname string) error {
	// Allocate memory for edges
	engine.edges = make(map[int64]map[int64]*spatial.Edge)

	// Read edges first
	fileEdges, err := os.Open(edgesFname)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Can't open edges file '%s'", edgesFname))
	}
	defer fileEdges.Close()
	readerEdges := csv.NewReader(fileEdges)
	readerEdges.Comma = ';'

	// Read header and build column lookup
	edgesHeader, err := readerEdges.Read()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Can't read header of edges file '%s'", edgesFname))
	}
	edgesLookup := NewCSVLookup(edgesHeader)
	if err := edgesLookup.MustHave("from_vertex_id", "to_vertex_id", "geom", "edge_id"); err != nil {
		return errors.Wrap(err, fmt.Sprintf("Invalid edges file '%s'", edgesFname))
	}
	costResolver, err := NewCostResolver(edgesLookup)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Invalid edges file '%s'", edgesFname))
	}
	// Read file line by line
	for {
		record, err := readerEdges.Read()
		if err == io.EOF {
			break
		}
		sourceVertex, err := edgesLookup.Int64(record, "from_vertex_id")
		if err != nil {
			return errors.Wrap(err, "edges file")
		}
		targetVertex, err := edgesLookup.Int64(record, "to_vertex_id")
		if err != nil {
			return errors.Wrap(err, "edges file")
		}
		weight, err := costResolver.Resolve(record)
		if err != nil {
			return errors.Wrap(err, "edges file")
		}
		edgeID, err := edgesLookup.Int64(record, "edge_id")
		if err != nil {
			return errors.Wrap(err, "edges file")
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

		coordinates, err := edgesLookup.String(record, "geom")
		if err != nil {
			return errors.Wrap(err, "edges file")
		}
		s2Polyline, err := spatial.WKTToS2PolylineFeature(coordinates)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Can't parse WKT geometry of the edge: from_vertex_id = '%d' | to_vertex_id = '%d' | geom = '%s'", sourceVertex, targetVertex, coordinates))
		}
		if _, ok := engine.edges[sourceVertex]; !ok {
			engine.edges[sourceVertex] = make(map[int64]*spatial.Edge)
		}
		lengthMeters := weight // fallback: if no length_meters column, use weight
		if edgesLookup.Has("length_meters") {
			lengthMeters, err = edgesLookup.Float64(record, "length_meters")
			if err != nil {
				return errors.Wrap(err, "edges file: length_meters")
			}
		}
		edge := spatial.Edge{
			ID:           edgeID,
			Source:        sourceVertex,
			Target:        targetVertex,
			Weight:        weight,
			LengthMeters: lengthMeters,
			Polyline:     s2Polyline,
		}
		engine.edges[sourceVertex][targetVertex] = &edge

		err = engine.storage.AddEdge(uint64(edgeID), &edge)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Can't add s2-polyline to engine: from_vertex_id = '%d' | to_vertex_id = '%d' | geom = '%s'", sourceVertex, targetVertex, coordinates))
		}
	}

	/* Now prepare order position and importance of each vertex */
	/* This helps to avade graph.PrepareContractionHierarchies() call */
	// Read vertices
	fileVertices, err := os.Open(verticesFname)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Can't open vertices file '%s'", verticesFname))
	}
	defer fileVertices.Close()
	readerVertices := csv.NewReader(fileVertices)
	readerVertices.Comma = ';'

	// Read header and build column lookup
	verticesHeader, err := readerVertices.Read()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Can't read header of vertices file '%s'", verticesFname))
	}
	verticesLookup := NewCSVLookup(verticesHeader)
	if err := verticesLookup.MustHave("vertex_id", "order_pos", "importance", "geom"); err != nil {
		return errors.Wrap(err, fmt.Sprintf("Invalid vertices file '%s'", verticesFname))
	}
	// Read file line by line
	for {
		record, err := readerVertices.Read()
		if err == io.EOF {
			break
		}
		vertexExternal, err := verticesLookup.Int64(record, "vertex_id")
		if err != nil {
			return errors.Wrap(err, "vertices file")
		}
		vertexOrderPos, err := verticesLookup.Int64(record, "order_pos")
		if err != nil {
			return errors.Wrap(err, "vertices file")
		}
		vertexImportance, err := verticesLookup.Int(record, "importance")
		if err != nil {
			return errors.Wrap(err, "vertices file")
		}
		vertexInternal, vertexFound := engine.graph.FindVertex(vertexExternal)
		if !vertexFound {
			return fmt.Errorf("vertex with Label = %d is not found in graph", vertexExternal)
		}
		engine.graph.Vertices[vertexInternal].SetOrderPos(vertexOrderPos)
		engine.graph.Vertices[vertexInternal].SetImportance(vertexImportance)

		coordinates, err := verticesLookup.String(record, "geom")
		if err != nil {
			return errors.Wrap(err, "vertices file")
		}
		s2Point, err := spatial.WKTToS2PointFeature(coordinates)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Can't parse WKT geometry of the vertex '%d' | geom = '%s'", vertexExternal, coordinates))
		}
		engine.vertices[vertexExternal] = &spatial.Vertex{
			ID:    vertexExternal,
			Point: &s2Point,
		}
	}

	/* After hierarchies prepared add shortcuts to graph */
	// Read contractions
	fileShortcuts, err := os.Open(shortcutsFname)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Can't open shortcuts file '%s'", shortcutsFname))
	}
	defer fileShortcuts.Close()
	readerShortcuts := csv.NewReader(fileShortcuts)
	readerShortcuts.Comma = ';'
	// Read header and build column lookup
	shortcutsHeader, err := readerShortcuts.Read()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Can't read header of shortcuts file '%s'", shortcutsFname))
	}
	shortcutsLookup := NewCSVLookup(shortcutsHeader)
	if err := shortcutsLookup.MustHave("from_vertex_id", "to_vertex_id", "weight", "via_vertex_id"); err != nil {
		return errors.Wrap(err, fmt.Sprintf("Invalid shortcuts file '%s'", shortcutsFname))
	}
	// Read file line by line
	for {
		record, err := readerShortcuts.Read()
		if err == io.EOF {
			break
		}
		sourceExternal, err := shortcutsLookup.Int64(record, "from_vertex_id")
		if err != nil {
			return errors.Wrap(err, "shortcuts file")
		}
		targetExternal, err := shortcutsLookup.Int64(record, "to_vertex_id")
		if err != nil {
			return errors.Wrap(err, "shortcuts file")
		}
		weight, err := shortcutsLookup.Float64(record, "weight")
		if err != nil {
			return errors.Wrap(err, "shortcuts file")
		}
		contractionExternal, err := shortcutsLookup.Int64(record, "via_vertex_id")
		if err != nil {
			return errors.Wrap(err, "shortcuts file")
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

	// Finalize import after loading all vertices, edges, and shortcuts (required for ch v1.10.0+)
	engine.graph.FinalizeImport()

	// Initialize thread-safe query pool for concurrent shortest path queries
	engine.queryPool = engine.graph.NewQueryPool()

	// Compute weakly connected components for the graph
	componentsResult := engine.computeWeakConnectedComponents()
	engine.vertexComponent = componentsResult.VertexComponent
	engine.bigComponentID = componentsResult.BigComponentID

	// Compute strongly connected components for the graph via Tarjan's algorithm
	sccResult := engine.computeStrongConnectedComponents()
	engine.vertexStrongComponent = sccResult.VertexComponent
	engine.bigStrongComponentID = sccResult.BigComponentID
	engine.isComponentVerySmall = sccResult.IsComponentVerySmall

	return nil
}
