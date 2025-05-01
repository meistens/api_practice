package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// define an error for the unmarshaljson() to return if
// unable to parse
var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

// declare custom runtime with type int32
type Runtime int32

// implement marshaljson() method on runtime type so that it
// satisfies the json.marshal interface
func (r Runtime) MarshalJSON() ([]byte, error) {
	// generating a string containing the movie runtime in the required format
	jsonValue := fmt.Sprintf("%d mins", r)

	// strconv.Quote() to wrap in double quotes
	quotedJSONValue := strconv.Quote(jsonValue)

	// convert quoted string value to a byte slice
	return []byte(quotedJSONValue), nil
}

// implement a unmarshaljson() method on the Runtime type so that
// it satisfies the json.unmarshaler interface
func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {
	// expecting a json value in the form of a string "<runtime>mins"
	// so the surrounding quotes are removed
	// if it cannot be unquoted, throw an error
	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// split the string to isolate the part containing the number
	parts := strings.Split(unquotedJSONValue, " ")

	// check parts of the string to make sure it was in expected format
	// if not, return the invalidruntime error
	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}

	// otherwise parse the string containing the number into an int32
	// if it fails, return same error
	i, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// convert the int32 to a runtime type and assign to the receiver
	*r = Runtime(i)

	return nil
}
