package form

import (
	"errors"
)

var (
	ErrNotImplemented              error = errors.New("not implemented yet")
	ErrNotPtrToStruct              error = errors.New("dst must be a pointer to a structure")
	ErrTypeParsingNotSlice         error = errors.New("given array contain more than one value on scalar type")
	ErrMultipleFilesForSingleField error = errors.New("multiple files provided for single file field")
	ErrUnsupportedFileFieldType    error = errors.New("unsupported file field type")
)
