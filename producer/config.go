package main

import (
	"flag"
	"os"
	"runtime"

	"github.com/filatovw/fc-pipeline/libs/config"
)

// Config container with all required parameters
type Config struct {
	Parallel int
	File     string
	Queue    config.Queue
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
	}
	if err := config.loadFromEnv(); err != nil {
		return nil, err
	}
	config.loadFromCLI()

	return config, nil
}

// loadFromEnv read parameters from environment
func (c *Config) loadFromEnv() error {
	if v := os.Getenv("FC_PRODUCER_QUEUE_ADDR"); v != "" {
		c.Queue.Addr = v
	}
	if v := os.Getenv("FC_PRODUCER_QUEUE_USER"); v != "" {
		c.Queue.User = v
	}
	if v := os.Getenv("FC_PRODUCER_QUEUE_PASS"); v != "" {
		c.Queue.Pass = v
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
		file      string
	)
	flag.IntVar(&parallel, "parallel", 0, "number of workers")
	flag.StringVar(&queueAddr, "queue-addr", "", "env: FC_PRODUCER_QUEUE_ADDR. Address of queue [default: 0.0.0.0:5672]")
	flag.StringVar(&queueUser, "queue-user", "", "env: FC_PRODUCER_QUEUE_USER. Queue user [default: fcuser]")
	flag.StringVar(&queuePass, "queue-pass", "", "env: FC_PRODUCER_QUEUE_PASS. Queue pass [default: fcpass]")
	flag.StringVar(&file, "file", "", "path to CSV file")
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
	if file != "" {
		c.File = file
	}
}
