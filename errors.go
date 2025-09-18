package form

import "errors"

var ErrNotImplemented error = errors.New("not implemented yet")
var ErrNotPtrToStruct error = errors.New("dst must be a pointer to a structure")
var ErrTypeParsingNotSlice error = errors.New("given array contain more than one value on scalar type")
