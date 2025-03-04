package utils

import (
	"sort"
	"strings"
)

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

func CapitalizeFirst(s string) string {
	if s == "" {
		return ""
	}

	runes := []rune(s)
	runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
	return string(runes)
}

func ParseVersions(versions []string) map[string]string {
	parsedVersions := make(map[string]string)

	for _, version := range versions {
		parts := strings.Split(version, "@")
		if len(parts) == 1 {
			parsedVersions[parts[0]] = "latest"
		} else {
			parsedVersions[parts[0]] = parts[1]
		}
	}

	return parsedVersions
}
