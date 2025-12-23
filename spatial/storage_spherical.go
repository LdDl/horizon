package spatial

import (
	"container/heap"

	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
	"github.com/google/btree"
)

// S2Storage Spatial datastore
/*
	storageLevel - level for S2
	edges - map of edges
	BTree - b-tree (wraps)
*/
type S2Storage struct {
	*btree.BTree
	edges        map[uint64]*Edge
	storageLevel int
}

// NewS2Storage Returns pointer to created S2Storage
/*
	storageLevel - level for S2
	degree - degree of b-tree
*/
func NewS2Storage(storageLevel int, degree int) *S2Storage {
	return &S2Storage{
		storageLevel: storageLevel,
		BTree:        btree.New(degree),
		edges:        make(map[uint64]*Edge),
	}
}

// GetEdge Returns edge by ID from storage
func (storage *S2Storage) GetEdge(edgeID uint64) *Edge {
	return storage.edges[edgeID]
}

// indexedItem Object in datastore
type indexedItem struct {
	edgesInCell []uint64
	s2.CellID
}

// Less Method to feet b-tree
func (ii indexedItem) Less(than btree.Item) bool {
	return uint64(ii.CellID) < uint64(than.(indexedItem).CellID)
}

// AddEdge Add edge (polyline) to storage
/*
	edgeID - unique identifier
	edge - edge
*/
func (storage *S2Storage) AddEdge(edgeID uint64, edge *Edge) error {
	coverer := s2.RegionCoverer{MinLevel: storage.storageLevel, MaxLevel: storage.storageLevel}
	cells := coverer.Covering(edge.Polyline)
	for _, cell := range cells {
		ii := indexedItem{CellID: cell}
		item := storage.BTree.Get(ii)
		if item != nil {
			ii = item.(indexedItem)
		}
		ii.edgesInCell = append(ii.edgesInCell, edgeID)
		storage.BTree.ReplaceOrInsert(ii)
	}
	storage.edges[edgeID] = edge
	return nil
}

// SearchInRadiusLonLat Returns edges in radius
/*
	lon - longitude
	lat - latitude
	radius - radius of search
*/
func (storage *S2Storage) SearchInRadiusLonLat(lon, lat float64, radius float64) (map[uint64]float64, error) {
	latlng := s2.LatLngFromDegrees(lat, lon)
	cell := s2.CellFromLatLng(latlng)
	centerPoint := s2.PointFromLatLng(latlng)
	centerAngle := radius / EarthRadius
	cap := s2.CapFromCenterAngle(centerPoint, s1.Angle(centerAngle))
	rc := s2.RegionCoverer{MaxLevel: storage.storageLevel, MinLevel: storage.storageLevel}
	cu := rc.Covering(cap)
	result := make(map[uint64]float64)
	for _, cellID := range cu {
		item := storage.BTree.Get(indexedItem{CellID: cellID})
		if item != nil {
			for _, edgeID := range item.(indexedItem).edgesInCell {
				polyline := storage.edges[edgeID]
				minEdge := s2.Edge{}
				minDist := s1.ChordAngle(0)
				for i := 0; i < polyline.Polyline.NumEdges(); i++ {
					if i == 0 {
						minEdge = polyline.Polyline.Edge(0)
						minDist = cell.DistanceToEdge(minEdge.V0, minEdge.V1)
						continue
					}
					edge := polyline.Polyline.Edge(i)
					distance := cell.DistanceToEdge(edge.V0, edge.V1)
					if distance < minDist {
						minDist = distance
					}
				}
				result[edgeID] = minDist.Angle().Radians() * EarthRadius
			}
		}
	}
	return result, nil
}

// FindInRadius implements Storage interface
func (storage *S2Storage) FindInRadius(pt s2.Point, radiusMeters float64) (map[uint64]float64, error) {
	return storage.SearchInRadius(pt, radiusMeters)
}

// FindNearestInRadius implements Storage interface
func (storage *S2Storage) FindNearestInRadius(pt s2.Point, radiusMeters float64, n int) ([]NearestObject, error) {
	return storage.NearestNeighborsInRadius(pt, radiusMeters, n)
}

