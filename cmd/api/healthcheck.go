package main

import (
	"encoding/json"
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

	// passing map to the json.marshal() function
	// marshal returns a []byte slice containing encoded json
	// if error, log it and send a generic err msg
	js, err := json.Marshal(data)
	if err != nil {
		app.logger.Println(err)
		http.Error(w, "the server encountered a problem and could not process your request", http.StatusInternalServerError)
		return
	}

	// append a newline to the json
	js = append(js, '\n')

	// at this point, the encoding should work
	w.Header().Set("Content-Type", "application/json")

	// w.write() to send the []byte slice containing json as the response
	w.Write(js)
}
