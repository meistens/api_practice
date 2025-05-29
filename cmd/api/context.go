package main

import (
	"context"
	"net/http"

	"github.com/meistens/api_practice/internal/data"
)

// define a custom contextkey type with underlying type string
type contextKey string

// convert string user to a contextKey type and assign to usercontextkey
const userContextKey = contextKey("user")

// return a new copy of request with contextsetuser()
func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

// retrieve user srruct from the request context
func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}
	return user
}
