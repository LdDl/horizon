package spatial

import (
	"container/heap"
	"math"

	"github.com/golang/geo/s2"
	"github.com/tidwall/rtree"
)

// EuclideanStorage Spatial datastore for Euclidean/Cartesian coordinates
type EuclideanStorage struct {
	rtree *rtree.RTreeG[uint64]
	edges map[uint64]*Edge
}

// NewEuclideanStorage Returns pointer to created EuclideanStorage
func NewEuclideanStorage() *EuclideanStorage {
	return &EuclideanStorage{
		rtree: &rtree.RTreeG[uint64]{},
		edges: make(map[uint64]*Edge),
	}
}

// GetEdge Returns edge by ID from storage
func (storage *EuclideanStorage) GetEdge(edgeID uint64) *Edge {
	return storage.edges[edgeID]
}

// AddEdge Add edge (polyline) to storage
func (storage *EuclideanStorage) AddEdge(edgeID uint64, edge *Edge) error {
	// Calculate bounding box from polyline points
	if edge.Polyline == nil || len(*edge.Polyline) == 0 {
		storage.edges[edgeID] = edge
		return nil
	}

	// Find min/max coordinates
	minX, minY := math.MaxFloat64, math.MaxFloat64
	maxX, maxY := -math.MaxFloat64, -math.MaxFloat64

	for _, pt := range *edge.Polyline {
		x, y := pt.Vector.X, pt.Vector.Y
		if x < minX {
			minX = x
		}
		if x > maxX {
			maxX = x
		}
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}
	}

	// Handle point edges with small buffer
	if minX == maxX {
		minX -= 0.0001
		maxX += 0.0001
	}
	if minY == maxY {
		minY -= 0.0001
		maxY += 0.0001
	}

	// Insert into R-tree: [minX, minY], [maxX, maxY], data
	storage.rtree.Insert([2]float64{minX, minY}, [2]float64{maxX, maxY}, edgeID)

	storage.edges[edgeID] = edge
	return nil
}

// FindInRadius implements Storage interface
// For EuclideanStorage: uses pt.Vector.X/Y as Cartesian coordinates
func (storage *EuclideanStorage) FindInRadius(pt s2.Point, radiusMeters float64) (map[uint64]float64, error) {
	x, y := pt.Vector.X, pt.Vector.Y

	result := make(map[uint64]float64)

	// Search R-tree for candidates in bounding box
	storage.rtree.Search(
		[2]float64{x - radiusMeters, y - radiusMeters},
		[2]float64{x + radiusMeters, y + radiusMeters},
		func(min, max [2]float64, edgeID uint64) bool {
			edge := storage.edges[edgeID]
			if edge == nil || edge.Polyline == nil {
				return true // continue searching
			}

			// Calculate minimum distance from point to edge polyline
			minDist := math.MaxFloat64
			for i := 0; i < len(*edge.Polyline)-1; i++ {
				p1 := (*edge.Polyline)[i]
				p2 := (*edge.Polyline)[i+1]
				dist := pointToSegmentDistance(x, y, p1.Vector.X, p1.Vector.Y, p2.Vector.X, p2.Vector.Y)
				if dist < minDist {
					minDist = dist
				}
			}

			// Only include if within radius
			if minDist <= radiusMeters {
				result[edgeID] = minDist
			}

			return true // continue searching
		},
	)

	return result, nil
}

// FindNearestInRadius implements Storage interface
// For EuclideanStorage: uses pt.Vector.X/Y as Cartesian coordinates
func (storage *EuclideanStorage) FindNearestInRadius(pt s2.Point, radiusMeters float64, n int) ([]NearestObject, error) {
	found, err := storage.FindInRadius(pt, radiusMeters)
	if err != nil {
		return nil, err
	}

	// Use heap for top-N selection
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

// pointToSegmentDistance calculates the minimum distance from point (px, py) to line segment (x1,y1)-(x2,y2)
func pointToSegmentDistance(px, py, x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1

	if dx == 0 && dy == 0 {
		// Segment is a point
		return math.Sqrt((px-x1)*(px-x1) + (py-y1)*(py-y1))
	}

	// Parameter t for the projection of point onto the line
	t := ((px-x1)*dx + (py-y1)*dy) / (dx*dx + dy*dy)

	if t < 0 {
		// Closest to first endpoint
		return math.Sqrt((px-x1)*(px-x1) + (py-y1)*(py-y1))
	} else if t > 1 {
		// Closest to second endpoint
		return math.Sqrt((px-x2)*(px-x2) + (py-y2)*(py-y2))
	}

	// Closest to interior point
	projX := x1 + t*dx
	projY := y1 + t*dy
	return math.Sqrt((px-projX)*(px-projX) + (py-projY)*(py-projY))
}
