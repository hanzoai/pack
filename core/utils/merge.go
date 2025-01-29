package utils

import (
	"reflect"
)

// MergeStructs merges multiple structs of the same type, with later values taking precedence.
// Only non-zero values from later structs will override earlier values.
func MergeStructs(dst interface{}, srcs ...interface{}) {
	for _, src := range srcs {
		if src == nil {
			continue
		}
		mergeStruct(dst, src)
	}
}

// mergeStruct merges two structs of the same type, taking non-zero values from src
func mergeStruct(dst, src interface{}) {
	dstValue := reflect.ValueOf(dst).Elem()
	srcValue := reflect.ValueOf(src)

	if srcValue.Kind() == reflect.Ptr {
		srcValue = srcValue.Elem()
	}

	// If it's not a struct, we can't iterate over fields
	if srcValue.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < srcValue.NumField(); i++ {
		srcField := srcValue.Field(i)
		dstField := dstValue.Field(i)

		switch srcField.Kind() {

		// Merge map key/values
		case reflect.Map:
			if !srcField.IsNil() {
				if dstField.IsNil() {
					dstField.Set(reflect.MakeMap(srcField.Type()))
				}
				for _, key := range srcField.MapKeys() {
					srcValue := srcField.MapIndex(key)
					dstValue := dstField.MapIndex(key)

					// If the map values are pointers to structs, merge them
					if srcValue.Kind() == reflect.Ptr && srcValue.Elem().Kind() == reflect.Struct {
						if !dstValue.IsValid() {
							// If destination doesn't have this key, create new struct
							dstField.SetMapIndex(key, srcValue)
						} else {
							// If destination has this key, merge the structs
							mergeStruct(dstValue.Interface(), srcValue.Interface())
						}
					} else {
						// For non-struct pointers or other types, just set the value
						dstField.SetMapIndex(key, srcValue)
					}
				}
			}

		// Use any non-nil slice to replace the destination slice
		// This includes empty slices
		case reflect.Slice:
			if !srcField.IsNil() {
				dstField.Set(srcField)
			}

		case reflect.Ptr:
			if !srcField.IsNil() {
				// If it's a pointer to a struct, we need to merge the structs
				if srcField.Elem().Kind() == reflect.Struct {
					if dstField.IsNil() {
						// If destination is nil, create a new struct
						dstField.Set(reflect.New(srcField.Elem().Type()))
					}

					// Recursively merge the struct contents
					mergeStruct(dstField.Interface(), srcField.Interface())
				} else {
					// For non-struct pointers (like *[]Command), only set if source points to non-nil value
					if dstField.IsNil() || !srcField.Elem().IsNil() {
						dstField.Set(srcField)
					}
				}
			}

		// Recursively merge nested structs
		case reflect.Struct:
			mergeStruct(dstField.Addr().Interface(), srcField.Interface())

		// Use any non-zero value to replace the destination value
		default:
			if !isZeroValue(srcField) {
				dstField.Set(srcField)
			}
		}
	}
}

// isZeroValue checks if a reflect.Value is its zero value
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	case reflect.Struct:
		// For structs, check if all fields are zero values
		for i := 0; i < v.NumField(); i++ {
			if !isZeroValue(v.Field(i)) {
				return false
			}
		}
		return true
	}
	return false
}
