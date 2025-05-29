package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/meistens/api_practice/internal/data"
	"github.com/meistens/api_practice/internal/validator"
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
	// define client struct to hold the rate limiter and last seen time
	// for each client
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	// declare mutex and a map to hold the clients IP addr and rate
	// limiters
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// launch a background goroutine which removes old entries from the clients map
	// once every minute
	go func() {
		for {
			time.Sleep(time.Minute)

			// lock mutex to prevent any rate limiter checks from happening
			// while the cleanup is taking place
			mu.Lock()

			// loop through all clients
			// if they haven't been seen within the last 3 minutes,
			// delete corresponding entry from the map
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			// unlock mutex
			mu.Unlock()
		}
	}()

	// function we are returning is a closure, which closes over the limiter
	// variable
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check if rate-limit is enabled
		if app.config.limiter.enabled {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			// lock mutex to prevent this code from being executed concurrently
			mu.Lock()

			// check to see if the IP addr already exists in the map
			// if it doesn't, init. a new rate limiter, add the IP addr and
			// limiter to the map
			if _, found := clients[ip]; !found {
				clients[ip] = &client{
					limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst),
				}
			}

			clients[ip].lastSeen = time.Now()

			// call Allow() method on rate limiter for the current IP addr
			// if request isn't allowed, unlock mutex and send a 429
			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(w, r)
				return
			}

			// unlock mutex before calling the next handler in the chain
			mu.Unlock()
		}
		next.ServeHTTP(w, r)
	})
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// add "Vary: Auth." header response
		w.Header().Add("Vary", "Authorization")

		// retrieve the value of the Auth. header from the request
		// This will return the empty string "" if no header is found
		authorizationHeader := r.Header.Get("Authorization")

		// if there is no auth. header found, use contextsetuser() helper
		// to add anonuser to the request context
		if authorizationHeader == "" {
			r = app.contextSetUser(r, data.AnonUser)
			next.ServeHTTP(w, r)
			return
		}

		// otherwise, we expect the value of the Auth. heade to be in the format
		// "Bearer <>"
		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidCredentialsResponse(w, r)
			return
		}

		// extract the actual token from header
		token := headerParts[1]

		// validate to see if it is the right format
		v := validator.New()

		// if token isn't valid, call the invalid() helper to send
		// the appropriate response rather than failed()
		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
			app.invalidAuthTokenResponse(w, r)
			return
		}

		// retrieve deetz of user associated with the auth. token
		// doing the same as above for invalid token
		// retrieve the deetz of the user associated with the auth. token
		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}
		// call contextsetuser() helper to add user info to the request ctx
		r = app.contextSetUser(r, user)

		// call next handler in the chain
		next.ServeHTTP(w, r)
	})
}
