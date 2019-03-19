package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/filatovw/fc-pipeline/queue"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/streadway/amqp"
)

type Config struct {
	Parallel int
	Queue    Queue
	DB       DB
}

type DB struct {
	Host string
	Port string
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
	config := Config{DB: DB{}, Queue: Queue{}}

	flag.IntVar(&config.Parallel, "parallel", runtime.NumCPU()*2, "number of workers")

	flag.StringVar(&config.Queue.Host, "queue-host", "0.0.0.0:5672", "address of queue (Default: 0.0.0.0:5672).")
	flag.StringVar(&config.Queue.User, "queue-user", "fcuser", "queue user (Default: fcuser)")
	flag.StringVar(&config.Queue.Pass, "queue-pass", "fcpass", "queue pass (Default: fcpass)")

	flag.StringVar(&config.DB.Host, "db-host", "0.0.0.0", "address of queue (Default: 0.0.0.0)")
	flag.StringVar(&config.DB.Port, "db-port", "5432", "address of queue (Default: 5432)")
	flag.StringVar(&config.DB.User, "db-user", "fcuser", "queue user (Default: fcuser)")
	flag.StringVar(&config.DB.Pass, "db-pass", "fcpass", "queue pass (Default: fcpass)")
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

	log.Printf("%#v", config)

	connstring := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.DB.Host, config.DB.Port, config.DB.User, config.DB.Pass, "userdata")
	log.Printf("%s", connstring)
	db, err := sqlx.Connect("postgres", connstring)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

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
		case m, ok := <-msgs:
			if !ok {
				return
			}
			msg := queue.Message{}
			if err := json.Unmarshal(m.Body, &msg); err != nil {
				log.Printf("message decode: %s", err)
				continue
			}
			if _, err := db.ExecContext(ctx, db.Rebind(`INSERT INTO contacts (name, email) VALUES (?, ?);`), msg.Name, msg.Email); err != nil {
				log.Printf("insert item %v, error: %s", msg, err)
				if err := m.Reject(false); err != nil {
					log.Printf("message reject: %s", err)
				}
				continue
			}
			if err := m.Ack(false); err != nil {
				log.Printf("message ack: %s", err)
			}
		}
	}
}
