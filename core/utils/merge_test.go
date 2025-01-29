package utils

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

type DeepNestedStruct struct {
	Level1 struct {
		Value  string
		Level2 NestedStruct
		IntMap map[string]int
	}
}

type TestStruct struct {
	String       string
	Int          int
	Float        float64
	Bool         bool
	StringPtr    *string
	IntSlice     []int
	StringMap    map[string]string
	Nested       NestedStruct
	Deep         DeepNestedStruct
	CommandSlice *[]string
}

type NestedStruct struct {
	Value    string
	IntValue int
}

func TestMergeStructs(t *testing.T) {
	str1 := "ptr1"
	str2 := "ptr2"

	tests := []struct {
		name     string
		dst      TestStruct
		src      TestStruct
		expected TestStruct
	}{
		{
			name: "merge basic types",
			dst: TestStruct{
				String: "old",
				Int:    1,
				Float:  1.0,
				Bool:   false,
			},
			src: TestStruct{
				String: "new",
				Int:    2,
				Float:  0, // zero value, should not override
				Bool:   true,
			},
			expected: TestStruct{
				String: "new",
				Int:    2,
				Float:  1.0,
				Bool:   true,
			},
		},
		{
			name: "merge pointers",
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
			name: "merge slices - empty source should replace",
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
			name: "merge slices - nil source should keep destination",
			dst: TestStruct{
				IntSlice: []int{1, 2},
			},
			src: TestStruct{}, // IntSlice will be nil
			expected: TestStruct{
				IntSlice: []int{1, 2},
			},
		},
		{
			name: "merge slices - non-empty source should replace",
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
			name: "merge maps",
			dst: TestStruct{
				StringMap: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
			src: TestStruct{
				StringMap: map[string]string{
					"key2": "new_value2",
					"key3": "value3",
				},
			},
			expected: TestStruct{
				StringMap: map[string]string{
					"key1": "value1",
					"key2": "new_value2",
					"key3": "value3",
				},
			},
		},
		{
			name: "merge nested structs",
			dst: TestStruct{
				Bool: true,
				Nested: NestedStruct{
					Value:    "old",
					IntValue: 1,
				},
			},
			src: TestStruct{
				Nested: NestedStruct{
					Value: "new",
				},
			},
			expected: TestStruct{
				Bool: true,
				Nested: NestedStruct{
					Value:    "new",
					IntValue: 1,
				},
			},
		},
		{
			name: "merge maps - nil source should keep destination",
			dst: TestStruct{
				StringMap: map[string]string{
					"key1": "value1",
				},
			},
			src: TestStruct{}, // StringMap will be nil
			expected: TestStruct{
				StringMap: map[string]string{
					"key1": "value1",
				},
			},
		},
		{
			name: "merge maps - empty source should merge",
			dst: TestStruct{
				StringMap: map[string]string{
					"key1": "value1",
				},
			},
			src: TestStruct{
				StringMap: map[string]string{},
			},
			expected: TestStruct{
				StringMap: map[string]string{
					"key1": "value1",
				},
			},
		},
		{
			name: "merge maps - should merge not replace",
			dst: TestStruct{
				StringMap: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
			src: TestStruct{
				StringMap: map[string]string{
					"key2": "new_value2",
					"key3": "value3",
				},
			},
			expected: TestStruct{
				StringMap: map[string]string{
					"key1": "value1",
					"key2": "new_value2",
					"key3": "value3",
				},
			},
		},
		{
			name: "merge deep nested structs",
			dst: TestStruct{
				Deep: DeepNestedStruct{
					Level1: struct {
						Value  string
						Level2 NestedStruct
						IntMap map[string]int
					}{
						Value: "old_level1",
						Level2: NestedStruct{
							Value:    "old_level2",
							IntValue: 1,
						},
						IntMap: map[string]int{
							"a": 1,
							"b": 2,
						},
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
						Value: "new_level1",
						Level2: NestedStruct{
							Value: "new_level2",
						},
						IntMap: map[string]int{
							"b": 20,
							"c": 30,
						},
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
						Value: "new_level1",
						Level2: NestedStruct{
							Value:    "new_level2",
							IntValue: 1,
						},
						IntMap: map[string]int{
							"a": 1,
							"b": 20,
							"c": 30,
						},
					},
				},
			},
		},
		{
			name: "merge zero value struct should not override",
			dst: TestStruct{
				Nested: NestedStruct{
					Value:    "keep",
					IntValue: 42,
				},
			},
			src: TestStruct{
				Nested: NestedStruct{}, // zero value struct
			},
			expected: TestStruct{
				Nested: NestedStruct{
					Value:    "keep",
					IntValue: 42,
				},
			},
		},
		{
			name: "merge pointer to slice - nil source should keep destination",
			dst: TestStruct{
				CommandSlice: &[]string{"cmd1", "cmd2"},
			},
			src: TestStruct{
				CommandSlice: nil,
			},
			expected: TestStruct{
				CommandSlice: &[]string{"cmd1", "cmd2"},
			},
		},
		{
			name: "merge pointer to slice - empty source should replace",
			dst: TestStruct{
				CommandSlice: &[]string{"cmd1", "cmd2"},
			},
			src: TestStruct{
				CommandSlice: &[]string{},
			},
			expected: TestStruct{
				CommandSlice: &[]string{},
			},
		},
		{
			name: "merge pointer to slice - non-empty source should replace",
			dst: TestStruct{
				CommandSlice: &[]string{"cmd1", "cmd2"},
			},
			src: TestStruct{
				CommandSlice: &[]string{"cmd3", "cmd4"},
			},
			expected: TestStruct{
				CommandSlice: &[]string{"cmd3", "cmd4"},
			},
		},
		{
			name: "merge pointer to slice - nil destination should take source",
			dst: TestStruct{
				CommandSlice: nil,
			},
			src: TestStruct{
				CommandSlice: &[]string{"cmd1", "cmd2"},
			},
			expected: TestStruct{
				CommandSlice: &[]string{"cmd1", "cmd2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.dst
			MergeStructs(&result, &tt.src)

			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf("merge structs mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMergeStructsMultiple(t *testing.T) {
	dst := TestStruct{
		String: "first",
		Int:    1,
	}

	src1 := TestStruct{
		String: "second",
		Float:  2.0,
	}

	src2 := TestStruct{
		String: "third",
		Bool:   true,
	}

	expected := TestStruct{
		String: "third",
		Int:    1,
		Float:  2.0,
		Bool:   true,
	}

	MergeStructs(&dst, &src1, &src2)
	require.Equal(t, expected, dst)
}

func TestMergeStructsNilSource(t *testing.T) {
	dst := TestStruct{
		String: "original",
		Int:    1,
	}

	expected := dst
	MergeStructs(&dst, nil)
	require.Equal(t, expected, dst)
}
