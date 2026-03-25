package maputils

import (
	"sort"
	"testing"

	"github.com/voxire/onion/internal/platform/utils/sliceutils"
)

func TestSplitMap(t *testing.T) {
	firstMap := map[string]int{"a": 1, "b": 2, "c": 3}

	even, odd := Split(firstMap, func(num int) bool {
		return num%2 == 0
	})

	if len(even) != 1 || !HasKey(even, "b") {
		t.Errorf("Should have one even number %d with key '%s'", len(even), "b")
	}

	if len(odd) != 2 || !HasKey(odd, "a") || !HasKey(odd, "c") {
		t.Errorf("Should have one even number %d with keys '%s' and '%s'", len(odd), "a", "c")
	}
}

func TestHasAllKeys(t *testing.T) {
	tests := []struct {
		name     string
		source   map[string]int
		keys     []string
		expected bool
	}{
		{
			name:     "All keys present",
			source:   map[string]int{"a": 1, "b": 2, "c": 3},
			keys:     []string{"a", "b"},
			expected: true,
		},
		{
			name:     "Some keys missing",
			source:   map[string]int{"a": 1, "b": 2},
			keys:     []string{"a", "c"},
			expected: false,
		},
		{
			name:     "No keys",
			source:   map[string]int{"a": 1, "b": 2},
			keys:     []string{},
			expected: true,
		},
		{
			name:     "Empty map",
			source:   map[string]int{},
			keys:     []string{"a"},
			expected: false,
		},
		{
			name:     "All keys present in empty map",
			source:   map[string]int{},
			keys:     []string{},
			expected: true,
		},
	}

	for _, tt := range tests {
		result := HasAllKeys(tt.source, tt.keys)
		if result != tt.expected {
			t.Errorf("%s: HasAllKeys(%v, %v) = %v; want %v", tt.name, tt.source, tt.keys, result, tt.expected)
		}
	}
}

func TestMergeMap(t *testing.T) {
	firstMap := map[string]int{"a": 1, "b": 2, "c": 3}
	secondMap := map[string]int{"d": 4, "e": 5, "f": 6}
	thirdMap := map[string]int{"g": 7, "h": 8, "i": 9}

	mergedMap := Merge(firstMap, secondMap, thirdMap)

	for k := range firstMap {
		if _, ok := mergedMap[k]; !ok {
			t.Errorf("Expected key %s in merged map, didn't find it", k)
		}
	}

	for k := range secondMap {
		if _, ok := mergedMap[k]; !ok {
			t.Errorf("Expected key %s in merged map, didn't find it", k)
		}
	}

	for k := range thirdMap {
		if _, ok := mergedMap[k]; !ok {
			t.Errorf("Expected key %s in merged map, didn't find it", k)
		}
	}
}

func TestGroupBy(t *testing.T) {
	type Item struct {
		Color string
		Name  string
	}
	items := []Item{
		{Color: "red", Name: "apple"},
		{Color: "red", Name: "strawberry"},
		{Color: "green", Name: "green apple"},
	}

	colorMap := GroupBy(items, func(item Item) string {
		return item.Color
	})

	if !HasKey(colorMap, "red") || !HasKey(colorMap, "green") {
		t.Errorf("Should have keys '%s' and '%s'", "red", "green")
	}

	if colorMap["red"][0].Name != "apple" || colorMap["red"][1].Name != "strawberry" {
		t.Errorf("Should have items '%s' and '%s'", "apple", "strawberry")
	}

	if colorMap["green"][0].Name != "green apple" {
		t.Errorf("Should have item '%s'", "green apple")
	}
}

func TestKeys(t *testing.T) {
	theMap := map[string]int{"a": 1, "b": 2, "c": 3}
	expectedKeys := []string{"a", "b", "c"}
	keys := Keys(theMap)

	sort.Strings(keys)
	sort.Strings(expectedKeys)

	for k, v := range keys {
		if v != expectedKeys[k] {
			t.Errorf("Expected key %s in the list of keys, didn't find it", v)
		}
	}
}

func TestMap(t *testing.T) {
	theMap := map[string]int{"a": 1, "b": 2, "c": 3}
	expectedMap := map[string]int{"a": 2, "b": 4, "c": 6}

	transformedMap := Map(theMap, func(s int) int {
		return s * 2
	})

	for k, v := range transformedMap {
		if v != expectedMap[k] {
			t.Errorf("Expected value at key %s to be %d, but found %d instead", k, expectedMap[k], v)
		}
	}
}

func TestMapKeys(t *testing.T) {
	theMap := map[int]int{1: 1, 2: 2, 3: 3}
	expectedKeys := []int{2, 4, 6}

	transformedKeysMap := MapKeys(theMap, func(s int) int {
		return s * 2
	})

	transformedKeys := Keys(transformedKeysMap)

	sort.Ints(transformedKeys)
	sort.Ints(expectedKeys)

	for k, v := range transformedKeys {
		if v != expectedKeys[k] {
			t.Errorf("Expected key to be %d but found %d instead", v, expectedKeys[k])
		}
	}
}

func TestMapFilterNil(t *testing.T) {
	a := "a"
	b := "b"
	c := "c"
	aMapWithNils := map[string]*string{
		"a": &a,
		"b": &b,
		"c": &c,
		"d": nil,
		"e": nil,
	}

	aMapWithoutNils := FilterNotNil(aMapWithNils)

	for _, v := range aMapWithoutNils {
		if v == nil {
			t.Errorf("Expected to not find any nil value in the map, but found one")
		}
	}
}

func TestAny(t *testing.T) {
	dataMap := map[string]int{
		"1": 1,
		"2": 2,
		"3": 3,
	}
	val, found := Any(dataMap)
	if !found {
		t.Errorf("Should have found a value in the map, no value found")

		return
	}
	if !sliceutils.Contains([]int{1, 2, 3}, val) {
		t.Errorf("Value should be either 1 or 2 or 3, %d returned", val)

		return
	}
	emptyMap := map[string]int{}
	_, found = Any(emptyMap)
	if found {
		t.Errorf("Any should have returned empty results, but a result was returned")
	}
}

func TestEquals(t *testing.T) {
	firstMap := map[string]int{"a": 1, "b": 2, "c": 3}
	secondMap := map[string]int{"b": 2, "c": 3, "a": 1}
	thirdMap := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}
	if !Equals(firstMap, secondMap) {
		t.Errorf("Maps should be equal")
	}
	if Equals(firstMap, thirdMap) {
		t.Errorf("Maps should not be equal")
	}
}

func TestGetOrderedFromMap(t *testing.T) {
	firstMapKeys := []string{"a", "b", "c"}
	firstMap := map[string]int{"a": 1, "c": 3, "b": 2}
	orderedValues := GetOrderedFromMap(firstMap, firstMapKeys)
	for i, k := range firstMapKeys {
		expectedValue := firstMap[k]
		if orderedValues[i] != expectedValue {
			t.Errorf("Expected value %d at index %d, found %d instead", expectedValue, i, orderedValues[i])
		}
	}
}
