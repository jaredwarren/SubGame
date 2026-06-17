package sliceutil

import "math/rand"

// Shuffle shuffles a copy of the slice using the provided rand.Rand.
func Shuffle[T any](slice []T, r *rand.Rand) []T {
	shuffled := make([]T, len(slice))
	copy(shuffled, slice)
	for i := len(shuffled) - 1; i > 0; i-- {
		j := r.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}
	return shuffled
}

// FilterInPlace filters the slice in-place without new allocations.
func FilterInPlace[T any](slice []T, keep func(T) bool) []T {
	activeCount := 0
	for _, item := range slice {
		if keep(item) {
			slice[activeCount] = item
			activeCount++
		}
	}
	var zero T
	for i := activeCount; i < len(slice); i++ {
		slice[i] = zero
	}
	return slice[:activeCount]
}
