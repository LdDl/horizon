package spatial

// nearestHeap implements heap.Interface for NearestObject
// Used by:
// - S2Storage
// - EuclideanStorage
// for top-N selection
type nearestHeap []NearestObject

func (h nearestHeap) Less(i, j int) bool { return h[i].DistanceTo < h[j].DistanceTo }
func (h nearestHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h nearestHeap) Len() int           { return len(h) }

func (h *nearestHeap) Push(x interface{}) {
	*h = append(*h, x.(NearestObject))
}

func (h *nearestHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
