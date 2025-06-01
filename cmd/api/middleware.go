package main

import (
	"errors"
	"expvar"
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

func (app *application) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	// rather than return, assign it to a variable
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// use contextgetuser() to retrieve user info from the req. context
		user := app.contextGetUser(r)

		// if the user is not activated, call inactiveaccountres()
		if !user.Activated {
			app.inactiveAccResponse(w, r)
			return
		}

		// call next handler in the chain
		next.ServeHTTP(w, r)
	})
	// authenticate and activate, it shuld be outside the func. i know...
	return app.requireAuthUser(fn)
}

func (app *application) requireAuthUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		if user.IsAnon() {
			app.authRequiredReponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) requirePermission(code string, next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// retrieve user from the request context
		user := app.contextGetUser(r)

		// get the slice of permissions for the user
		permissions, err := app.models.Permissions.GetAllUserPerms(user.ID)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		// check if the slice includes the required requirePermission
		// if not, return a 403
		if !permissions.Include(code) {
			app.notPermittedResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	}
	// wrap around the requireactivateduser() before returning
	return app.requireActivatedUser(fn)
}

// CORS enabler (for browser compat.)
func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		// Add the "Vary: Access-Control-Request-Method" header.
		w.Header().Add("Vary", "Access-Control-Request-Method")
		origin := r.Header.Get("Origin")
		if origin != "" {
			for i := range app.config.cors.trustedOrigins {
				if origin == app.config.cors.trustedOrigins[i] {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					// Check if the request has the HTTP method OPTIONS and contains the
					// "Access-Control-Request-Method" header. If it does, then we treat
					// it as a preflight request.
					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
						// Set the necessary preflight response headers, as discussed
						// previously.
						w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
						w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
						// Write the headers along with a 200 OK status and return from
						// the middleware with no further action.
						w.WriteHeader(http.StatusOK)
						return
					}
					break
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

// metrics
func (app *application) metrics(next http.Handler) http.Handler {
	// initalize the new expvar variables when middleware chain is first built
	totalReqReceived := expvar.NewInt("total_requests_received")
	totalResSent := expvar.NewInt("total_responses_sent")
	totalProcTimeMicrosecs := expvar.NewInt("total_processing_time_µs")

	// run for every request...
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// record the time the process started the request
		start := time.Now()

		// use the add() method to increment the number of responses
		totalReqReceived.Add(1)

		// call next handler in the chain
		next.ServeHTTP(w, r)

		// on the way back up the middleware chain, increment responses
		// by 1
		totalResSent.Add(1)

		// calc. microseconds since the request process
		// increment the total processing time by 1
		duration := time.Since(start).Microseconds()
		totalProcTimeMicrosecs.Add(duration)
	})
}
