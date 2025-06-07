package main

import (
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	// init new router instance
	router := httprouter.New()

	// convert the notFoundResponse() helper to a http.Handler using HandlerFunc() adapter
	// then set as a custom err handler for 404
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	// register relevant methods, patterns and handler funcs for
	// the endpoints using handlerfunc()
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	// Use the requirePermission() middleware on each of the /v1/movies** endpoints,
	// passing in the required permission code as the first parameter.
	router.HandlerFunc(http.MethodGet, "/v1/movies", app.requirePermission("movies:read", app.listMoviesHandler))
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.requirePermission("movies:write", app.createMovieHandler))
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.requirePermission("movies:read", app.showMovieHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.requirePermission("movies:write", app.updateMovieHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.requirePermission("movies:write", app.deleteMovieHandler))

	// updated
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)
	// users
	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)

	// authentication
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthTokenHandler)

	// PUT /v1/users/password endpoint
	router.HandlerFunc(http.MethodPut, "/v1/users/password", app.updateUserPassHandler)

	// password reset endpoint
	router.HandlerFunc(http.MethodPost, "/v1/tokens/password-reset", app.createPassResetHandler)

	router.HandlerFunc(http.MethodPost, "/v1/tokens/activation", app.createActivationTokenHandler)

	// debug
	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())

	// return the httprouter instance, but with the panic recovery
	// middleware, but now wrapped with ratelimit middleware
	// use authenticate() middleware on all requests
	// now with CORS, POSITIONING IS IMPORTANT!!!!
	return app.metrics(app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router)))))
}