// SearchInRadius Returns edges in radius
/*
	pt - s2.Point
	radius - radius of search
*/
func (storage *S2Storage) SearchInRadius(pt s2.Point, radius float64) (map[uint64]float64, error) {
	cell := s2.CellFromPoint(pt)
	centerPoint := pt
	centerAngle := radius / EarthRadius
	cap := s2.CapFromCenterAngle(centerPoint, s1.Angle(centerAngle))
	rc := s2.RegionCoverer{MaxLevel: storage.storageLevel, MinLevel: storage.storageLevel}
	cu := rc.Covering(cap)
	result := make(map[uint64]float64)
	for _, cellID := range cu {
		item := storage.BTree.Get(indexedItem{CellID: cellID})
		if item != nil {
			for _, edgeID := range item.(indexedItem).edgesInCell {
				polyline := storage.edges[edgeID]
				minEdge := s2.Edge{}
				minDist := s1.ChordAngle(0)
				for i := 0; i < polyline.Polyline.NumEdges(); i++ {
					if i == 0 {
						minEdge = polyline.Polyline.Edge(0)
						minDist = cell.DistanceToEdge(minEdge.V0, minEdge.V1)
						continue
					}
					edge := polyline.Polyline.Edge(i)
					distance := cell.DistanceToEdge(edge.V0, edge.V1)
					if distance < minDist {
						minDist = distance
					}
				}
				result[edgeID] = minDist.Angle().Radians() * EarthRadius
			}
		}
	}
	return result, nil
}

// NearestObject Nearest object to given point
/*
	EdgeID - unique identifier
	DistanceTo - distance to object
*/
type NearestObject struct {
	EdgeID     uint64
	DistanceTo float64
}

// NearestNeighborsInRadius Returns edges in radius with max objects restriction (KNN)
/*
	pt - s2.Point
	radius - radius of search
	n - first N closest edges
*/
func (storage *S2Storage) NearestNeighborsInRadius(pt s2.Point, radius float64, n int) ([]NearestObject, error) {
	found, err := storage.SearchInRadius(pt, radius)
	if err != nil {
		return nil, err
	}
	h := &nearestHeap{}
	heap.Init(h)
	for k, v := range found {
		heap.Push(h, NearestObject{k, v})
	}
	l := h.Len()
	if l < n {
		n = l
	}
	ans := make([]NearestObject, n)
	for i := 0; i < n; i++ {
		ans[i] = heap.Pop(h).(NearestObject)
	}
	return ans, nil
}

// maxSearchRings is the maximum number of rings to expand during FindNearest
const maxSearchRings = 50

// FindNearest implements Storage interface using iterative cell expansion.
// Expands search from center cell outward until n edges are found.
// Uses incremental frontier expansion to avoid recalculating BFS on each ring.
func (storage *S2Storage) FindNearest(pt s2.Point, n int) ([]NearestObject, error) {
	if n <= 0 {
		return nil, nil
	}

	centerCell := s2.CellFromPoint(pt).ID().Parent(storage.storageLevel)
	cell := s2.CellFromPoint(pt)

	// Track visited cells and found edges
	visited := make(map[s2.CellID]bool)
	found := make(map[uint64]float64)

	// Frontier-based expansion: start with center cell
	frontier := []s2.CellID{centerCell}
	visited[centerCell] = true

	cellSize := storage.cellSizeMeters()

	for ring := 0; ring <= maxSearchRings && len(frontier) > 0; ring++ {
		// Process all cells in current frontier (ring)
		for _, cellID := range frontier {
			item := storage.BTree.Get(indexedItem{CellID: cellID})
			if item == nil {
				continue
			}

			for _, edgeID := range item.(indexedItem).edgesInCell {
				if _, exists := found[edgeID]; exists {
					continue
				}

				polyline := storage.edges[edgeID]
				if polyline == nil || polyline.Polyline == nil {
					continue
				}

				// Calculate minimum distance to edge
				minDist := s1.ChordAngle(0)
				for i := 0; i < polyline.Polyline.NumEdges(); i++ {
					edge := polyline.Polyline.Edge(i)
					distance := cell.DistanceToEdge(edge.V0, edge.V1)
					if i == 0 || distance < minDist {
						minDist = distance
					}
				}
				found[edgeID] = minDist.Angle().Radians() * EarthRadius
			}
		}

		// Early exit check
		if len(found) >= n && ring > 0 {
			ringRadius := cellSize * float64(ring)

			// Find minimum distance among candidates
			minFoundDist := float64(1e18)
			for _, dist := range found {
				if dist < minFoundDist {
					minFoundDist = dist
				}
			}

			// If closest edge is closer than ring boundary, we can stop
			if minFoundDist < ringRadius {
				break
			}
		}

		// Expand frontier to next ring
		// Collect all unvisited neighbors of current frontier
		nextFrontier := make([]s2.CellID, 0, len(frontier)*4)
		for _, cellID := range frontier {
			// Edge neighbors (4 cells sharing an edge)
			for _, neighbor := range cellID.EdgeNeighbors() {
				if !visited[neighbor] {
					visited[neighbor] = true
					nextFrontier = append(nextFrontier, neighbor)
				}
			}
			// Vertex neighbors (cells sharing only a vertex - corners)
			for _, neighbor := range cellID.VertexNeighbors(storage.storageLevel) {
				if !visited[neighbor] {
					visited[neighbor] = true
					nextFrontier = append(nextFrontier, neighbor)
				}
			}
		}
		frontier = nextFrontier
	}

	// Build result using heap for top-N selection
	h := &nearestHeap{}
	heap.Init(h)
	for k, v := range found {
		heap.Push(h, NearestObject{k, v})
	}

	l := h.Len()
	if l < n {
		n = l
	}

	ans := make([]NearestObject, n)
	for i := 0; i < n; i++ {
		ans[i] = heap.Pop(h).(NearestObject)
	}
	return ans, nil
}

