package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/meistens/api_practice/internal/data"
	"github.com/meistens/api_practice/internal/validator"
)

func (app *application) createAuthTokenHandler(w http.ResponseWriter, r *http.Request) {
	// parse mail and password from request body
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// validate email and password provided by the client
	v := validator.New()

	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// lookup user record based on mail address
	// if no matches, invalid credentials and send a 401
	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// check if the provided password matches the actual password for the user
	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// if password doesn't match, invalid creds
	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}
	// generate new token if it matches
	token, err := app.models.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	//encode token in JSON and send in the response along with 201
	err = app.writeJSON(w, http.StatusCreated, envelope{"authentication_token": token}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
