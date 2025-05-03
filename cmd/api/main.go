package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// declare string containing semver
// will be automatically generated later in book, but for now
// hardcode
const version = "1.0.0"

// define a config struct to hold all config settings for app
// for now, these inside will do (VC if more is added to see)
// env (dev, stage, prod), port self-explanatory
// cmd_line *FLAGS* will be used when app starts
type config struct {
	port int
	env  string
	// add a db struct field to hold the config settings for the db connection
	// pool
	db struct {
		dsn string
	}
}

// define app struct to hold deps for the HTTP handlers,
// helpers, and middleware
// for now, it contains the copy of the config struct and
// a logger
type application struct {
	config config
	logger *log.Logger
}

func main() {
	//declares an instance of the config struct
	var cfg config

	// read the value of the port and env cmd-flags into
	// the config struct.
	// default to using the port number 4000 and env "dev"
	// if no other flags are provided
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (dev|stage|prod)")
	// read dsn value from the db-dsn flag into the config struct
	// default flag dev is used if none is provided
	flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://greenlight:greenlight12345@localhost/greenlight?sslmode=disable", "PostreSQL DSN")

	flag.Parse()

	// init. a new logger which writes to stdout
	// prefixed with current date and time
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// call opendb() helper function to create conn. pool, passing the
	// config struct
	// if it returns an error, log it and exit
	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal(err)
	}

	// defer a call to db.close() so that conn. pool is closed before
	// main() exits
	defer db.Close()

	logger.Printf("conn. pool established")

	// declare an instance of the app struct
	// containing the config struct and logger
	app := &application{
		config: cfg,
		logger: logger,
	}

	// declare a new servemux and add a /v1/healthcheck route
	// which dispatches requests to the healthcheckhanddler
	// method
	//mux := http.NewServeMux()
	//mux.HandleFunc("/v1/healthcheck", app.healthcheckHandler)

	// declare http server with some sensible timeout setting
	// wich listens on the port provided in the config struct
	// and uses the servemux created above as the handler
	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", cfg.port),
		// use the httprouter instance returned by app.routes() as the server handler
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// start server
	// sheesh, way too much stuff to do using a built-in pkg
	logger.Printf("starting %s server on %s", cfg.env, srv.Addr)
	err = srv.ListenAndServe()
	logger.Fatal(err)
}

// opendb()
func openDB(cfg config) (*sql.DB, error) {
	// use sql.open() to create an empty conn. pool using dsn from the config struct
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	// create context with a 5-sec timeout deadline
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// establish new conn. to db, pass context created as param
	// if conn. isn't established within 5sec., return error
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}
