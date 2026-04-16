package horizon

import "github.com/LdDl/horizon/spatial"

// EdgeEntry represents a single outgoing edge in the adjacency list.
type EdgeEntry struct {
	Target int64
	Edge   *spatial.Edge
}

// EdgeGraph stores edges in a flat lookup map (keyed by [source, target])
// plus an adjacency list for neighbor iteration. This replaces the nested
// map[int64]map[int64]*spatial.Edge for better cache locality, fewer allocations,
// and single hash lookup instead of two.
type EdgeGraph struct {
	lookup map[[2]int64]*spatial.Edge
	adj    map[int64][]EdgeEntry
}

// NewEdgeGraph creates an empty EdgeGraph.
func NewEdgeGraph() *EdgeGraph {
	return &EdgeGraph{
		lookup: make(map[[2]int64]*spatial.Edge),
		adj:    make(map[int64][]EdgeEntry),
	}
}

// Set adds or replaces an edge from src to dst.
func (g *EdgeGraph) Set(src, dst int64, edge *spatial.Edge) {
	key := [2]int64{src, dst}
	if _, exists := g.lookup[key]; !exists {
		g.adj[src] = append(g.adj[src], EdgeEntry{Target: dst, Edge: edge})
	} else {
		// Update existing entry in adjacency list
		for i := range g.adj[src] {
			if g.adj[src][i].Target == dst {
				g.adj[src][i].Edge = edge
				break
			}
		}
	}
	g.lookup[key] = edge
}

// Get returns the edge from src to dst, or nil if not found.
func (g *EdgeGraph) Get(src, dst int64) *spatial.Edge {
	return g.lookup[[2]int64{src, dst}]
}

// Neighbors returns all outgoing edges from src.
func (g *EdgeGraph) Neighbors(src int64) []EdgeEntry {
	return g.adj[src]
}

// Adj returns the full adjacency map (for iterating all sources).
func (g *EdgeGraph) Adj() map[int64][]EdgeEntry {
	return g.adj
}

// Len returns the number of source vertices that have outgoing edges.
func (g *EdgeGraph) Len() int {
	return len(g.adj)
}
