package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"maps"

	"github.com/julienschmidt/httprouter"
)

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
	js, err := json.Marshal(data)
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
