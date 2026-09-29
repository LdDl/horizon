package spatial

import (
	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
	"github.com/google/btree"
)

type nearestCell struct {
	id       s2.CellID
	distance s1.ChordAngle
}

type nearestCellQueue []nearestCell

func (queue *nearestCellQueue) push(cell nearestCell) {
	*queue = append(*queue, cell)
	for child := len(*queue) - 1; child > 0; {
		parent := (child - 1) / 2
		if (*queue)[parent].distance <= (*queue)[child].distance {
			break
		}
		(*queue)[parent], (*queue)[child] = (*queue)[child], (*queue)[parent]
		child = parent
	}
}

func (queue *nearestCellQueue) pop() nearestCell {
	result := (*queue)[0]
	last := len(*queue) - 1
	(*queue)[0] = (*queue)[last]
	*queue = (*queue)[:last]
	for parent := 0; ; {
		child := parent*2 + 1
		if child >= len(*queue) {
			break
		}
		if child+1 < len(*queue) && (*queue)[child+1].distance < (*queue)[child].distance {
			child++
		}
		if (*queue)[parent].distance <= (*queue)[child].distance {
			break
		}
		(*queue)[parent], (*queue)[child] = (*queue)[child], (*queue)[parent]
		parent = child
	}
	return result
}

func (search *sphericalNearestSearch) enqueue(queue *nearestCellQueue, id s2.CellID) {
	if !search.cap.IntersectsCell(s2.CellFromCellID(id)) {
		return
	}
	if id.Level() == search.storage.storageLevel {
		if item := search.storage.BTree.Get(indexedItem{CellID: id}); item != nil {
			search.visit(item.(indexedItem))
		}
		return
	}
	// Visit the final four children directly without range probes or queue entries.
	if id.Level()+1 == search.storage.storageLevel {
		for _, child := range id.Children() {
			search.enqueue(queue, child)
		}
		return
	}
	var first, last s2.CellID
	var firstItem indexedItem
	search.storage.BTree.AscendGreaterOrEqual(indexedItem{CellID: id.RangeMin()}, func(item btree.Item) bool {
		firstItem = item.(indexedItem)
		first = firstItem.CellID
		return false
	})
	if first == 0 || first > id.RangeMax() {
		return
	}
	search.storage.BTree.DescendLessOrEqual(indexedItem{CellID: id.RangeMax()}, func(item btree.Item) bool {
		last = item.(indexedItem).CellID
		return false
	})
	// All occupied keys in this range share this ancestor; empty levels need no visits.
	level, _ := first.CommonAncestorLevel(last)
	id = first.Parent(level)
	cell := s2.CellFromCellID(id)
	if search.cap.IntersectsCell(cell) {
		if first == last {
			search.visit(firstItem)
			return
		}
		queue.push(nearestCell{id: id, distance: cell.Distance(search.point)})
	}
}

func (search *sphericalNearestSearch) visit(item indexedItem) {
	if search.remaining == 0 {
		search.truncated = true
		return
	}
	search.remaining--
	for _, id := range item.edgesInCell {
		search.add(id)
	}
}
