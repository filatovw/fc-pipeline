package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/filatovw/fc-pipeline/queue"

	"github.com/streadway/amqp"
)

type Config struct {
	Parallel int
	Queue    Queue
	DB       DB
}

type DB struct {
	Host string
	User string
	Pass string
}

type Queue struct {
	Host string
	User string
	Pass string
}

func (q *Queue) Addr() string {
	return fmt.Sprintf("amqp://%s:%s@%s/", q.User, q.Pass, q.Host)
}

func main() {
	// read parameters
	config := Config{}

	flag.IntVar(&config.Parallel, "parallel", runtime.NumCPU()*2, "number of workers")

	flag.StringVar(&config.Queue.Host, "queue-host", "0.0.0.0:5672", "address of queue (Default: 0.0.0.0:5672).")
	flag.StringVar(&config.Queue.User, "queue-user", "fc-rabbitmq-user", "queue user (Default: fc-rabbitmq-user)")
	flag.StringVar(&config.Queue.Pass, "queue-pass", "fc-rabbitmq-pass", "queue pass (Default: fc-rabbitmq-pass)")

	flag.StringVar(&config.Queue.Host, "db-host", "0.0.0.0:5432", "address of queue (Default: 0.0.0.0:5672).")
	flag.StringVar(&config.Queue.User, "db-user", "fc-postgres-user", "queue user (Default: fc-postgres-user)")
	flag.StringVar(&config.Queue.Pass, "db-pass", "fc-postgres-pass", "queue pass (Default: fc-postgres-pass)")
	flag.Parse()

	logger := log.New(os.Stdout, "consumer", log.Lmicroseconds|log.LstdFlags|log.Llongfile)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	// catch SIGINT, SYGTERM
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		s := <-sigs
		log.Printf("stopped with signal: %s", s)
		cancel()
	}()

	logger.Printf("wait for messages")

	connection, err := amqp.Dial(config.Queue.Addr())
	if err != nil {
		logger.Fatalf("failed to establish connection to Queue service: %s", err)
	}
	defer connection.Close()
	ch, err := connection.Channel()
	if err != nil {
		logger.Printf("failed to create channel: %s", err)
		return
	}

	q, err := queue.Declare(ch, "csv2db")
	if err != nil {
		logger.Printf("failed to declare queue: %s", err)
		return
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		logger.Printf("failed to bind to queue: %s", err)
	}
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-msgs:
			if !ok {
				return
			}

			log.Printf("%s", msg.Body)
			if err := msg.Ack(false); err != nil {
				log.Printf("message ack: %s", err)
			}
		}

	}
}
