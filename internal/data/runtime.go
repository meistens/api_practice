package data

import (
	"fmt"
	"strconv"
)

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
