package sliceutils

import (
	cryptorand "crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

func Map[S, T any](source []S, transformer func(S) T) []T {
	result := make([]T, len(source))
	for i, v := range source {
		result[i] = transformer(v)
	}

	return result
}

func ToBoolMap[K comparable](source []K) map[K]bool {
	result := map[K]bool{}
	for _, item := range source {
		result[item] = true
	}

	return result
}

func EmptyIfNil[S any](s []S) []S {
	if s == nil {
		return []S{}
	}

	return s
}

func Contains[S comparable](s []S, t S) bool {
	for _, val := range s {
		if val == t {
			return true
		}
	}

	return false
}

func Intersect[S comparable](s, t []S) []S {
	tMap := make(map[S]bool, len(t))
	for _, tVal := range t {
		tMap[tVal] = true
	}

	var results []S

	for _, sVal := range s {
		if _, ok := tMap[sVal]; ok {
			results = append(results, sVal)
		}
	}

	return results
}

func SoftEquals[S comparable](s, t []S) bool {
	if len(s) != len(t) {
		return false
	}

	sMap := ToValueCountMap(s)
	tMap := ToValueCountMap(t)

	for key, sCount := range sMap {
		tCount, exists := tMap[key]
		if !exists || sCount != tCount {
			return false
		}
	}

	return true
}

func Diff[S comparable](s, t []S) []S {
	tMap := make(map[S]bool, len(t))
	for _, tVal := range t {
		tMap[tVal] = true
	}

	var results []S

	for _, sVal := range s {
		if _, ok := tMap[sVal]; !ok {
			results = append(results, sVal)
		}
	}

	return results
}

func Equals[S comparable](s, t []S) bool {
	if len(s) != len(t) {
		return false
	}

	for index := range s {
		if s[index] != t[index] {
			return false
		}
	}

	return true
}

func Unique[S comparable](slice []S) []S {
	keys := make(map[S]bool, len(slice))
	list := make([]S, 0)

	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true

			list = append(list, entry)
		}
	}

	return list
}

func CountIf[T any](slice []T, getField func(T) bool) int {
	count := 0

	for _, item := range slice {
		if getField(item) {
			count++
		}
	}

	return count
}

func UniqueBy[T, S comparable](slice []T, getKey func(T) S) []T {
	keys := make(map[S]bool, len(slice))
	list := make([]T, 0)

	for _, entry := range slice {
		key := getKey(entry)
		if _, value := keys[key]; !value {
			keys[key] = true

			list = append(list, entry)
		}
	}

	return list
}

func Filter[V any](source []V, filter func(V) bool) []V {
	var result []V

	for _, v := range source {
		if filter(v) {
			result = append(result, v)
		}
	}

	return result
}

func FilterEmptyString(source []string) []string {
	return Filter(source, func(v string) bool {
		return v != ""
	})
}

func Flatten2D[V any](sources [][]V) []V {
	var result []V
	for _, source := range sources {
		result = Concat(result, source)
	}

	return result
}

func Concat[V any](sources ...[]V) []V {
	totalLength := 0
	for _, source := range sources {
		totalLength += len(source)
	}

	result := make([]V, totalLength)

	index := 0

	for _, source := range sources {
		for _, v := range source {
			result[index] = v
			index++
		}
	}

	return result
}

func ConcatUnique[V comparable](sources ...[]V) []V {
	uniqueMap := make(map[V]bool)

	totalLength := 0
	for _, source := range sources {
		totalLength += len(source)
	}

	result := make([]V, 0, totalLength)

	for _, source := range sources {
		for _, v := range source {
			if !uniqueMap[v] {
				uniqueMap[v] = true

				result = append(result, v)
			}
		}
	}

	return result
}

func ConcatUniqueAny[V any](sources ...[]V) []V {
	seen := make(map[string]struct{})

	var result []V

	toKey := func(v V) string {
		return fmt.Sprintf("%#v", v)
	}

	for _, source := range sources {
		for _, v := range source {
			key := toKey(v)
			if _, exists := seen[key]; !exists {
				seen[key] = struct{}{}

				result = append(result, v)
			}
		}
	}

	return result
}

func Index[V comparable](source []V, target V) (int, bool) {
	for index, v := range source {
		if v == target {
			return index, true
		}
	}

	return -1, false
}

func Shuffle[S any](slice []S) {
	for i := 1; i < len(slice); i++ {
		randomIndex, err := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			continue
		}
		r := int(randomIndex.Int64())
		if i != r {
			slice[r], slice[i] = slice[i], slice[r]
		}
	}
}

func Remove[V comparable](source []V, index int) []V {
	if index >= len(source) || index < 0 {
		return source
	}

	return append(source[:index], source[index+1:]...)
}

func ToMap[K comparable, V, S any](source []S, f func(S) (K, V)) map[K]V {
	result := make(map[K]V, len(source))

	for _, s := range source {
		k, v := f(s)
		result[k] = v
	}

	return result
}

func Chunk[K any](s []K, chunkSize int) [][]K {
	if len(s) == 0 {
		return [][]K{}
	}

	var chunks [][]K

	length := len(s)

	for i := 0; i < length; i += chunkSize {
		end := i + chunkSize
		if end > length {
			end = length
		}

		chunks = append(chunks, s[i:end])
	}

	return chunks
}

func ToInterface[M comparable](m []M) []any {
	s := make([]interface{}, len(m))
	for i, v := range m {
		s[i] = v
	}

	return s
}

func ToAny[T any](items []T) []any {
	data := make([]any, len(items))
	for i, v := range items {
		data[i] = v
	}

	return data
}

func ToPlaceholders[M comparable](m []M) string {
	if len(m) == 0 {
		return ""
	}

	placeholders := make([]string, 0, len(m))
	for range m {
		placeholders = append(placeholders, "?")
	}

	return strings.Join(placeholders, ",")
}

func Insert[V any](slice []V, element V, index int) []V {
	if index < 0 {
		index = len(slice) + index + 1
	}

	if index > len(slice) {
		return slice
	}

	if len(slice) == index {
		return append(slice, element)
	}

	slice = append(slice[:index+1], slice[index:]...)
	slice[index] = element

	return slice
}

func RemoveDuplicate[T comparable](t []T) []T {
	keys := make(map[T]bool)
	list := []T{}

	for _, entry := range t {
		if _, value := keys[entry]; !value {
			keys[entry] = true

			list = append(list, entry)
		}
	}

	return list
}

func Take[T any](t []T, count int) []T {
	if len(t) < count {
		return t
	}

	return t[:count]
}

func ToValueCountMap[S comparable](s []S) map[S]int {
	result := map[S]int{}
	for _, item := range s {
		result[item]++
	}

	return result
}

func IntToString(intSlice []int) []string {
	stringSlice := make([]string, len(intSlice))
	for i, v := range intSlice {
		stringSlice[i] = strconv.Itoa(v)
	}

	return stringSlice
}

func ToLowerStrings(s []string) []string {
	result := make([]string, len(s))
	for i, v := range s {
		result[i] = strings.ToLower(v)
	}

	return result
}
