package myeasyform

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type kind uint

const (
	Invalid kind = iota
	Bool
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	Uintptr
	Float32
	Float64
	Complex64
	Complex128
	Array
	Chan
	Func
	Interface
	Map
	Pointer
	Slice
	String
	Struct
	UnsafePointer

	Time
	FileHeader
)

// getByTags searches for a struct field within the provided reflect.Type whose tag value matches the given key.
// It returns a pointer to the index of the matching field if found, or nil otherwise.
// The comparison is case-insensitive and uses the tagName constant to retrieve the tag value.
// The tag value is parsed using parseTagStr to extract the field name for comparison.
func getByTags(t reflect.Type, key string) (index *int) {
	var numField = t.NumField()
	for i := range numField {
		field := t.Field(i)
		tagContent := field.Tag.Get(tagName)
		name, _ := parseTagStr(tagContent)
		if name == strings.ToLower(key) || name == key {
			return &i
		}
	}
	return nil
}

func getKind(t reflect.Type) kind {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.String() {
	case "time.Time":
		return Time
	}

	if t.Kind() != reflect.Invalid {
		return kind(t.Kind())
	}
	return Invalid
}

// Convert attempts to convert a slice of strings `v` into a value of the specified reflect.Type `t`.
// It supports conversion to basic Go types (bool, int, uint, float, string), time.Time, and slices/arrays of these types.
// If `t` is a pointer type, the returned value will be a pointer to the converted value.
// For slices and arrays, each element in `v` is recursively converted to the element type.
// Returns the converted value as `any`, or an error if conversion fails or the input is invalid.
//
// Supported types:
//   - bool, *bool
//   - int, int8, int16, int32, int64 and their pointer types
//   - uint, uint8, uint16, uint32, uint64 and their pointer types
//   - float32, float64 and their pointer types
//   - string, *string
//   - time.Time, *time.Time
//   - slices and arrays of the above types
//
// Errors:
//   - Returns an error if the input slice length is not 1 for scalar types.
//   - Returns an error if parsing fails for the target type.
//   - Returns an error if the type is not supported.
//
// Example usage:
//
//	val, err := Convert([]string{"42"}, reflect.TypeOf(int(0)))
func convert(v []string, t reflect.Type) (any, error) {
	isPtr := t.Kind() == reflect.Ptr
	baseType := t
	if isPtr {
		baseType = t.Elem()
	}

	k := getKind(t)

	switch k {
	// ─────────────── Booleans ───────────────
	case Bool:
		if len(v) != 1 {
			return nil, fmt.Errorf("%s on field %s", ErrTypeParsingNotSlice, baseType.Name())
		}
		b, err := strconv.ParseBool(v[0])
		if err != nil {
			return nil, err
		}
		if isPtr {
			return &b, nil
		}
		return b, nil

	// ─────────────── Integers ───────────────
	case Int, Int8, Int16, Int32, Int64:
		if len(v) != 1 {
			return nil, fmt.Errorf("%s on field %s", ErrTypeParsingNotSlice, baseType.Name())
		}
		bitsize := []int{0, 8, 16, 32, 64}
		val, err := strconv.ParseInt(v[0], 0, bitsize[k-2])
		if err != nil {
			return nil, err
		}
		if isPtr {
			tmp := reflect.New(baseType)
			tmp.Elem().SetInt(val)
			return tmp.Interface(), nil
		}
		return val, nil

	// ─────────────── unsigned ───────────────
	case Uint, Uint8, Uint16, Uint32, Uint64:
		if len(v) != 1 {
			return nil, fmt.Errorf("%s on field %s", ErrTypeParsingNotSlice, baseType.Name())
		}
		bitsize := []int{0, 8, 16, 32, 64}
		val, err := strconv.ParseUint(v[0], 0, bitsize[k-7])
		if err != nil {
			return nil, err
		}
		if isPtr {
			tmp := reflect.New(baseType)
			tmp.Elem().SetUint(val)
			return tmp.Interface(), nil
		}
		return val, nil

	// ─────────────── Floats ───────────────
	case Float32, Float64:
		if len(v) != 1 {
			return nil, fmt.Errorf("%s on field %s", ErrTypeParsingNotSlice, baseType.Name())
		}
		val, err := strconv.ParseFloat(v[0], 64)
		if err != nil {
			return nil, err
		}
		if isPtr {
			tmp := reflect.New(baseType)
			tmp.Elem().SetFloat(val)
			return tmp.Interface(), nil
		}
		return val, nil

	// ─────────────── Strings ───────────────
	case String:
		if len(v) != 1 {
			return nil, fmt.Errorf("%s on field %s", ErrTypeParsingNotSlice, baseType.Name())
		}
		if isPtr {
			return &v[0], nil
		}
		return v[0], nil

	// ─────────────── time.Time ───────────────
	case Time:
		if len(v) != 1 {
			return nil, fmt.Errorf("%s on field %s", ErrTypeParsingNotSlice, baseType.Name())
		}
		parsed, err := time.Parse(time.RFC3339, v[0])
		if err != nil {
			return nil, err
		}
		if isPtr {
			return &parsed, nil
		}
		return parsed, nil

	// ─────────────── Slices and Arrays ───────────────
	case Array, Slice:
		// on gère tout slice ou array
		elemType := t.Elem()
		// isElemPtr := elemType.Kind() == reflect.Ptr
		// baseElemType := elemType
		// if isElemPtr {
		// 	baseElemType = elemType.Elem()
		// }

		slice := reflect.MakeSlice(t, len(v), len(v))

		for i := 0; i < len(v); i++ {
			// Appelle Convert récursivement sur 1 élément
			elemVal, err := convert([]string{v[i]}, elemType)
			if err != nil {
				return nil, fmt.Errorf("slice element %d: %w", i, err)
			}
			rv := reflect.ValueOf(elemVal)
			if rv.Type().ConvertibleTo(elemType) {
				rv = rv.Convert(elemType)
			}
			slice.Index(i).Set(rv)
		}

		return slice.Interface(), nil
	}

	return nil, fmt.Errorf("invalid type %s", t.String())
}

