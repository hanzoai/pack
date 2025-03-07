package utils

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
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

func TestParseVersions(t *testing.T) {
	input := []string{
		"basic@1.0.0",
		"caret@^2.4",
		"tilde@~3.1.3",
		"vprefix@v4.0.0",
		"xnotation@14.x",
		"range@>=22 <23",
		"wildcard@*",
	}

	parsedVersions := ParsePackageWithVersion(input)

	expected := map[string]string{
		"basic":     "1.0.0",
		"caret":     "^2.4",
		"tilde":     "~3.1.3",
		"vprefix":   "v4.0.0",
		"xnotation": "14.x",
		"range":     ">=22 <23",
		"wildcard":  "*",
	}

	require.Equal(t, expected, parsedVersions)
}

func TestExtractSemverVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple version", "1.2.3", "1.2.3"},
		{"v prefix", "v1.2.3", "1.2.3"},
		{"major.minor version", "1.2", "1.2"},
		{"major version only", "1", "1"},
		{"invalid format", "1.2.3.4", "1.2.3"},
		{"with prefix", "version1.2.3", "1.2.3"},
		{"with suffix", "1.2.3-beta", "1.2.3"},
		{"empty string", "", ""},
		{"non-numeric", "a.b.c", ""},
		{"mixed format", "v1.2.x", "1.2"},
		{"python style version", "python-3.10.7", "3.10.7"},
		{"version in text", "runtime version is 2.4.1 or higher", "2.4.1"},
		{"multiple versions", "supports both 1.2.3 and 4.5.6", "1.2.3"},
		{"with suffix and prefix", "myapp1.2.3-rc1", "1.2.3"},
		{"major version in text", "python 3 or higher", "3"},
		{"major.minor in text", "requires node 14.2", "14.2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractSemverVersion(tt.input)
			if result != tt.expected {
				t.Errorf("ExtractSemverVersion(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
