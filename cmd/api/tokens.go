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

// generate password reset token, to be sent to user mail address
func (app *application) createPassResetHandler(w http.ResponseWriter, r *http.Request) {
	// parse and validate user email address
	var input struct {
		Email string `json:"email"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	if data.ValidateEmail(v, input.Email); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// retieve corresponding user record for the email address
	// if not found, return an error message to the client
	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("email", "no matching email address found")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// return an error msg if user is not activated
	if !user.Activated {
		v.AddError("email", "user account must be activated")
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// if not all above, create a new password reset token with a 45-min expiry time
	token, err := app.models.Tokens.New(user.ID, 45*time.Minute, data.ScoprPassReset)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// email user with their password reset token in the background
	app.background(func() {
		data := map[string]interface{}{
			"passwordResetToken": token.Plaintext,
		}

		// since mail address may be case sensitive, either correct user misuse
		// or select from the db since that was take care of at the db level
		// and we are going with selecting from db after getting thei input
		err = app.mailer.Send(user.Email, "token_password_reset.tmpl", data)
		if err != nil {
			app.logger.PrintError(err, nil)
		}
	})

	// send a 202 and confirmation msg to the user
	env := envelope{"message": "an email will be sent to you containing the password reset instructions"}

	err = app.writeJSON(w, http.StatusAccepted, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// standalone activation token handler
func (app *application) createActivationTokenHandler(w http.ResponseWriter, r *http.Request) {
	// parse and validate mail
	var input struct {
		Email string `json:"email"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	if data.ValidateEmail(v, input.Email); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// retrieve user record that matches mail, else throw error
	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("email", "no matching email address found")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// return error if user has already been activated
	if user.Activated {
		v.AddError("email", "user has already been activated")
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// otherwise create a new activation token
	token, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// email user with their additional activation token
	app.background(func() {
		data := map[string]interface{}{
			"activationToken": token.Plaintext,
		}

		// as usual, mail is case sensitive, so using the ones already stored
		// in db because that already handled at the db level
		err := app.mailer.Send(user.Email, "token_activation.tmpl", data)
		if err != nil {
			app.logger.PrintError(err, nil)
		}
	})

	// send a 202
	env := envelope{"message": "an email will be sent to you containing activation instructions"}

	err = app.writeJSON(w, http.StatusAccepted, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
