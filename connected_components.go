package horizon

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
