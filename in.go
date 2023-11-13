package main

// Checks using linear search if an item is present in the slice
func In[T comparable](slice []T, candidate T) bool {
	for _, got := range slice {
		if got == candidate {
			return true
		}
	}
	return false
}
