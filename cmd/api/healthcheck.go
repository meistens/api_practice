package main

import (
	"fmt"
	"net/http"
)

// declare a handler which writes a plain-txt res with info
// about the app status, operating env and version
func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "status: available")
	fmt.Fprintf(w, "environment: %s\n", app.config.env)
	fmt.Fprintf(w, "version: %s\n", version)
}
