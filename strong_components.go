package horizon

// Strongly Connected Components (SCC) are used instead of Weakly Connected Components
// for routing candidate selection. The key difference:
//
// - Weak components: treat directed graph as undirected, guarantee only that some path
//   exists (ignoring edge direction). Two vertices in the same weak component may have
//   no directed path between them.
//
// - Strong components: guarantee that every vertex can reach every other vertex via
//   directed paths. If source and target are in the same SCC, a route is guaranteed.
//
// Example: A -> B -> C (one-way chain)
//
//	Weak: {A, B, C} are in one component (connected if we ignore direction)
//	Strong: {A}, {B}, {C} are separate SCCs (no way to go from C back to A)
//	Result: routing from C to A will fail, and SCC correctly predicts this
//
// For road networks, the largest SCC typically contains the main road network where
// bidirectional travel is possible (directly or via alternative routes). Small SCCs
// represent dead-ends, one-way streets without return paths, or isolated road segments.
//
// This approach is inspired by OSRM's preprocessing strategy.
// See: https://github.com/Project-OSRM/osrm-backend/blob/master/include/util/tarjan_scc.hpp

// SMALL_COMPONENT_SIZE - components with fewer vertices are considered to be very small and deprioritized during routing.
const SMALL_COMPONENT_SIZE = 1000

// StrongComponentsResult holds the result of strongly connected components evaluation.
type StrongComponentsResult struct {
	// matches each vertex ID to its SCC component ID
	VertexComponent map[int64]int64
	// ID of the largest component (-1 if no components found)
	BigComponentID int64
	// matches component ID to number of vertices in that component
	ComponentSizes map[int64]int
	// marks whether a component is really small
	IsComponentVerySmall map[int64]bool
	// overall number of components found
	TotalComponents int64
}

// tarjanState holds the state for Tarjan's SCC algorithm
type tarjanState struct {
	index      int64
	indexMap   map[int64]int64
	lowLink    map[int64]int64
	onStack    map[int64]bool
	stack      []int64
	components [][]int64
}

// newTarjanState creates a new state for Tarjan's algorithm
func newTarjanState() *tarjanState {
	return &tarjanState{
		index:    0,
		indexMap: make(map[int64]int64),
		lowLink:  make(map[int64]int64),
		onStack:  make(map[int64]bool),
		stack:    make([]int64, 0),
	}
}

// tarjanFrame represents a single stack frame for iterative Tarjan's DFS.
// It replaces the recursive call stack: vertex being processed + position in its neighbor list.
type tarjanFrame struct {
	v   int64
	pos int // next neighbor index to process in adjacency[v]
}

// strongConnect is the iterative DFS function for Tarjan's algorithm.
// Uses an explicit call stack to avoid Go stack growth overhead on large graphs.
// adjacency is a pre-built map of vertex -> []neighbor (built once, shared across calls).
// callStack is a reusable buffer to avoid repeated allocations across calls.
func (engine *MapEngine) strongConnect(root int64, state *tarjanState, adjacency map[int64][]int64, callStack []tarjanFrame) []tarjanFrame {
	// Reset and initialize root frame
	callStack = callStack[:1]
	callStack[0] = tarjanFrame{v: root, pos: 0}
	state.indexMap[root] = state.index
	state.lowLink[root] = state.index
	state.index++
	state.stack = append(state.stack, root)
	state.onStack[root] = true

	for len(callStack) > 0 {
		frame := &callStack[len(callStack)-1]
		neighbors := adjacency[frame.v]

		if frame.pos < len(neighbors) {
			neighbor := neighbors[frame.pos]
			frame.pos++

			if _, visited := state.indexMap[neighbor]; !visited {
				// "Recurse": push new frame
				state.indexMap[neighbor] = state.index
				state.lowLink[neighbor] = state.index
				state.index++
				state.stack = append(state.stack, neighbor)
				state.onStack[neighbor] = true

				callStack = append(callStack, tarjanFrame{v: neighbor, pos: 0})
			} else if state.onStack[neighbor] {
				if state.indexMap[neighbor] < state.lowLink[frame.v] {
					state.lowLink[frame.v] = state.indexMap[neighbor]
				}
			}
		} else {
			// All neighbors processed — "return" from this frame
			v := frame.v
			callStack = callStack[:len(callStack)-1]

			// Post-order: update parent's lowLink (equivalent to after recursive call returns)
			if len(callStack) > 0 {
				parent := &callStack[len(callStack)-1]
				if state.lowLink[v] < state.lowLink[parent.v] {
					state.lowLink[parent.v] = state.lowLink[v]
				}
			}

			// If v is a root node, pop the SCC stack and generate a component
			if state.lowLink[v] == state.indexMap[v] {
				component := make([]int64, 0)
				for {
					w := state.stack[len(state.stack)-1]
					state.stack = state.stack[:len(state.stack)-1]
					state.onStack[w] = false
					component = append(component, w)
					if w == v {
						break
					}
				}
				state.components = append(state.components, component)
			}
		}
	}
}

// computeStrongConnectedComponents finds all strongly connected components using Tarjan's algorithm.
// A strongly connected component is a maximal set of vertices where every vertex
// can reach every other vertex via directed paths. See the ref. https://en.wikipedia.org/wiki/Tarjan%27s_strongly_connected_components_algorithm
func (engine *MapEngine) computeStrongConnectedComponents() StrongComponentsResult {
	state := newTarjanState()

	// Pre-build flat adjacency list (CSR-style): one allocation for all neighbor data.
	// adjacencyFlat holds all neighbor IDs concatenated; adjacency[v] is a sub-slice into it.
	// This replaces 40k+ individual slice allocations with a single flat buffer.
	totalEdges := 0
	vertices := make(map[int64]bool)
	for src, targets := range engine.edges {
		vertices[src] = true
		totalEdges += len(targets)
		for dst := range targets {
			vertices[dst] = true
		}
	}
	adjacencyFlat := make([]int64, 0, totalEdges)
	adjacency := make(map[int64][]int64, len(engine.edges))
	for src, targets := range engine.edges {
		start := len(adjacencyFlat)
		for dst := range targets {
			adjacencyFlat = append(adjacencyFlat, dst)
		}
		adjacency[src] = adjacencyFlat[start:len(adjacencyFlat):len(adjacencyFlat)]
	}

	// Run Tarjan's algorithm from each unvisited vertex
	for v := range vertices {
		if _, visited := state.indexMap[v]; !visited {
			engine.strongConnect(v, state, adjacency)
		}
	}

	// Build result
	vertexComponent := make(map[int64]int64)
	componentSizes := make(map[int64]int)
	componentIsVerySmall := make(map[int64]bool)

	var bigComponentID int64 = -1
	var bigComponentSize int = 0

	for componentID, component := range state.components {
		compID := int64(componentID)
		size := len(component)
		componentSizes[compID] = size
		componentIsVerySmall[compID] = size < SMALL_COMPONENT_SIZE

		for _, v := range component {
			vertexComponent[v] = compID
		}

		if size > bigComponentSize {
			bigComponentSize = size
			bigComponentID = compID
		}
	}

	return StrongComponentsResult{
		VertexComponent:      vertexComponent,
		BigComponentID:       bigComponentID,
		ComponentSizes:       componentSizes,
		IsComponentVerySmall: componentIsVerySmall,
		TotalComponents:      int64(len(state.components)),
	}
}
