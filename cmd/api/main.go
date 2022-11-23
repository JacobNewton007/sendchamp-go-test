package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/JacobNewton007/sendchamp-go-test/internal/data"
	"github.com/JacobNewton007/sendchamp-go-test/internal/jsonlog"
	"github.com/JacobNewton007/sendchamp-go-test/internal/rabbitmq"
	_ "github.com/go-sql-driver/mysql"
	"github.com/rabbitmq/amqp091-go"
)

var (
	version   = "1.0.0"
	buildTime string
)

type config struct {
	port int
	env  string
	db   struct {
		dsn      string
		username string
		password string
		hostname string
		dbname   string

		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}

	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}

	rabbitmq struct {
		uri string
	}

	cors struct {
		trustedOrigins []string
	}
}

type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	rMq    rabbitmq.RabbitMQ
	wg     sync.WaitGroup
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "API server ports")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "MySQL DSN")

	// Read the connection pool settings from command-line flags into the config struct.
	// Notice the default values that we're using?
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "MySQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "MySQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "MySQL max connection idle time")

	// Create command line flags to read the setting values into the config struct.
	// Notice that we use true as the default for the 'enabled' setting?
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.StringVar(&cfg.rabbitmq.uri, "rabbitmq-uri", "", "RabbitMQ uri")

	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})

	displayVersion := flag.Bool("version", false, "Display version and exist")

	flag.Parse()

	if *displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		fmt.Printf("Build time:\t%s\n", buildTime)
		os.Exit(0)
	}
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	// Defer a call to db.Close() so that the connection pool is closed before the
	// main() function exits.
	defer db.Close()
	logger.PrintInfo("database connection pool established", nil)

	rabbitConn, err := OpenRabbitQueue(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	defer rabbitConn.Close()
	logger.PrintInfo("rabbitmq connection pool established", nil)
	// Use the data.NewModels() function to initialize a Models struct, passing in the
	// connection pool as a parameter.
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		rMq: ,
	}

	err = app.server()
	if err != nil {
		logger.PrintFatal(err, nil)
	}
}

func openDB(cfg config) (*sql.DB, error) {
	// Use sql.Open() to create an empty connection pool, using the DSN from the config
	// struct.
	db, err := sql.Open("mysql", cfg.db.dsn)
	if err != nil {
		return nil, err
	}
	// set the maximum number of open (in-use + idle) connections in the pool. Note that
	// passing a value less than or equal to 0 will mean there is no limit
	db.SetMaxOpenConns(cfg.db.maxOpenConns)

	// Set the number of idle connections in the pool. Again, passing a value
	// less than or equal to 0 will mean there is no limit.
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	// Use the time.ParseDuration() function to convert the idle timeout duration string
	// to a time.Duration type.
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	// Set the maximum idle timeout.
	db.SetConnMaxIdleTime(duration)

	// Create a context with a 5-second timeout deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	// Return the sql.DB connection pool.
	return db, nil
}

func OpenRabbitQueue(cfg config) (*amqp091.Connection, error) {
	conn, err := amqp091.Dial(cfg.rabbitmq.uri)
	if err != nil {
		panic(err)
	}
	return conn, nil
}

// func dsn(username, password, hostname, dbName string) string {
// 	return fmt.Sprintf("%s:%s@tcp(%s)/%s", username, password, hostname, dbName)
// }
