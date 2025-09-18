package form

import "strings"

// Tag specifies the struct field tag key used for form parsing.
var Tag string = "form"

// parseTagStr parses a struct tag string and returns the field name and a boolean indicating
// whether the "omitempty" option is set. The tag string is expected to be in the format "name"
// or "name,omitempty". If the input string is empty, it returns an empty name and false for omitEmpty.
func parseTagStr(str string) (name string, omitEmpty bool) {
	if str == "" {
		return "", false
	}
	parts := strings.Split(str, ",")
	if len(parts) == 1 {
		name = parts[0]
		omitEmpty = false
	} else if parts[1] == "omitempty" {
		name = parts[0]
		omitEmpty = true
	}
	return
}
