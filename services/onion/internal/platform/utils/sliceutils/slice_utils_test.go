package sliceutils

import (
	"reflect"
	"testing"
)

func TestSliceFilterEmptyString(t *testing.T) {
	strings := FilterEmptyString([]string{"hello", "", "world"})

	if len(strings) != 2 {
		t.Errorf("should have filtered out second element")
	}

	if strings[0] != "hello" || strings[1] != "world" {
		t.Errorf("should have an array containing only hello,world")
	}
}

func TestFlatten2D(t *testing.T) {
	slice2d := [][]string{
		{"hello"},
		{"world"},
	}

	slice1d := Flatten2D(slice2d)

	if len(slice1d) != 2 {
		t.Errorf("should have a slice with 2 elements")
	}

	if slice1d[0] != "hello" || slice1d[1] != "world" {
		t.Errorf("should have a slice containing hello,world")
	}
}

func TestCountIf(t *testing.T) {
	if CountIf([]int{1, 2, 3, 4, 5}, func(i int) bool { return i%2 == 0 }) != 2 {
		t.Errorf("should have counted 2 even numbers only")
	}

	if CountIf([]int{1, 2, 3, 4, 5}, func(i int) bool { return i < 0 }) != 0 {
		t.Errorf("should not have counted anything")
	}

	if CountIf([]int{1, 2, 3, 4, 5}, func(_ int) bool { return true }) != 5 {
		t.Errorf("should have counted everything")
	}

	if CountIf([]int{}, func(_ int) bool { return true }) != 0 {
		t.Errorf("should not have counted anything")
	}
}

func TestUniqueBy(t *testing.T) {
	type testDataItem struct {
		ID int
	}

	type testDataItemCase struct {
		Input    []testDataItem
		Expected []testDataItem
	}

	testCases := []testDataItemCase{
		{
			Input:    []testDataItem{},
			Expected: []testDataItem{},
		},
		{
			Input:    []testDataItem{{ID: 1}, {ID: 2}, {ID: 3}, {ID: 2}},
			Expected: []testDataItem{{ID: 1}, {ID: 2}, {ID: 3}},
		},
		{
			Input:    []testDataItem{{ID: 1}, {ID: 2}, {ID: 3}},
			Expected: []testDataItem{{ID: 1}, {ID: 2}, {ID: 3}},
		},
		{
			Input:    []testDataItem{{ID: 1}, {ID: 1}, {ID: 1}},
			Expected: []testDataItem{{ID: 1}},
		},
	}
	for _, testCase := range testCases {
		result := UniqueBy(testCase.Input, func(item testDataItem) int { return item.ID })
		if !reflect.DeepEqual(result, testCase.Expected) {
			t.Errorf("Expected %v, but got %v", testCase.Expected, result)
		}
	}
}

func TestListMap(t *testing.T) {
	list := []int{1, 2, 3, 4, 5, 6}
	expectedList := []int{2, 4, 6, 8, 10, 12}
	transformedList := Map(list, func(s int) int {
		return s * 2
	})

	for k, v := range transformedList {
		if v != expectedList[k] {
			t.Errorf("Expected to find %d in the list but found %d", expectedList[k], v)
		}
	}
}

func TestListContains(t *testing.T) {
	intList := []int{1, 2, 3, 4, 5, 6, 7}
	missingInt := 8

	containsInt := Contains(intList, missingInt)

	if containsInt {
		t.Errorf("Expected int list to not contain %d but found it", missingInt)
	}

	stringList := []string{"a", "b", "c"}
	existingString := "b"

	containsString := Contains(stringList, existingString)

	if !containsString {
		t.Errorf("Expected string list to contain %s but did not find it", existingString)
	}
}

func TestIntersect(t *testing.T) {
	list1 := []int{1, 2, 2, 2}
	list2 := []int{2, 1}
	intersectionExpectedResults := []int{1, 2, 2, 2}
	intersectionActualResults := Intersect(list1, list2)
	if !Equals(intersectionExpectedResults, intersectionActualResults) {
		t.Errorf("Intersection is not equal")

		return
	}
	list1 = []int{2, 1}
	list2 = []int{1, 2, 2, 2}
	intersectionExpectedResults = []int{2, 1}
	intersectionActualResults = Intersect(list1, list2)
	if !Equals(intersectionExpectedResults, intersectionActualResults) {
		t.Errorf("Intersection is not equal")

		return
	}
	list1 = []int{1, 2}
	list2 = []int{3, 4, 5, 6}
	intersectionExpectedResults = []int{}
	intersectionActualResults = Intersect(list1, list2)
	if !Equals(intersectionExpectedResults, intersectionActualResults) {
		t.Errorf("Intersection is not equal")

		return
	}
}

