package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	// init new router instance
	router := httprouter.New()

	// convert the notFoundResponse() helper to a http.Handler using HandlerFunc() adapter
	// then set as a custom err handler for 404
	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	// same shii, with other responses too
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	// register relevant methods, patterns and handler funcs for
	// the endpoints using handlerfunc()
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.updateMovieHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.deleteMovieHandler)
	router.HandlerFunc(http.MethodGet, "/v1/movies", app.listMoviesHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)

	// users
	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)

	// authentication
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthTokenHandler)

	// return the httprouter instance, but with the panic recovery
	// middleware, but now wrapped with ratelimit middleware
	return app.recoverPanic(app.rateLimit(router))
}
