package main

import (
	"errors"
	"net/http"

	"github.com/meistens/api_practice/internal/data"
	"github.com/meistens/api_practice/internal/validator"
)

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	// create anonymous struct to hold expected data from request body
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// parse the request body into the anonymous struct
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// copy data from request body into a new user struct
	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	// use passord.set() method to generate and store the hashed and plaintext
	// passwords
	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	// validate user struct and return the error msg to the client
	// if any of the checks fail
	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// insert the user data into the db
	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		// something causing it to return server error instead of custom msg
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Call the Send() method on our Mailer, passing in the user's email address,
	// name of the template file, and the User struct containing the new user's data.
	// launch a goroutine which runs an anon function that sends the welcome mail
	app.background(func() {
		err = app.mailer.Send(user.Email, "user_welcome.tmpl", user)
		if err != nil {
			// if there is an error sending, use app.logger.PrintError()
			// to manage it instead of app.servererrorresponse() helper
			app.logger.PrintError(err, nil)
		}
	})

	// write json response containing the user data along with 202 code
	err = app.writeJSON(w, http.StatusAccepted, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
