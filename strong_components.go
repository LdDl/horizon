package horizon

// Threshold for determining if a component is really small
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

// strongConnect is the recursive DFS function for Tarjan's algorithm
func (engine *MapEngine) strongConnect(v int64, state *tarjanState) {
	// Set the depth index for v to the smallest unused index
	state.indexMap[v] = state.index
	state.lowLink[v] = state.index
	state.index++
	state.stack = append(state.stack, v)
	state.onStack[v] = true

	// Consider successors of v (only outgoing edges for SCC)
	for neighbor := range engine.edges[v] {
		if _, visited := state.indexMap[neighbor]; !visited {
			// Successor has not yet been visited; recurse on it
			engine.strongConnect(neighbor, state)
			if state.lowLink[neighbor] < state.lowLink[v] {
				state.lowLink[v] = state.lowLink[neighbor]
			}
		} else if state.onStack[neighbor] {
			// Successor is on stack and hence in the current SCC
			if state.indexMap[neighbor] < state.lowLink[v] {
				state.lowLink[v] = state.indexMap[neighbor]
			}
		}
	}

	// If v is a root node, pop the stack and generate an SCC
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

// computeStrongConnectedComponents finds all strongly connected components using Tarjan's algorithm.
// A strongly connected component is a maximal set of vertices where every vertex
// can reach every other vertex via directed paths. See the ref. https://en.wikipedia.org/wiki/Tarjan%27s_strongly_connected_components_algorithm
func (engine *MapEngine) computeStrongConnectedComponents() StrongComponentsResult {
	state := newTarjanState()

	// Collect all vertices
	vertices := make(map[int64]bool)
	for src, targets := range engine.edges {
		vertices[src] = true
		for dst := range targets {
			vertices[dst] = true
		}
	}

	// Run Tarjan's algorithm from each unvisited vertex
	for v := range vertices {
		if _, visited := state.indexMap[v]; !visited {
			engine.strongConnect(v, state)
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
