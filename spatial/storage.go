package spatial

import "github.com/golang/geo/s2"

// StorageType represents the type of spatial storage
type StorageType int

const (
	// StorageTypeSpherical uses S2 geometry for WGS84 coordinates (lon/lat)
	StorageTypeSpherical StorageType = iota
	// StorageTypeEuclidean uses Euclidean geometry for Cartesian coordinates (future)
	StorageTypeEuclidean
)

// Storage is the interface for spatial storage implementations
type Storage interface {
	// AddEdge adds an edge to the storage
	AddEdge(edgeID uint64, edge *Edge) error

	// GetEdge returns an edge by ID
	GetEdge(edgeID uint64) *Edge

	// FindInRadius searches for edges within a radius from point
	// For spherical: uses s2.Point on unit sphere
	// For Euclidean: uses s2.Point.Vector.X/Y as Cartesian coordinates
	// Returns map of edge ID to distance
	FindInRadius(pt s2.Point, radiusMeters float64) (map[uint64]float64, error)

	// FindNearestInRadius returns the N nearest edges within a radius
	// For spherical: uses s2.Point on unit sphere
	// For Euclidean: uses s2.Point.Vector.X/Y as Cartesian coordinates
	FindNearestInRadius(pt s2.Point, radiusMeters float64, n int) ([]NearestObject, error)
}

// StorageOptions holds configuration for creating a Storage
type StorageOptions struct {
	StorageType  StorageType
	StorageLevel int // S2 cell level for spherical storage
	BTreeDegree  int // B-tree degree
}

// StorageOption is a functional option for configuring Storage
type StorageOption func(*StorageOptions)

// WithStorageLevel sets the S2 cell level for spherical storage
func WithStorageLevel(level int) StorageOption {
	return func(o *StorageOptions) {
		o.StorageLevel = level
	}
}

// WithBTreeDegree sets the B-tree degree
func WithBTreeDegree(degree int) StorageOption {
	return func(o *StorageOptions) {
		o.BTreeDegree = degree
	}
}

// NewStorage creates a new Storage based on provided options
func NewStorage(storageType StorageType, opts ...StorageOption) Storage {
	// Default options
	options := &StorageOptions{
		StorageType:  storageType,
		StorageLevel: 17,
		BTreeDegree:  35,
	}

	// Apply provided options
	for _, opt := range opts {
		opt(options)
	}

	switch options.StorageType {
	case StorageTypeSpherical:
		return NewS2Storage(options.StorageLevel, options.BTreeDegree)
	default:
		panic("Need to implement Euclidean storage")
		return NewS2Storage(options.StorageLevel, options.BTreeDegree)
	}
}
