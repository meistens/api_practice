package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/meistens/api_practice/internal/data"
	"github.com/meistens/api_practice/internal/validator"
)

// add createMovieHandler for the POST /v1/movies endpoint
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	// declare an anon struct to hold info expected to be in the http request body
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	// use the new readjson() helper to decode the request body
	// into the input struct
	// if this returns an error, send the client a error msg
	// along with a 400 bad request status code
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	// init. new validator instance
	// check if there are no errors (check validator.go for a list of em)
	v := validator.New()
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// call insert() method on the movies model, passing in a ptr to the
	// validated movie struct
	// this will create a record in the database and update the movie struct
	// with the system-generated info
	err = app.models.Movies.Insert(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// include location header to let the client know which url they can find
	// the newly created resource by making an empty http.Header map and using Set()
	// to include the header
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	// write a json response with a 201 created status code, movie data
	// in the response body, and location header
	err = app.writeJSON(w, http.StatusCreated, envelope{"movie": movie}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
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
