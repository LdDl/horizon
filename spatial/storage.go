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

// Implement heap (for getting top-N elements)
type s2Heap []NearestObject

func (h s2Heap) Less(i, j int) bool { return h[i].DistanceTo < h[j].DistanceTo }
func (h s2Heap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h s2Heap) Len() int           { return len(h) }

func (h *s2Heap) Push(x interface{}) {
	*h = append(*h, x.(NearestObject))
}

func (h *s2Heap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
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
	h := &s2Heap{}
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
