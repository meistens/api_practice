package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"maps"

	"github.com/julienschmidt/httprouter"
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
	// decode request body into the target dest
	err := json.NewDecoder(r.Body).Decode(dest)
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

		// if a non-nil pointer is passed to decode()
		// return a invalid unmarshal error
		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		// anything else, return error msg as-is
		default:
			return err
		}
	}
	return nil
}
