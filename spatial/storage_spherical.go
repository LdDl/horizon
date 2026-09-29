package spatial

import (
	"container/heap"
	"sort"

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
	return storage.SearchInRadius(s2.PointFromLatLng(latlng), radius)
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
	centerPoint := pt
	centerAngle := radius / EarthRadius
	cap := s2.CapFromCenterAngle(centerPoint, s1.Angle(centerAngle))
	rc := s2.RegionCoverer{MaxLevel: storage.storageLevel, MinLevel: storage.storageLevel}
	cu := rc.Covering(cap)
	result := make(map[uint64]float64)
	radiusAngle := s1.Angle(centerAngle)
	for _, cellID := range cu {
		item := storage.BTree.Get(indexedItem{CellID: cellID})
		if item != nil {
			for _, edgeID := range item.(indexedItem).edgesInCell {
				if _, exists := result[edgeID]; exists {
					continue
				}
				polyline := storage.edges[edgeID]
				if polyline == nil || polyline.Polyline == nil || polyline.Polyline.NumEdges() < 1 {
					continue
				}
				// Bounding-cap prune: if the polyline's cap is entirely outside
				// the search radius, skip the per-segment scan.
				if polyline.BoundRadius > 0 {
					chord := s2.ChordAngleBetweenPoints(pt, polyline.BoundCenter)
					if chord.Angle()-polyline.BoundRadius > radiusAngle {
						continue
					}
				}
				result[edgeID] = sphericalPolylineDistance(pt, polyline.Polyline)
			}
		}
	}
	// Keep all scanned distances until filtering so rejected edges are not rescanned in another cell.
	for edgeID, distance := range result {
		if !(distance <= radius) {
			delete(result, edgeID)
		}
	}
	return result, nil
}

// sphericalPolylineDistance returns the minimum distance from the query point to the segments, in meters.
func sphericalPolylineDistance(pt s2.Point, line *s2.Polyline) float64 {
	minDist := s1.InfChordAngle()
	for i := 0; i < line.NumEdges(); i++ {
		edge := line.Edge(i)
		minDist, _ = s2.UpdateMinDistance(pt, edge.V0, edge.V1, minDist)
	}
	return minDist.Angle().Radians() * EarthRadius
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

// maxNearestCellVisits bounds intersecting leaf visits before a complete edge scan.
const maxNearestCellVisits = 256

// FindNearest returns up to n distinct edges ordered by distance and then edge ID.
// Occupied cells are visited until every remaining cell is disjoint from the search cap.
func (storage *S2Storage) FindNearest(pt s2.Point, n int) ([]NearestObject, error) {
	return storage.findNearest(pt, n, maxNearestCellVisits), nil
}

func (storage *S2Storage) findNearest(pt s2.Point, n, cellBudget int) []NearestObject {
	if n <= 0 || len(storage.edges) == 0 {
		return nil
	}
	if n > len(storage.edges) {
		n = len(storage.edges)
	}
	search := sphericalNearestSearch{
		storage:   storage,
		point:     pt,
		limit:     n,
		seen:      make(map[uint64]bool),
		nearest:   make([]NearestObject, 0, n),
		remaining: cellBudget,
	}

	search.cap = s2.FullCap()
	if n < len(storage.edges) && cellBudget > 0 {
		center := s2.CellFromPoint(pt).ID().Parent(storage.storageLevel)
		if item := storage.BTree.Get(indexedItem{CellID: center}); item != nil {
			search.remaining--
			for _, id := range item.(indexedItem).edgesInCell {
				search.add(id)
			}
		}
		// Any k seeds give an upper bound; the hierarchy still certifies the answer.
		if len(search.nearest) < n {
			storage.BTree.AscendGreaterOrEqual(indexedItem{CellID: center.Next()}, func(item btree.Item) bool {
				search.visit(item.(indexedItem))
				return false
			})
		}
		var buffer [64]nearestCell
		queue := nearestCellQueue(buffer[:0])
		roots := search.cap.CellUnionBound()
		for i, root := range roots {
			if root.Level() > storage.storageLevel {
				root = root.Parent(storage.storageLevel)
			}
			duplicate := false
			for j := 0; j < i; j++ {
				if roots[j] == root {
					duplicate = true
					break
				}
			}
			roots[i] = root
			if !duplicate {
				search.enqueue(&queue, root)
			}
		}
		for len(queue) > 0 && !search.truncated {
			item := queue.pop()
			cell := s2.CellFromCellID(item.id)
			// Queue distances only schedule work. Only geometric disjointness prunes it.
			if !search.cap.IntersectsCell(cell) {
				continue
			}
			for _, child := range item.id.Children() {
				search.enqueue(&queue, child)
			}
		}
		if !search.truncated {
			return search.result()
		}
	}

	// Each remaining edge occurs once in this scan; growing seen here is unnecessary.
	for id, edge := range storage.edges {
		if !search.seen[id] {
			search.consider(id, edge)
		}
	}
	return search.result()
}

type sphericalNearestSearch struct {
	storage   *S2Storage
	point     s2.Point
	limit     int
	seen      map[uint64]bool
	nearest   nearestHeap
	cap       s2.Cap
	remaining int
	truncated bool
}

func (search *sphericalNearestSearch) add(id uint64) {
	if search.seen[id] {
		return
	}
	search.seen[id] = true
	search.consider(id, search.storage.edges[id])
}

func (search *sphericalNearestSearch) consider(id uint64, edge *Edge) {
	if edge == nil || edge.Polyline == nil || edge.Polyline.NumEdges() < 1 {
		return
	}
	distance := sphericalPolylineDistance(search.point, edge.Polyline)
	candidate := NearestObject{EdgeID: id, DistanceTo: distance}
	if len(search.nearest) < search.limit {
		search.nearest = append(search.nearest, candidate)
		// Keep the worst selected candidate at the root of a bounded max-heap.
		for child := len(search.nearest) - 1; child > 0; {
			parent := (child - 1) / 2
			if !search.nearest.Less(parent, child) {
				break
			}
			search.nearest.Swap(parent, child)
			child = parent
		}
		search.updateCap()
		return
	}
	worst := search.nearest[0]
	if distance > worst.DistanceTo || distance == worst.DistanceTo && id >= worst.EdgeID {
		return
	}
	search.nearest[0] = candidate
	for parent := 0; ; {
		child := 2*parent + 1
		if child >= len(search.nearest) {
			break
		}
		if child+1 < len(search.nearest) && search.nearest.Less(child, child+1) {
			child++
		}
		if !search.nearest.Less(parent, child) {
			break
		}
		search.nearest.Swap(parent, child)
		parent = child
	}
	search.updateCap()
}

func (search *sphericalNearestSearch) updateCap() {
	if len(search.nearest) == search.limit {
		search.cap = s2.CapFromCenterAngle(search.point, s1.Angle(search.nearest[0].DistanceTo/EarthRadius))
	}
}

func (search *sphericalNearestSearch) result() []NearestObject {
	sort.Sort(search.nearest)
	return search.nearest
}