func TestEquals(t *testing.T) {
	list1 := []int{1, 2, 3, 4, 5}
	list2 := []int{3, 4, 5, 6, 7, 8}
	if Equals(list1, list2) {
		t.Errorf("These two lists shouldn't be equal")

		return
	}
	list1 = []int{1, 2, 3, 4, 5}
	list2 = []int{5, 4, 3, 2, 1}
	if Equals(list1, list2) {
		t.Errorf("These two lists shouldn't be equal")

		return
	}
	list1 = []int{1, 2, 3, 4, 5}
	list2 = []int{1, 2, 3, 4, 5}
	if !Equals(list1, list2) {
		t.Errorf("These two lists should be equal, Equals returned false.")
	}
}

func TestSoftEquals(t *testing.T) {
	list1 := []int{1, 2, 3, 4, 5}
	list2 := []int{3, 4, 5, 6, 7, 8}
	if SoftEquals(list1, list2) {
		t.Errorf("These two lists shouldn't be equal")

		return
	}
	list1 = []int{1, 2, 3, 4, 5, 6}
	list2 = []int{5, 4, 3, 2, 1}
	if SoftEquals(list1, list2) {
		t.Errorf("These two lists shouldn't be equal")

		return
	}
	list1 = []int{1, 2, 3, 4, 5, 5}
	list2 = []int{5, 4, 3, 2, 1, 1}
	if SoftEquals(list1, list2) {
		t.Errorf("These two lists shouldn't be equal")

		return
	}
	list1 = []int{1, 2, 3, 4, 5, 5}
	list2 = []int{5, 4, 3, 2, 1}
	if SoftEquals(list1, list2) {
		t.Errorf("These two lists shouldn't be equal")

		return
	}
	list1 = []int{1, 2, 3, 4, 5, 6}
	list2 = []int{5, 4, 3, 2, 1, 3}
	if SoftEquals(list1, list2) {
		t.Errorf("These two lists shouldn't be equal")

		return
	}
	list1 = []int{1, 2, 3, 4, 5}
	list2 = []int{1, 2, 3, 4, 5}
	if !SoftEquals(list1, list2) {
		t.Errorf("These two lists should be equal, Equals returned false.")
	}
	list1 = []int{1, 2, 3, 5, 4}
	list2 = []int{1, 2, 3, 4, 5}
	if !SoftEquals(list1, list2) {
		t.Errorf("These two lists should be equal, Equals returned false.")
	}
}

func TestRemove(t *testing.T) {
	aSlice := []int{1, 2, 3, 4, 5}

	aSlice = Remove(aSlice, 3)
	expected := []int{1, 2, 3, 5}

	if !Equals(aSlice, expected) {
		t.Errorf("Resulting slice after removal did not match expected slice")
	}
}

func TestToMap(t *testing.T) {
	aSlice := []int{1, 2, 3, 4, 5}
	aMap := ToMap(aSlice, func(i int) (int, int) {
		return i * 2, i
	})

	expectedMap := map[int]int{
		2:  1,
		4:  2,
		6:  3,
		8:  4,
		10: 5,
	}

	eq := reflect.DeepEqual(aMap, expectedMap)

	if !eq {
		t.Errorf("Output map did not match expected map")
	}
}

func TestChunk(t *testing.T) {
	tests := []struct {
		name      string
		input     []string
		chunkSize int
		expected  [][]string
	}{
		{
			name:      "Empty input",
			input:     []string{},
			chunkSize: 2,
			expected:  [][]string{},
		},
		{
			name:      "Single chunk",
			input:     []string{"a", "b"},
			chunkSize: 2,
			expected:  [][]string{{"a", "b"}},
		},
		{
			name:      "Multiple chunks",
			input:     []string{"a", "b", "c", "d"},
			chunkSize: 2,
			expected:  [][]string{{"a", "b"}, {"c", "d"}},
		},
		{
			name:      "Incomplete chunk",
			input:     []string{"a", "b", "c"},
			chunkSize: 2,
			expected:  [][]string{{"a", "b"}, {"c"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Chunk(tt.input, tt.chunkSize)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Chunk(%v, %d) = %v, expected %v", tt.input, tt.chunkSize, result, tt.expected)
			}
		})
	}
}

