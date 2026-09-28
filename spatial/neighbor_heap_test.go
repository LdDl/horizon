package spatial

import (
	"container/heap"
	"math"
	"reflect"
	"testing"
)

type nearestOrderCase struct {
	name string
	want []NearestObject
}

func eachNearestPermutation(objects []NearestObject, visit func([]NearestObject)) {
	objects = append([]NearestObject(nil), objects...)
	var permute func(int)
	permute = func(start int) {
		if start == len(objects) {
			visit(objects)
			return
		}
		for i := start; i < len(objects); i++ {
			objects[start], objects[i] = objects[i], objects[start]
			permute(start + 1)
			objects[start], objects[i] = objects[i], objects[start]
		}
	}
	permute(0)
}

func TestNearestHeapOrderIndependentOfInsertion(t *testing.T) {
	cases := []nearestOrderCase{
		{name: "equal distances", want: []NearestObject{{2, 10}, {7, 10}, {9, 10}}},
		{name: "distance before ID", want: []NearestObject{{9, 1}, {2, 2}, {7, 2}}},
		{name: "one ULP is not a tie", want: []NearestObject{{9, 1}, {7, math.Nextafter(1, 2)}, {2, math.Nextafter(math.Nextafter(1, 2), 2)}}},
		{name: "signed zero", want: []NearestObject{{2, 0}, {7, math.Copysign(0, -1)}, {9, 0}}},
		{name: "full width IDs", want: []NearestObject{{1 << 53, 10}, {1<<53 + 1, 10}, {math.MaxUint64, 10}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			eachNearestPermutation(tc.want, func(order []NearestObject) {
				for _, incremental := range []bool{false, true} {
					for n := 0; n <= len(order); n++ {
						h := &nearestHeap{}
						if incremental {
							for _, object := range order {
								heap.Push(h, object)
							}
						} else {
							*h = append(*h, order...)
							heap.Init(h)
						}
						got := make([]NearestObject, n)
						for i := range got {
							got[i] = heap.Pop(h).(NearestObject)
						}
						if !reflect.DeepEqual(got, tc.want[:n]) {
							t.Fatalf("order=%v incremental=%v n=%d: got %v, want %v", order, incremental, n, got, tc.want[:n])
						}
					}
				}
			})
		})
	}
}
