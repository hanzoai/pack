package utils

import "sort"

func RemoveDuplicates[T comparable](sliceList []T) []T {
	allKeys := make(map[T]bool)
	list := []T{}
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

// MergeStringSlicePointers combines multiple string slice pointers, deduplicates values, and sorts them
func MergeStringSlicePointers(slices ...*[]string) *[]string {
	if len(slices) == 0 {
		return nil
	}

	var allStrings []string
	for _, slice := range slices {
		if slice != nil {
			allStrings = append(allStrings, *slice...)
		}
	}

	if len(allStrings) == 0 {
		return nil
	}

	// Deduplicate and sort
	seen := make(map[string]bool)
	var uniqueStrings []string
	for _, s := range allStrings {
		if !seen[s] {
			seen[s] = true
			uniqueStrings = append(uniqueStrings, s)
		}
	}
	sort.Strings(uniqueStrings)
	return &uniqueStrings
}
