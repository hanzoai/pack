package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type TestStruct struct {
	// Basic types
	String string
	Int    int
	Bool   bool

	// Special types
	StringPtr    *string
	IntSlice     []int
	StringMap    map[string]string
	CommandSlice *[]string

	// Nested structs
	Nested NestedStruct
	Deep   DeepNestedStruct
}

type NestedStruct struct {
	Value    string
	IntValue int
}

type DeepNestedStruct struct {
	Level1 struct {
		Value  string
		Level2 NestedStruct
		IntMap map[string]int
	}
}

func TestMergeBasicTypes(t *testing.T) {
	str1 := "value1"
	str2 := "value2"

	tests := []struct {
		name     string
		dst      TestStruct
		src      TestStruct
		expected TestStruct
	}{
		{
			name: "zero values should not override",
			dst: TestStruct{
				String: "keep",
				Int:    42,
			},
			src: TestStruct{
				String: "",    // zero value
				Int:    0,     // zero value
				Bool:   false, // zero value
			},
			expected: TestStruct{
				String: "keep",
				Int:    42,
			},
		},
		{
			name: "non-zero values should override",
			dst: TestStruct{
				String: "old",
				Int:    1,
			},
			src: TestStruct{
				String: "new",
				Int:    2,
				Bool:   true,
			},
			expected: TestStruct{
				String: "new",
				Int:    2,
				Bool:   true,
			},
		},
		{
			name: "pointers should override when non-nil",
			dst: TestStruct{
				StringPtr: &str1,
			},
			src: TestStruct{
				StringPtr: &str2,
			},
			expected: TestStruct{
				StringPtr: &str2,
			},
		},
		{
			name: "nil pointers should not override",
			dst: TestStruct{
				StringPtr: &str1,
			},
			src: TestStruct{}, // nil pointer
			expected: TestStruct{
				StringPtr: &str1,
			},
		},
	}

	runTests(t, tests)
}

func TestMergeSlices(t *testing.T) {
	tests := []struct {
		name     string
		dst      TestStruct
		src      TestStruct
		expected TestStruct
	}{
		{
			name: "nil slice should not override",
			dst: TestStruct{
				IntSlice: []int{1, 2},
			},
			src: TestStruct{}, // nil slice
			expected: TestStruct{
				IntSlice: []int{1, 2},
			},
		},
		{
			name: "empty slice should override",
			dst: TestStruct{
				IntSlice: []int{1, 2},
			},
			src: TestStruct{
				IntSlice: []int{},
			},
			expected: TestStruct{
				IntSlice: []int{},
			},
		},
		{
			name: "non-empty slice should override",
			dst: TestStruct{
				IntSlice: []int{1, 2},
			},
			src: TestStruct{
				IntSlice: []int{3, 4},
			},
			expected: TestStruct{
				IntSlice: []int{3, 4},
			},
		},
		{
			name: "pointer to slice - nil should not override",
			dst: TestStruct{
				CommandSlice: &[]string{"cmd1"},
			},
			src: TestStruct{}, // nil pointer
			expected: TestStruct{
				CommandSlice: &[]string{"cmd1"},
			},
		},
		{
			name: "pointer to slice - empty should override",
			dst: TestStruct{
				CommandSlice: &[]string{"cmd1"},
			},
			src: TestStruct{
				CommandSlice: &[]string{},
			},
			expected: TestStruct{
				CommandSlice: &[]string{},
			},
		},
	}

	runTests(t, tests)
}

func TestMergeMaps(t *testing.T) {
	tests := []struct {
		name     string
		dst      TestStruct
		src      TestStruct
		expected TestStruct
	}{
		{
			name: "nil map should not override",
			dst: TestStruct{
				StringMap: map[string]string{"key1": "value1"},
			},
			src: TestStruct{}, // nil map
			expected: TestStruct{
				StringMap: map[string]string{"key1": "value1"},
			},
		},
		{
			name: "empty map should merge",
			dst: TestStruct{
				StringMap: map[string]string{"key1": "value1"},
			},
			src: TestStruct{
				StringMap: map[string]string{},
			},
			expected: TestStruct{
				StringMap: map[string]string{"key1": "value1"},
			},
		},
		{
			name: "maps should merge values",
			dst: TestStruct{
				StringMap: map[string]string{
					"keep":     "old",
					"override": "old",
				},
			},
			src: TestStruct{
				StringMap: map[string]string{
					"override": "new",
					"add":      "new",
				},
			},
			expected: TestStruct{
				StringMap: map[string]string{
					"keep":     "old",
					"override": "new",
					"add":      "new",
				},
			},
		},
	}

	runTests(t, tests)
}

func TestMergeNestedStructs(t *testing.T) {
	tests := []struct {
		name     string
		dst      TestStruct
		src      TestStruct
		expected TestStruct
	}{
		{
			name: "zero value nested struct should not override",
			dst: TestStruct{
				Nested: NestedStruct{
					Value:    "keep",
					IntValue: 42,
				},
			},
			src: TestStruct{
				Nested: NestedStruct{}, // zero value
			},
			expected: TestStruct{
				Nested: NestedStruct{
					Value:    "keep",
					IntValue: 42,
				},
			},
		},
		{
			name: "nested struct should merge fields",
			dst: TestStruct{
				Deep: DeepNestedStruct{
					Level1: struct {
						Value  string
						Level2 NestedStruct
						IntMap map[string]int
					}{
						Value: "old",
						Level2: NestedStruct{
							Value:    "old",
							IntValue: 1,
						},
						IntMap: map[string]int{"a": 1},
					},
				},
			},
			src: TestStruct{
				Deep: DeepNestedStruct{
					Level1: struct {
						Value  string
						Level2 NestedStruct
						IntMap map[string]int
					}{
						Value: "new",
						Level2: NestedStruct{
							Value: "new",
						},
						IntMap: map[string]int{"b": 2},
					},
				},
			},
			expected: TestStruct{
				Deep: DeepNestedStruct{
					Level1: struct {
						Value  string
						Level2 NestedStruct
						IntMap map[string]int
					}{
						Value: "new",
						Level2: NestedStruct{
							Value:    "new",
							IntValue: 1,
						},
						IntMap: map[string]int{
							"a": 1,
							"b": 2,
						},
					},
				},
			},
		},
	}

	runTests(t, tests)
}

func runTests(t *testing.T, tests []struct {
	name     string
	dst      TestStruct
	src      TestStruct
	expected TestStruct
}) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.dst
			MergeStructs(&result, &tt.src)
			require.Equal(t, tt.expected, result)
		})
	}
}
