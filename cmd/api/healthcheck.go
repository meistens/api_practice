package main

import (
	"fmt"
	"net/http"
)

// declare a handler which writes a plain-txt res with info
// about the app status, operating env and version
func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	// creating a fixed-format JSON res from a string literal
	js := `{"status": "available", "environment": %q, "version": %q}`
	js = fmt.Sprintf(js, app.config.env, version)

	// setting 'content-type: application/json' header on the res
	w.Header().Set("Content-Type", "application/json")

	// write JSON as http res body
	w.Write([]byte(js))
}
