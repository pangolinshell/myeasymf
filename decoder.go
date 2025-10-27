package form

import (
	"fmt"
	"net/http"
	"reflect"
)

type Decoder struct {
	r *http.Request
}

func NewDecoder(r *http.Request) *Decoder {
	return &Decoder{r: r}
}

// Decode populates the fields of the destination struct 'dst' with values from the provided multipart.Form 'src'.
// It matches form keys to struct fields using custom tag logic (via getByTags), and converts string values to the
// appropriate field types using the Convert function. For file fields, it assigns []*multipart.FileHeader directly.
// The destination must be a pointer to a struct. Returns an error if types are incompatible, fields cannot be set,
// or if the destination is not a pointer to a struct. Currently, always returns ErrNotImplemented at the end.
func (d *Decoder) Decode(dst any) error {
	if dst == nil {
		return fmt.Errorf("destination is nil")
	}

	ptrType := reflect.TypeOf(dst)
	if ptrType.Kind() != reflect.Ptr {
		return ErrNotPtrToStruct
	}

	structType := ptrType.Elem()
	if structType.Kind() != reflect.Struct {
		return ErrNotPtrToStruct
	}

	structValue := reflect.ValueOf(dst).Elem()
	src := d.r.MultipartForm

	for key, value := range src.Value {
		var tagIndex int
		if len(value) < 1 {
			continue
		}
		if i := getByTags(structType, key); i != nil {
			tagIndex = *i
		} else {
			continue
		}

		fieldType := structType.Field(tagIndex).Type

		fieldVal := structValue.Field(tagIndex)

		if !fieldVal.CanSet() {
			return fmt.Errorf("cannot set field %s", structType.Field(tagIndex).Name)
		}

		newValue, err := convert(value, fieldType)
		if err != nil {
			return err
		}

		rv, err := checkAndConvert(fieldVal.Type(), newValue)
		if err != nil {
			return fmt.Errorf("champ %s: %w", structType.Field(tagIndex).Name, err)
		}

		fieldVal.Set(rv)

	}
	for key, value := range src.File {
		var tagIndex int
		if i := getByTags(structType, key); i != nil {
			tagIndex = *i
		} else {
			continue
		}

		fieldVal := structValue.Field(tagIndex)
		if !fieldVal.CanSet() {
			return fmt.Errorf("cannot set field %s", structType.Field(tagIndex).Name)
		}

		fieldType := fieldVal.Type()

		switch fieldType.Kind() {
		case reflect.Slice:
			// Handles []*multipart.FileHeader
			rv, err := checkAndConvert(fieldType, value)
			if err != nil {
				return fmt.Errorf("field %s: %w", structType.Field(tagIndex).Name, err)
			}
			fieldVal.Set(rv)

		case reflect.Pointer:
			// Handles *multipart.FileHeader
			if len(value) == 0 {
				continue
			}
			if len(value) > 1 {
				return fmt.Errorf("field %s: %w", structType.Field(tagIndex).Name, ErrMultipleFilesForSingleField)
			}
			rv, err := checkAndConvert(fieldType, value[0])
			if err != nil {
				return fmt.Errorf("field %s: %w", structType.Field(tagIndex).Name, err)
			}
			fieldVal.Set(rv)

		default:
			return fmt.Errorf("field %s: %w", structType.Field(tagIndex).Name, ErrUnsupportedFileFieldType)
		}
	}

	return nil
}
