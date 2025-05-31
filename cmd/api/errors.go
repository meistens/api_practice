package main

import (
	"fmt"
	"net/http"
)

func (app *application) logError(r *http.Request, err error) {
	// FINALLY, some use!!!!!!!!!!!!!
	app.logger.PrintError(err, map[string]string{
		"request_method": r.Method,
		"request_url":    r.URL.String(),
	})
}

// func. sends JSON-formatted msgs to the client with a given status code from other functions
func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	env := envelope{"error": message}

	// helper logs err response, returns a 500 after
	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(500)
	}
}

// used when the app encounters an unexpected problem at runtime
func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)

	message := "server problem"
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}

// used for 404
func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "404 not found..."
	app.errorResponse(w, r, http.StatusNotFound, message)
}

// used for 405
func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

// used for 422
func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

// use for 409
func (app *application) editConflictResponse(w http.ResponseWriter, r *http.Request) {
	message := "unable to update record due to an edit conflict, do try again in a few seconds"
	app.errorResponse(w, r, http.StatusConflict, message)
}

// use for 429|rate limit exceeds
func (app *application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	message := "rate limit exceeded"
	app.errorResponse(w, r, http.StatusTooManyRequests, message)
}

// use for 401
func (app *application) invalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	message := "invalid auth. credentials"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

// invalid auth. token, uses 401 BUT for this instance
func (app *application) invalidAuthTokenResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")

	message := "invalid or missing authentication token"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

// 401, for auth
func (app *application) authRequiredReponse(w http.ResponseWriter, r *http.Request) {
	message := "you must be authenticated to access this resource"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

// 403, for inactive account
func (app *application) inactiveAccResponse(w http.ResponseWriter, r *http.Request) {
	message := "your user account must be activated to access this reource"
	app.errorResponse(w, r, http.StatusForbidden, message)
}

func (app *application) notPermittedResponse(w http.ResponseWriter, r *http.Request) {
	message := "your user account doesn't have the necessary permissions to access this resource"
	app.errorResponse(w, r, http.StatusForbidden, message)
}
