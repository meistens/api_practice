package main

import (
	"fmt"
	"net/http"

	"golang.org/x/time/rate"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// create a deferred func. (which will always be run in the event of a panic)
		defer func() {
			// use builtin recover func. to check if there has been a panic
			if err := recover(); err != nil {
				// if there was a panic, set a "conn: close" header on
				// the response. This acts as a trigger to make Go's http
				// router automatically close the current conn after
				// a response has been sent
				w.Header().Set("Connection", "Close")
				// value returned by recover() has the type interface{}
				// so we use fmt.Errorf() to normalize it into an error
				// and call serverErrorResponse() helper
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler {
	// init. a new rate limiter which allows an average of 2reqs/s, max. 4reqs
	// in a single burst
	limiter := rate.NewLimiter(2, 4)

	// function we are returning is a closure, which closes over the limiter
	// variable
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// call limiter.ALlow() to see if the request is permitted
		// if not, 429 called using rateLimitExceededResponse()
		if !limiter.Allow() {
			app.rateLimitExceededResponse(w, r)
		}
		next.ServeHTTP(w, r)
	})
}
