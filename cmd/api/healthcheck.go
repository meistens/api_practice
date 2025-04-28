package main

import (
	"net/http"
)

// declare a handler which writes a plain-txt res with info
// about the app status, operating env and version
func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	// map of information that is needed to send in the response
	data := map[string]string{
		"status":      "available",
		"environment": app.config.env,
		"version":     version,
	}

	err := app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		app.logger.Println(err)
		http.Error(w, "Server encountered a problem, request unsuccessful", http.StatusInternalServerError)
	}
}
