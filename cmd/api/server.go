package main

import (
	"context"
	"errors"
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

	// create context for coordinating shutdow between servers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// start profiing server with shared context
	_ = app.startProfilingServer(ctx)

	// create shutdownerror channel
	shutdownError := make(chan error)

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

		// cancel context to signal profiling server to shutdown
		cancel()

		// create a context with a 5-sec timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// call shutdown() on server
		// but sends on the shutdownchannel
		// if it returns success or error
		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		// log message to say waiting for any bg gorouts
		// to finish tasks
		app.logger.PrintInfo("completing background tasks", map[string]string{
			"addr": srv.Addr,
		})

		// call wait() to block until waitgroup() counter is 0
		// then return nil when it is done with no issues
		app.wg.Wait()
		shutdownError <- nil
	}()

	// starting server msg
	app.logger.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env":  app.config.env,
	})

	// calling shutdown() on server will cause listenandserve()
	// to immediately return a http.errserverclosed
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// otherwise we wait to receive the return value from shutdown()
	// on the shutdownerror channel
	// if return value is an error, we know there was a problem
	// with the graceful shutdown and we return the error
	err = <-shutdownError
	if err != nil {
		return err
	}

	// at this point we know that the graceful shutdown completed successfully
	// and log a 'stopped server' message
	app.logger.PrintInfo("stopped server", map[string]string{
		"addr": srv.Addr,
	})

	return nil
}
