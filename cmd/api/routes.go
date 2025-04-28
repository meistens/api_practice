package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() *httprouter.Router {
	// init new router instance
	router := httprouter.New()

	// register relevant methods, patterns and handler funcs for
	// the endpoints using handlerfunc()
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandleFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)
	router.HandleFunc(http.MethodGet, "v1/movies/:id", app.showMovieHandler)

	// return the httprouter instance
	return router
}
