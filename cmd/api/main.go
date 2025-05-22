package main

import (
	"context"
	"database/sql"
	"flag"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/meistens/api_practice/internal/data"
	"github.com/meistens/api_practice/internal/jsonlog"
	"github.com/meistens/api_practice/internal/mailer"
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
	// add a db struct field to hold the config settings for the db conn. pool
	// add other fields for the conn. pool
	//
	db struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	// add a new limiter struct containing fields for requests-per-secs
	// and burst values, and a bool field which can use to
	// enable/disable rate limiting
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
}

// define app struct to hold deps for the HTTP handlers,
// helpers, and middleware
type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
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
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostreSQL DSN")
	// read conn. pool settings from cmd flags into the config struct
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgesSQL max connection idle time")

	// Create command line flags to read the setting values into the config struct.
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	// Read the SMTP server configuration settings into the config struct, using the
	// Mailtrap settings as the default values.
	flag.StringVar(&cfg.smtp.host, "smtp-host", "smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 2525, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "c9e57d048da0c1", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "4942bda56789a2", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Greenlight <no-reply@testmail.net>", "SMTP sender")

	flag.Parse()

	// init. a new logger which writes to stdout
	// prefixed with current date and time
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	// call opendb() helper function to create conn. pool, passing the
	// config struct
	// if it returns an error, log it and exit
	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	// defer a call to db.close() so that conn. pool is closed before
	// main() exits
	defer db.Close()
	logger.PrintInfo("db conn. pool established", nil)

	// declare an instance of the app struct
	// containing the config struct, logger, models
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	// call app.serve() to start server
	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}
}

// opendb()
func openDB(cfg config) (*sql.DB, error) {
	// use sql.open() to create an empty conn. pool using dsn from the config struct
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	// set max. number of open (in-use+idle) connections in the pool
	db.SetMaxOpenConns(cfg.db.maxOpenConns)

	// set max. number of idle conns. in the pool
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	// use time.parseduration() to convert the idle timeout duration string to a time.duration type
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	//set max. idle timeout
	db.SetConnMaxIdleTime(duration)

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
