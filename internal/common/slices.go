package common

// FilterEmpty returns a new slice containing only the non-zero values from the input.
// It filters out zero values (e.g., "", 0, nil for pointers/interfaces).
func FilterEmpty[T comparable](items ...T) []T {
	result := make([]T, 0, len(items))
	var zero T
	for _, item := range items {
		if item != zero {
			result = append(result, item)
		}
	}
	return result
}