func TestInsert(t *testing.T) {
	slice := []int{2, 3, 4}
	expected := []int{1, 2, 3, 4}
	result := Insert(slice, 1, 0)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, but got %v", expected, result)
	}

	slice = []int{1, 2, 4}
	expected = []int{1, 2, 3, 4}
	result = Insert(slice, 3, 2)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, but got %v", expected, result)
	}

	slice = []int{1, 2, 3}
	expected = []int{1, 2, 3, 4}
	result = Insert(slice, 4, 3)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, but got %v", expected, result)
	}

	slice = []int{}
	expected = []int{1}
	result = Insert(slice, 1, 0)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, but got %v", expected, result)
	}

	slice = []int{1, 2, 3}
	result = Insert(slice, 4, 4)
	if len(result) != len(slice) {
		t.Errorf("Expected slice length to remain unchanged, but got %v", len(result))
	}
	if !reflect.DeepEqual(result, slice) {
		t.Errorf("Expected the original slice to be unchanged, but got %v", result)
	}

	slice = []int{1, 2, 3}
	result = Insert(slice, 4, -1)
	expected = []int{1, 2, 3, 4}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, but got %v", expected, result)
	}
}

func TestDiff(t *testing.T) {
	var s1 []int
	var t1 []int
	var expected1 []int
	result1 := Diff(s1, t1)
	if !reflect.DeepEqual(result1, expected1) {
		t.Errorf("Expected %v, but got %v", expected1, result1)
	}

	var s2 []string
	t2 := []string{"a", "b", "c"}
	var expected2 []string
	result2 := Diff(s2, t2)
	if !reflect.DeepEqual(result2, expected2) {
		t.Errorf("Expected %v, but got %v", expected2, result2)
	}

	s3 := []int{1, 2, 3, 4}
	t3 := []int{2, 3}
	expected3 := []int{1, 4}
	result3 := Diff(s3, t3)
	if !reflect.DeepEqual(result3, expected3) {
		t.Errorf("Expected %v, but got %v", expected3, result3)
	}

	s4 := []string{"apple", "banana", "cherry"}
	t4 := []string{"orange", "kiwi"}
	expected4 := []string{"apple", "banana", "cherry"}
	result4 := Diff(s4, t4)
	if !reflect.DeepEqual(result4, expected4) {
		t.Errorf("Expected %v, but got %v", expected4, result4)
	}
}

func TestRemoveDuplicate(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "Empty input",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "No duplicates",
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "Duplicates",
			input:    []string{"a", "b", "a"},
			expected: []string{"a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoveDuplicate(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("RemoveDuplicate(%v) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConcatInt(t *testing.T) {
	tests := []struct {
		name     string
		sources  [][]int
		expected []int
	}{
		{
			name:     "Concatenation of empty int slices",
			sources:  [][]int{{}, {}},
			expected: []int{},
		},
		{
			name:     "Concatenation of non-empty int slices",
			sources:  [][]int{{1, 2}, {3, 4}},
			expected: []int{1, 2, 3, 4},
		},
		{
			name:     "Concatenation of int slices with different lengths",
			sources:  [][]int{{1, 2, 3}, {4, 5}},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "Concatenation of multiple int slices",
			sources:  [][]int{{1}, {2, 3}, {4, 5, 6}},
			expected: []int{1, 2, 3, 4, 5, 6},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Concat(tt.sources...)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("Concat() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestConcatUnique(t *testing.T) {
	tests := []struct {
		name     string
		sources  [][]int
		expected []int
	}{
		{
			name:     "Concatenation of empty int slices",
			sources:  [][]int{{}, {}},
			expected: []int{},
		},
		{
			name:     "Concatenation of non-empty int slices",
			sources:  [][]int{{1, 2}, {3, 4}},
			expected: []int{1, 2, 3, 4},
		},
		{
			name:     "Concatenation of int slices with different lengths",
			sources:  [][]int{{1, 2, 3}, {4, 5}},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "Concatenation of multiple int slices",
			sources:  [][]int{{1}, {2, 3}, {4, 5, 6}},
			expected: []int{1, 2, 3, 4, 5, 6},
		},
		{
			name:     "Concatenation of int slices with duplicates",
			sources:  [][]int{{1, 2, 3}, {2, 3, 4}},
			expected: []int{1, 2, 3, 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConcatUnique(tt.sources...)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("ConcatUnique() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestToLowerStrings(t *testing.T) {
	testCases := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "Empty input",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "Lowercase strings",
			input:    []string{"hello", "world"},
			expected: []string{"hello", "world"},
		},
		{
			name:     "Uppercase strings",
			input:    []string{"HELLO", "WORLD"},
			expected: []string{"hello", "world"},
		},
		{
			name:     "Mixed case strings",
			input:    []string{"Hello", "World"},
			expected: []string{"hello", "world"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := ToLowerStrings(testCase.input)
			if !reflect.DeepEqual(result, testCase.expected) {
				t.Errorf("ToLowerStrings(%v) = %v, expected %v", testCase.input, result, testCase.expected)
			}
		})
	}
}
