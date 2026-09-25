package maputils

func Merge[K comparable, V any](all ...map[K]V) map[K]V {
	result := make(map[K]V)

	for _, m := range all {
		for k, v := range m {
			result[k] = v
		}
	}

	return result
}

func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	return keys
}

func GetOrderedFromMap[K comparable, V any](items map[K]V, orderedKeys []K) []V {
	var values []V

	for _, k := range orderedKeys {
		if HasKey(items, k) {
			values = append(values, items[k])
		}
	}

	return values
}

func GroupBy[K comparable, V any](items []V, getKey func(V) K) map[K][]V {
	result := make(map[K][]V)

	for _, item := range items {
		key := getKey(item)
		if !HasKey(result, key) {
			result[key] = []V{}
		}

		result[key] = append(result[key], item)
	}

	return result
}
func Map[K comparable, V, T any](source map[K]V, transformer func(V) T) map[K]T {
	result := make(map[K]T)
	for k, v := range source {
		result[k] = transformer(v)
	}

	return result
}

func MapKeys[K comparable, V any, T comparable](source map[K]V, transformer func(K) T) map[T]V {
	result := make(map[T]V)
	for k, v := range source {
		result[transformer(k)] = v
	}

	return result
}

func HasKey[K comparable, V any](source map[K]V, key K) bool {
	if _, ok := source[key]; ok {
		return true
	}

	return false
}
func HasAllKeys[K comparable, V any](source map[K]V, keys []K) bool {
	for _, key := range keys {
		if _, ok := source[key]; !ok {
			return false
		}
	}

	return true
}
func Filter[K comparable, V any](source map[K]V, filter func(V) bool) map[K]V {
	result := make(map[K]V)

	for k, v := range source {
		if filter(v) {
			result[k] = v
		}
	}

	return result
}

func Split[K comparable, V any](source map[K]V, filter func(V) bool) (included, excluded map[K]V) {
	meetsCondition := make(map[K]V)
	doesntMeetCondition := make(map[K]V)

	for k, v := range source {
		if filter(v) {
			meetsCondition[k] = v
		} else {
			doesntMeetCondition[k] = v
		}
	}

	return meetsCondition, doesntMeetCondition
}

func FilterNotNil[K comparable, V any](source map[K]*V) map[K]*V {
	return Filter(source, func(v *V) bool {
		return v != nil
	})
}

func Any[K comparable, V any](source map[K]V) (V, bool) {
	for _, v := range source {
		return v, true
	}

	return *new(V), false
}

func Equals[K, V comparable](source, target map[K]V) bool {
	if len(source) != len(target) {
		return false
	}

	for k, v := range source {
		if !HasKey(target, k) || target[k] != v {
			return false
		}
	}

	return true
}