// checkAndConvert attempts to convert the given value to the specified fieldType using reflection.
// It handles nil values by returning the zero value of the target type, direct type matches,
// convertible types, pointer/value conversions, and slices (recursively converting elements).
// Returns the converted reflect.Value or an error if conversion is not possible.
func checkAndConvert(fieldType reflect.Type, value interface{}) (reflect.Value, error) {
	rv := reflect.ValueOf(value)

	// Nil → zero value
	if !rv.IsValid() || (rv.Kind() == reflect.Ptr && rv.IsNil()) {
		return reflect.Zero(fieldType), nil
	}

	// Exact type
	if rv.Type() == fieldType {
		return rv, nil
	}

	// Direcly convertible
	if rv.Type().ConvertibleTo(fieldType) {
		return rv.Convert(fieldType), nil
	}

	// Expected pointer but provided value
	if fieldType.Kind() == reflect.Ptr && rv.Type().ConvertibleTo(fieldType.Elem()) {
		ptr := reflect.New(fieldType.Elem())
		ptr.Elem().Set(rv.Convert(fieldType.Elem()))
		return ptr, nil
	}

	// Expected value but pointer provided
	if rv.Kind() == reflect.Ptr && rv.Type().Elem().ConvertibleTo(fieldType) && fieldType.Kind() != reflect.Ptr {
		return rv.Elem().Convert(fieldType), nil
	}

	// Slices
	if fieldType.Kind() == reflect.Slice && rv.Kind() == reflect.Slice {
		elemType := fieldType.Elem()
		n := rv.Len()
		slice := reflect.MakeSlice(fieldType, n, n)
		for i := 0; i < n; i++ {
			elemVal := rv.Index(i)
			converted, err := checkAndConvert(elemType, elemVal.Interface())
			if err != nil {
				return reflect.Value{}, fmt.Errorf("slice element %d: %w", i, err)
			}
			slice.Index(i).Set(converted)
		}
		return slice, nil
	}

	return reflect.Value{}, fmt.Errorf("type incompatible: %s -> %s", rv.Type(), fieldType)
}
