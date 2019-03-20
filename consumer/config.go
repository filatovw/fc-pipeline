package main

import (
	"flag"
	"os"
	"runtime"
	"strconv"

	"github.com/filatovw/fc-pipeline/libs/config"
	"github.com/pkg/errors"
)

// Config container with all required parameters
type Config struct {
	Parallel int
	Queue    config.Queue
	DB       config.DB
}

// LoadConfig create config and fill it with values from environment variables and command line
func LoadConfig() (*Config, error) {
	// read parameters
	config := &Config{
		Parallel: runtime.NumCPU(),
		Queue: config.Queue{
			Addr: "localhost:5672",
			User: "fcuser",
			Pass: "fcpass",
		},
		DB: config.DB{
			Host: "localhost",
			Port: 5432,
			User: "fcuser",
			Pass: "fcpass",
		},
	}
	if err := config.loadFromEnv(); err != nil {
		return nil, err
	}
	config.loadFromCLI()

	return config, nil
}

// loadFromEnv read parameters from environment
func (c *Config) loadFromEnv() error {
	if v := os.Getenv("FC_CONSUMER_QUEUE_ADDR"); v != "" {
		c.Queue.Addr = v
	}
	if v := os.Getenv("FC_CONSUMER_QUEUE_USER"); v != "" {
		c.Queue.User = v
	}
	if v := os.Getenv("FC_CONSUMER_QUEUE_PASS"); v != "" {
		c.Queue.Pass = v
	}

	if v := os.Getenv("FC_CONSUMER_DB_HOST"); v != "" {
		c.DB.Host = v
	}
	if v := os.Getenv("FC_CONSUMER_DB_PORT"); v != "" {
		port, err := strconv.Atoi(v)
		if err != nil {
			return errors.Wrapf(err, "loadFromEnv")
		}
		c.DB.Port = port
	}
	if v := os.Getenv("FC_CONSUMER_DB_USER"); v != "" {
		c.DB.User = v
	}
	if v := os.Getenv("FC_CONSUMER_DB_PASS"); v != "" {
		c.DB.Pass = v
	}
	return nil
}

// loadFromCLI read parameters from command line
func (c *Config) loadFromCLI() {
	var (
		parallel  int
		queueAddr string
		queueUser string
		queuePass string
		dbHost    string
		dbPort    int
		dbUser    string
		dbPass    string
	)
	flag.IntVar(&parallel, "parallel", 0, "number of workers")
	flag.StringVar(&queueAddr, "queue-addr", "", "env: FC_CONSUMER_QUEUE_ADDR. Address of queue [default: 0.0.0.0:5672]")
	flag.StringVar(&queueUser, "queue-user", "", "env: FC_CONSUMER_QUEUE_USER. Queue user [default: fcuser]")
	flag.StringVar(&queuePass, "queue-pass", "", "env: FC_CONSUMER_QUEUE_PASS. Queue pass [default: fcpass]")
	flag.StringVar(&dbHost, "db-host", "", "env: FC_CONSUMER_DB_HOST. Database host [default: 0.0.0.0]")
	flag.IntVar(&dbPort, "db-port", 0, "env: FC_CONSUMER_DB_PORT. Database port [default: 5432]")
	flag.StringVar(&dbUser, "db-user", "", "env: FC_CONSUMER_DB_USER. Database user [default: fcuser]")
	flag.StringVar(&dbPass, "db-pass", "", "env: FC_CONSUMER_DB_PASS. Database pass [default: fcpass]")
	flag.Parse()
	if parallel > 0 {
		c.Parallel = parallel
	}
	if queueAddr != "" {
		c.Queue.Addr = queueAddr
	}
	if queueUser != "" {
		c.Queue.User = queueUser
	}
	if queuePass != "" {
		c.Queue.Pass = queuePass
	}
	if dbHost != "" {
		c.DB.Host = dbHost
	}
	if dbPort > 0 {
		c.DB.Port = dbPort
	}
	if dbUser != "" {
		c.DB.User = dbUser
	}
	if dbPass != "" {
		c.DB.Pass = dbPass
	}
}
