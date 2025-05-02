package validator

import (
	"regexp"
)

// declare a regex for sanity checks of email addresses
var (
	EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

// declare a new validator type which contains a map of validation errors
type Validator struct {
	Errors map[string]string
}

// New is a helper which creates a new validator instance with
// an empty errors map
func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// valid returns true if errors map doesn't contain any entries
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// adderr adds an error msg to the map
func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// check adds an error msg to the map only if a validation
// check is not ok
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// In returns true if a specific value is in a list of strings
func In(value string, list ...string) bool {
	for i := range list {
		if value == list[i] {
			return true
		}
	}
	return false
}

// matches returns true if a string value matches a specific regex
// pattern
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// unique returns true if all string values in a slice are unique
func Unique(values []string) bool {
	uniqueValues := make(map[string]bool)

	for _, value := range values {
		uniqueValues[value] = true
	}
	return len(values) == len(uniqueValues)
}
