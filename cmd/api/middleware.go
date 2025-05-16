package main

import (
	"fmt"
	"net/http"
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
