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

	// updated middleware on the five /v1/movies** endpoint
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.requireActivatedUser(app.createMovieHandler))
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.requireActivatedUser(app.showMovieHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.requireActivatedUser(app.updateMovieHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.requireActivatedUser(app.deleteMovieHandler))
	router.HandlerFunc(http.MethodGet, "/v1/movies", app.requireActivatedUser(app.listMoviesHandler))

	// updated
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)
	// users
	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)

	// authentication
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthTokenHandler)

	// return the httprouter instance, but with the panic recovery
	// middleware, but now wrapped with ratelimit middleware
	// use authenticate() middleware on all requests
	return app.recoverPanic(app.rateLimit(app.authenticate(router)))
}
