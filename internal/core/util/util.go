package util

func UniqueSlice[T comparable](s []T) []T {
	seen := make(map[T]struct{}, len(s))
	result := make([]T, 0, len(s))

	for _, i := range s {
		if _, found := seen[i]; !found {
			seen[i] = struct{}{}
			result = append(result, i)
		}
	}

	return result
}
