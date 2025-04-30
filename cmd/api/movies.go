package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/meistens/api_practice/internal/data"
)

// add createMovieHandler for the POST /v1/movies endpoint
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	// declare an anon struct to hold info expected to be in the http request body
	var input struct {
		Title   string   `json:"title"`
		Year    int32    `json:"year"`
		Runtime int32    `json:"runtime"`
		Genres  []string `json:"genres"`
	}

	// init a new json.decoder() instance which reads from the request body
	// us decode() method to decode the body contents into the input struct
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}
	fmt.Fprintf(w, "%+v\n", input)
}

// showMovieHandler made simpler via readIDParams() in cmd/api/helpers.go
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// create new instance of the Movie struct
	movie := data.Movie{
		ID:        id,
		CreatedAt: time.Now(),
		Title:     "Ted",
		Runtime:   120,
		Genres:    []string{"drama", "romance", "war"},
		Version:   1,
	}

	// create an envelope{"movie": movie} instance and pass it to wrtiejson()
	// instead of passing the plain movie struct
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
