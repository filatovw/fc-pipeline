package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/filatovw/fc-pipeline/libs/queue"

	"github.com/streadway/amqp"
)

func main() {
	config, err := LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	stdlog := log.New(os.Stdout, "producer", log.Lmicroseconds|log.LstdFlags|log.Lshortfile)
	stdlog.Printf("started")
	defer func() {
		stdlog.Printf("stopped")
	}()

	qch, q, err := queue.Connect(config.Queue)
	if err != nil {
		stdlog.Print(err)
		return
	}
	defer qch.Close()

	// open csv file
	f, err := os.Open(config.File)
	if err != nil {
		stdlog.Fatalf("failed to open file: %s", config.File)
	}
	defer f.Close()
	// csv reader
	r := csv.NewReader(f)

	records := make(chan []string)
	wg := &sync.WaitGroup{}

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

	// create pool
	for i := config.Parallel; i > 0; i-- {
		wg.Add(1)
		go worker(ctx, i, wg, stdlog, qch, *q, records)
	}

	numRows := 0
	defer func() {
		stdlog.Printf("file: %s, rows processed: %d", config.File, numRows)
	}()
	// process file row by row
	for {
		record, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			stdlog.Printf("read from file: %s", err)
			return
		}
		records <- record
		numRows++
	}
	close(records)
	wg.Wait()
}

// worker validate input record and send it to queue. If queue is not reachable - stop worker.
func worker(ctx context.Context, id int, wg *sync.WaitGroup, stdlog *log.Logger, qch *amqp.Channel, q amqp.Queue, input <-chan []string) {
	defer func() {
		wg.Done()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case value, ok := <-input:
			if !ok {
				return
			}

			if err := validateRecord(value); err != nil {
				stdlog.Print(err)
				continue
			}
			msg := queue.Message{Name: value[0], Email: value[1]}
			body, err := json.Marshal(msg)
			if err != nil {
				stdlog.Printf("error: json encode: %s", err)
			}
			if err := qch.Publish("", q.Name, false, false, amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				ContentType:  "application/json",
				Body:         []byte(body),
			}); err != nil {
				stdlog.Printf("error: publish to queue: %s", err)
				return
			}
		}
	}
}
