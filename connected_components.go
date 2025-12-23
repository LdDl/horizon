package horizon

// WeakComponentsResult holds the result of weakly connected components evaluation.
type WeakComponentsResult struct {
	// matches each vertex ID to its component ID
	VertexComponent map[int64]int64
	// ID of the largest component
	// Will be -1 if no components found
	BigComponentID int64
	// matches component ID to number of vertices in that component
	ComponentSizes map[int64]int
	// overall number of components found
	TotalComponents int64
}

// bfsMarkWeakComponent performs BFS traversal starting from 'start' vertex
// and marks all reachable vertices with given componentID.
// Graph is treated as undirected (weakly connected components - see the ref. https://en.wikipedia.org/wiki/Weak_component)
// Returns the number of vertices in the component.
func (engine *MapEngine) bfsMarkWeakComponent(start int64, componentID int64, visited map[int64]bool, vertexComponent map[int64]int64) int {
	if visited[start] {
		return 0
	}

	queue := []int64{start}
	size := 0

	for len(queue) > 0 {
		v := queue[0]
		queue = queue[1:]

		if visited[v] {
			continue
		}
		visited[v] = true
		vertexComponent[v] = componentID
		size++

		// Add neighbors via outgoing edges (v -> neighbor)
		for neighbor := range engine.edges[v] {
			if !visited[neighbor] {
				queue = append(queue, neighbor)
			}
		}

		// Add neighbors via incoming edges (neighbor -> v)
		// We treat graph as undirected for weakly connected components
		for sourceVertex, targets := range engine.edges {
			if _, hasEdgeToV := targets[v]; hasEdgeToV {
				if !visited[sourceVertex] {
					queue = append(queue, sourceVertex)
				}
			}
		}
	}

	return size
}

// computeWeakConnectedComponents finds all weakly connected components in the graph.
// It iterates over all vertices, runs BFS from each unvisited vertex,
// and identifies the biggest component (by vertex count).
func (engine *MapEngine) computeWeakConnectedComponents() WeakComponentsResult {
	visited := make(map[int64]bool)
	vertexComponent := make(map[int64]int64)
	componentSizes := make(map[int64]int)

	var componentID int64 = 0
	var bigComponentID int64 = -1
	var bigComponentSize int = 0

	// Collect all unique vertex IDs from edges
	vertices := make(map[int64]bool)
	for src, targets := range engine.edges {
		vertices[src] = true
		for dst := range targets {
			vertices[dst] = true
		}
	}

	// Process each vertex
	for v := range vertices {
		if visited[v] {
			continue
		}

		size := engine.bfsMarkWeakComponent(v, componentID, visited, vertexComponent)
		componentSizes[componentID] = size

		// Track the biggest component
		if size > bigComponentSize {
			bigComponentSize = size
			bigComponentID = componentID
		}

		componentID++
	}

	return WeakComponentsResult{
		VertexComponent: vertexComponent,
		BigComponentID:  bigComponentID,
		ComponentSizes:  componentSizes,
		TotalComponents: componentID,
	}
}
