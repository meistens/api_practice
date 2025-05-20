package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {
	// declare http server using same settings as main() func
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// start a background goroutine
	go func() {
		// create a quit channel which carries os.signal values
		quit := make(chan os.Signal, 1)

		// use signal.notify() to listen for incoming SIGINT
		// and SIGTERM signals, relay them to quit channel
		// other signals will retain their default behaviour
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		// read the signal from the quit channel
		// this code will block until a signal is received
		s := <-quit

		// log a msg to say that the signal has been caught
		app.logger.PrintInfo("caught signal", map[string]string{
			"signal": s.String(),
		})

		// exit app. with a 0 status code
		os.Exit(0)
	}()

	// starting server msg
	app.logger.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env":  app.config.env,
	})
	// start server
	return srv.ListenAndServe()
}
