package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/streadway/amqp"

	"github.com/filatovw/fc-pipeline/libs/queue"
)

func main() {
	config, err := LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	stdlog := log.New(os.Stdout, "consumer", log.Lmicroseconds|log.LstdFlags|log.Lshortfile)
	stdlog.Printf("started")
	defer func() {
		stdlog.Printf("stopped")
	}()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	// catch SIGINT, SYGTERM
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		s := <-sigs
		stdlog.Printf("stopped with signal: %s", s)
		cancel()
	}()

	// connect to storage
	storage, err := newStorage(config.DB)
	if err != nil {
		stdlog.Print(err)
		return
	}
	defer storage.Close()

	// connect to queue
	qch, q, err := queue.Connect(config.Queue)
	if err != nil {
		stdlog.Print(err)
		return
	}
	defer qch.Close()

	// get consumer channel
	msgs, err := qch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		stdlog.Printf("failed to bind to queue: %s", err)
		return
	}
	wg := &sync.WaitGroup{}
	// start goroutines pool
	for i := config.Parallel; i > 0; i-- {
		wg.Add(1)
		go worker(ctx, i, wg, stdlog, storage, msgs)
	}
	wg.Wait()
}

// worker read message from queue and save it to database. If queue is not reachable - stop worker.
func worker(ctx context.Context, id int, wg *sync.WaitGroup, stdlog *log.Logger, storage *storage, msgs <-chan amqp.Delivery) {
	defer func() {
		wg.Done()
	}()
	var msg queue.Message
	for {
		select {
		case <-ctx.Done():
			return
		case m, ok := <-msgs:
			if !ok {
				return
			}
			if err := json.Unmarshal(m.Body, &msg); err != nil {
				stdlog.Printf("message decode: %s", err)
				if err := m.Reject(false); err != nil {
					stdlog.Printf("message reject: %s", err)
					return
				}
				continue
			}
			if err := storage.Insert(ctx, msg); err != nil {
				stdlog.Printf("insert item: %s", err)
				if err := m.Reject(false); err != nil {
					stdlog.Printf("message reject: %s", err)
					return
				}
				continue
			}
			if err := m.Ack(false); err != nil {
				stdlog.Printf("message ack: %s", err)
				return
			}
			stdlog.Printf("message processed: %#v", msg)
		}
	}
}
