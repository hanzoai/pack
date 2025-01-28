package utils

import (
	"reflect"
	"testing"
)

func TestRemoveDuplicates(t *testing.T) {
	numbers := []int{1, 2, 2, 3, 3, 3, 4}
	expectedNumbers := []int{1, 2, 3, 4}
	result := RemoveDuplicates(numbers)
	if !reflect.DeepEqual(result, expectedNumbers) {
		t.Errorf("RemoveDuplicates() with ints = %v, want %v", result, expectedNumbers)
	}

	strings := []string{"a", "b", "a", "c", "b", "c", "d"}
	expectedStrings := []string{"a", "b", "c", "d"}
	strResult := RemoveDuplicates(strings)
	if !reflect.DeepEqual(strResult, expectedStrings) {
		t.Errorf("RemoveDuplicates() with strings = %v, want %v", strResult, expectedStrings)
	}
}

func TestMergeStringSlicePointers(t *testing.T) {
	tests := []struct {
		name     string
		inputs   []*[]string
		expected *[]string
	}{
		{
			name:     "nil input",
			inputs:   nil,
			expected: nil,
		},
		{
			name:     "empty input array",
			inputs:   nil,
			expected: nil,
		},
		{
			name: "single nil slice",
			inputs: []*[]string{
				nil,
			},
			expected: nil,
		},
		{
			name: "single empty slice",
			inputs: []*[]string{
				{},
			},
			expected: nil,
		},
		{
			name: "single slice with values",
			inputs: []*[]string{
				{"a", "b", "c"},
			},
			expected: &[]string{"a", "b", "c"},
		},
		{
			name: "multiple slices with duplicates",
			inputs: []*[]string{
				{"a", "b"},
				{"b", "c"},
				{"a", "d"},
			},
			expected: &[]string{"a", "b", "c", "d"},
		},
		{
			name: "mix of nil and non-nil slices",
			inputs: []*[]string{
				nil,
				{"a", "b"},
				nil,
				{"b", "c"},
			},
			expected: &[]string{"a", "b", "c"},
		},
		{
			name: "unsorted input should return sorted output",
			inputs: []*[]string{
				{"z", "y"},
				{"b", "a"},
				{"d", "c"},
			},
			expected: &[]string{"a", "b", "c", "d", "y", "z"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeStringSlicePointers(tt.inputs...)

			// Check if both are nil
			if result == nil && tt.expected == nil {
				return
			}

			// Check if one is nil but not the other
			if (result == nil && tt.expected != nil) || (result != nil && tt.expected == nil) {
				t.Errorf("MergeStringSlicePointers() = %v, want %v", result, tt.expected)
				return
			}

			// Compare actual values
			if !reflect.DeepEqual(*result, *tt.expected) {
				t.Errorf("MergeStringSlicePointers() = %v, want %v", *result, *tt.expected)
			}
		})
	}
}