// getCellsAtRing returns all cells at a given ring distance from center
// ring 0 = just the center cell
// ring 1 = 8 neighbors (edge + vertex neighbors)
// ring 2 = outer ring of those, etc.
func (storage *S2Storage) getCellsAtRing(center s2.CellID, ring int) []s2.CellID {
	if ring == 0 {
		return []s2.CellID{center}
	}

	// For ring N, we get all neighbors at distance N
	// Using a simple approach: get all cells within ring, subtract cells within ring-1
	cellsWithin := storage.getCellsWithinRing(center, ring)
	if ring == 1 {
		// Ring 1 is just the immediate neighbors
		return cellsWithin
	}

	cellsInner := storage.getCellsWithinRing(center, ring-1)
	innerSet := make(map[s2.CellID]bool)
	for _, c := range cellsInner {
		innerSet[c] = true
	}

	var result []s2.CellID
	for _, c := range cellsWithin {
		if !innerSet[c] {
			result = append(result, c)
		}
	}
	return result
}

// getCellsWithinRing returns all cells within ring distance (inclusive)
func (storage *S2Storage) getCellsWithinRing(center s2.CellID, ring int) []s2.CellID {
	if ring == 0 {
		return []s2.CellID{center}
	}

	visited := make(map[s2.CellID]bool)
	visited[center] = true
	current := []s2.CellID{center}

	for r := 0; r < ring; r++ {
		var next []s2.CellID
		for _, c := range current {
			// Get all 8 neighbors (4 edge + 4 vertex)
			for _, neighbor := range c.EdgeNeighbors() {
				if !visited[neighbor] {
					visited[neighbor] = true
					next = append(next, neighbor)
				}
			}
			// Vertex neighbors for corners
			for _, neighbor := range c.VertexNeighbors(storage.storageLevel) {
				if !visited[neighbor] {
					visited[neighbor] = true
					next = append(next, neighbor)
				}
			}
		}
		current = next
	}

	result := make([]s2.CellID, 0, len(visited))
	for c := range visited {
		result = append(result, c)
	}
	return result
}

// cellSizeMeters returns approximate cell size in meters at the storage level
func (storage *S2Storage) cellSizeMeters() float64 {
	// S2 cell sizes (approximate, at equator):
	// Level 0: ~9000 km, Level 10: ~10 km, Level 15: ~300 m, Level 20: ~10 m, Level 30: ~1 cm
	// Formula: size ≈ 9000km / 2^level
	return 9000000.0 / float64(uint64(1)<<uint(storage.storageLevel))
}
