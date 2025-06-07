package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"
)

// start a separate HTTP server for pprof endpoints
func (app *application) startProfilingServer(ctx context.Context) *http.Server {
	// if not enabled, return nil
	if !app.config.profiling.enabled {
		return nil
	}

	// set up on a different port (check for free ones)
	profilingMux := http.NewServeMux()

	// register pprof handlers on our custom mux
	profilingMux.HandleFunc("/debug/pprof/", http.DefaultServeMux.ServeHTTP)
	profilingMux.HandleFunc("/debug/pprof/cmdline", http.DefaultServeMux.ServeHTTP)
	profilingMux.HandleFunc("/debug/pprof/profile", http.DefaultServeMux.ServeHTTP)
	profilingMux.HandleFunc("/debug/pprof/symbol", http.DefaultServeMux.ServeHTTP)
	profilingMux.HandleFunc("/debug/pprof/trace", http.DefaultServeMux.ServeHTTP)

	// add expvar endpoint
	profilingMux.HandleFunc("/debug/vars", http.DefaultServeMux.ServeHTTP)

	// add health check endpoint for profiling server
	profilingMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("profiling server is healthy"))
	})

	profilingServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.profiling.port),
		Handler:      profilingMux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	app.logger.PrintInfo("starting profiling server", map[string]string{
		"addr": profilingServer.Addr,
	})

	// start server in goroutine
	app.wg.Add(1)
	go func() {
		defer app.wg.Done()

		if err := profilingServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.logger.PrintError(err, map[string]string{
				"component": "profiling_server",
			})
		}
	}()

	// graceful shutdown handler
	app.wg.Add(1)
	go func() {
		defer app.wg.Done()

		// wait for context cancellation (shutdown signal)
		<-ctx.Done()

		app.logger.PrintInfo("shutting down profiling server", map[string]string{
			"addr": profilingServer.Addr,
		})

		// create a context with timeout for shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := profilingServer.Shutdown(shutdownCtx); err != nil {
			app.logger.PrintError(err, map[string]string{
				"component": "profiling_server_shutdown",
			})
		} else {
			app.logger.PrintInfo("profiling server stopped", map[string]string{
				"addr": profilingServer.Addr,
			})
		}
	}()

	return profilingServer
}

// set runtime params for better performance
func (app *application) optimizeRuntime() {
	// set GOMAXPROCS to number of CPUs if not set
	if runtime.GOMAXPROCS(0) == 1 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}
}
