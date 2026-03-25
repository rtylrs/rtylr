package sliceutils

import (
	"reflect"
	"testing"
)

type SomeDataStructure struct {
	ID       int
	Priority int
}

func TestSorter(t *testing.T) {
	aList := []SomeDataStructure{
		{1, 2},
		{3, 5},
		{2, -1},
		{7, -5},
		{5, 5},
	}

	sortedList := []SomeDataStructure{
		{3, 5},
		{5, 5},
		{1, 2},
		{2, -1},
		{7, -5},
	}

	OrderedBy(func(p1, p2 SomeDataStructure) bool {
		return p1.Priority > p2.Priority
	},
		func(p1, p2 SomeDataStructure) bool {
			return p1.ID < p2.ID
		},
	).Sort(aList)

	if !reflect.DeepEqual(aList, sortedList) {
		t.Errorf("Not sorted by priority, id correctly")
	}
}
