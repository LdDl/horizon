package horizon

import (
	"github.com/LdDl/ch"
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

// PrepareGraph Insertes vertices and edges into MapEngine
/*
	edges - set of edges (map[from_vertex]map[to_vertex]Edge)
*/
func (engine *MapEngine) PrepareGraph(edges map[int64]map[int64]*Edge) {
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
