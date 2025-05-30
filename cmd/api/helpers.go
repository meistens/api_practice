package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"maps"

	"github.com/julienschmidt/httprouter"
	"github.com/meistens/api_practice/internal/validator"
)

// define an envelope type
type envelope map[string]any

// retrieve the 'id' url param from the current req. context
// convert it to an integer and return it
// if unsuccessful, return 0 and error
func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id param")
	}
	return id, nil
}

// defining a writejson() helper for sending responses
// takes the dest responsewriter, status code to send, data to encode, and  header map
// containing any additional http requests to be inccluded in res
func (app *application) writeJSON(w http.ResponseWriter, status int, data any, headers http.Header) error {
	// encode data to json, returning error if there was one
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	// append a newline for terminal applications
	js = append(js, '\n')

	maps.Copy(w.Header(), headers)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

// JSON err decoding func
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dest any) error {
	// use http.Maxbytereader() to limit size of the request body to 1MiB
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	// init json.decoder
	// call disallowunknownfields() method on it before decoding
	// this means if json from body included any field that cannot be mapped
	// the decoder will return an error instead of ignoring
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	// decode request body into the target dest
	err := dec.Decode(dest)
	if err != nil {
		// if there is an error during decoding, start trige
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		// use errors.As() to check if the error has the
		// *json.syntaxerror type
		// if it does, return a plain-english error message
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d", syntaxError.Offset)

		// check for unexpected eof by using errors.is() and
		// return a generic error msg
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		// check for unmarshal type errors
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect json type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type(at character %d)", unmarshalTypeError.Offset)

		// if request body is empty, iof.error() will be
		// returned by decode()
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		// if json contains a field which cannot be mapped
		// to target dest, decode() will return an error msg in the format
		// "json:unknown field: "<name>""
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		// if request body exceeds 1MiB, decode will fail with
		// req body too large error
		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)

		// if a non-nil pointer is passed to decode()
		// return a invalid unmarshal error
		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		// anything else, return error msg as-is
		default:
			return err
		}
	}

	// call decode() again using a ptr to an empty anon struct
	// as the dest.
	// if body = single json, return eof error, else return stuff
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single json value")
	}
	return nil
}

// readString helper func.
// returns a string value from the query string, or the provided default value
// if no matching key is found
func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	// extract the value from a given key from the query string
	// if no key exists, return the empty string
	s := qs.Get(key)

	// if no key exists, return the default value
	if s == "" {
		return defaultValue
	}
	// else return the string
	return s
}

// readCSV helper func.
// reads a string value from the query string, splits it into a slice on the comma char
// if no matching key, return the provided default value
func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	// extract value from query string
	csv := qs.Get(key)

	// if no key exists, or value is empty, return default value
	if csv == "" {
		return defaultValue
	}
	// otherwise parse the value into a []string slice and return the default value
	return strings.Split(csv, ",")
}

// readInt helper reads a stringvalue from the query string, converts it to an
// int before returning
// if no matching key found, return the provided default value
// if it cannot be converted, record err msg in provided validator instance
func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	// extract value from query string
	s := qs.Get(key)

	// if no key, or is empty, return default
	if s == "" {
		return defaultValue
	}
	// convert value to an int
	// if it fails, add an err msg to the validator instance
	// return the default value
	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an int value")
		return defaultValue
	}
	// otherwise return the converted int value
	return i
}

// catching panics from background goroutines
// background helper accepts an arbitary function as a param
func (app *application) background(fn func()) {
	// increment waitgroup() counter
	app.wg.Add(1)

	// launch background goroutine
	go func() {
		// defer to decrement waitgroup counter before the goroutine returns
		defer app.wg.Done()

		// recover any panic
		defer func() {
			if err := recover(); err != nil {
				app.logger.PrintError(fmt.Errorf("%s", err), nil)
			}
		}()
		// execute the arbitrary function that was passed as the param
		fn()
	}()
}
