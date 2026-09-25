package sliceutils

import "sort"

type MultiSorter[T any] struct {
	slices []T
	less   []func(p1, p2 T) bool
}

func (ms *MultiSorter[T]) Sort(slices []T) {
	ms.slices = slices
	sort.Sort(ms)
}

func OrderedBy[T any](less ...func(p1, p2 T) bool) *MultiSorter[T] {
	return &MultiSorter[T]{
		less: less,
	}
}

func (ms *MultiSorter[T]) Len() int {
	return len(ms.slices)
}

func (ms *MultiSorter[T]) Swap(i, j int) {
	ms.slices[i], ms.slices[j] = ms.slices[j], ms.slices[i]
}

func (ms *MultiSorter[T]) Less(i, j int) bool {
	p, q := ms.slices[i], ms.slices[j]

	var k int
	for k = 0; k < len(ms.less)-1; k++ {
		less := ms.less[k]

		switch {
		case less(p, q):
			return true
		case less(q, p):
			return false
		}
	}

	return ms.less[k](p, q)
}
