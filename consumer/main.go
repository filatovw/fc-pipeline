package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/filatovw/fc-pipeline/libs/config"
	"github.com/filatovw/fc-pipeline/libs/queue"
	_ "github.com/lib/pq"
)

type Config struct {
	Parallel int
	Queue    config.Queue
	DB       config.DB
}

func main() {
	// read parameters
	config := Config{}
	flag.IntVar(&config.Parallel, "parallel", runtime.NumCPU()*2, "number of workers")

	flag.StringVar(&config.Queue.Addr, "queue-addr", "0.0.0.0:5672", "address of queue (Default: 0.0.0.0:5672)")
	flag.StringVar(&config.Queue.User, "queue-user", "fcuser", "queue user (Default: fcuser)")
	flag.StringVar(&config.Queue.Pass, "queue-pass", "fcpass", "queue pass (Default: fcpass)")

	flag.StringVar(&config.DB.Host, "db-host", "0.0.0.0", "address of queue (Default: 0.0.0.0)")
	flag.IntVar(&config.DB.Port, "db-port", 5432, "address of queue (Default: 5432)")
	flag.StringVar(&config.DB.User, "db-user", "fcuser", "queue user (Default: fcuser)")
	flag.StringVar(&config.DB.Pass, "db-pass", "fcpass", "queue pass (Default: fcpass)")
	flag.Parse()

	logger := log.New(os.Stdout, "consumer", log.Lmicroseconds|log.LstdFlags|log.Lshortfile)

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

	storage, err := newStorage(config.DB)
	if err != nil {
		log.Print(err)
		return
	}
	defer storage.Close()

	qch, q, err := queue.Connect(config.Queue)
	if err != nil {
		log.Print(err)
		return
	}
	defer qch.Close()

	msgs, err := qch.Consume(q.Name, "", false, false, false, false, nil)
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
			if err := storage.Insert(ctx, msg); err != nil {
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
