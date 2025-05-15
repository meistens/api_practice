package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

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

	// call Get() to fetch the data fora specific movie
	// also add a errors.is() to know if it returned an error so as to send a 404
	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// create an envelope{"movie": movie} instance and pass it to wrtiejson()
	// instead of passing the plain movie struct
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// updteMovieHandler for PUT endpoint
func (app *application) updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	// extract movie id from url
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// fetch existing movie record from the db, sending a 404 if none found
	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// if request contains a x-expected-version header, verify that the movie
	// version in the db matches the expected versions specified
	if r.Header.Get("X-Expected-Version") != "" {
		if strconv.FormatInt(int64(movie.Version), 32) != r.Header.Get("X-Expected-Version") {
			app.editConflictResponse(w, r)
			return
		}
	}

	// declare an input struct to hold the expected data from the client
	// also, use pointers to allow for partial update of a particular field
	// instead of all fields if necessary
	var input struct {
		Title   *string       `json:"title"`
		Year    *int32        `json:"year"`
		Runtime *data.Runtime `json:"runtime"`
		Genres  []string      `json:"genres"`
	}
	// read the json request body data into the input struct
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// copy the values from the request body to the appropriate fields of the
	// movie record
	// due to use of ptrs, they are dereferenced using * to get the underlying value
	// before assigning to the movie record
	if input.Title != nil {
		movie.Title = *input.Title
	}
	if input.Year != nil {
		movie.Year = *input.Year
	}
	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}
	if input.Genres != nil {
		movie.Genres = input.Genres
	}

	// validate
	v := validator.New()
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// pass the updated record to the update() method
	err = app.models.Movies.Update(movie)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// write updated record in a json response
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// delete endpoint
func (app *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	// extract id from url
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// delete movie from db, sending 404 if no matching record
	err = app.models.Movies.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// return 200
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "movie successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// listMoviesHandler endpoint
func (app *application) listMoviesHandler(w http.ResponseWriter, r *http.Request) {
	// define an input struct to hold the values from the request query string
	// plus to differentiate it from the other structs used by other handlers
	var input struct {
		Title        string
		Genres       []string
		data.Filters // future me sees this, make a folder of debuffs for different classes similar to this, you know what to do when you see this
	}

	// new validator instance
	v := validator.New()

	// call r.URL.query() to get the url.Values map containing the query string data
	qs := r.URL.Query()

	// use helpers read* to extract title and genres query string values, falling back
	// to defaults - empty string and slice respectively - if they are not provided
	// by the client
	input.Title = app.readString(qs, "title", "")
	input.Genres = app.readCSV(qs, "genres", []string{})

	// get the page and page_size query string values as ints
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	// extract the sort query string value, falling back to 'id' if not provided
	// by the client
	input.Filters.Sort = app.readString(qs, "sort", "id")
	// add supported sort values for this endpoint to the sort safelist
	input.Filters.SortSafelist = []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}

	// execute validation checks on the Filters struct and send a response
	// containing any errors
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// call getall() to retrieve movies, passing in the various filter params
	movies, metadata, err := app.models.Movies.GetAll(input.Title, input.Genres, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// send json response containing movie data
	err = app.writeJSON(w, http.StatusOK, envelope{"movies": movies, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
