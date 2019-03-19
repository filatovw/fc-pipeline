package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"sync"
	"syscall"

	"github.com/filatovw/fc-pipeline/libs/config"
	"github.com/filatovw/fc-pipeline/libs/queue"

	"github.com/streadway/amqp"
)

var emailPattern = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

type Config struct {
	Parallel int
	File     string
	Queue    config.Queue
}

func main() {
	config := Config{}
	flag.IntVar(&config.Parallel, "parallel", runtime.NumCPU()*2, "number of workers")
	flag.StringVar(&config.File, "file", "", "path to CSV file")
	flag.StringVar(&config.Queue.Addr, "queue-addr", "0.0.0.0:5672", "address of queue (Default: 0.0.0.0:5672)")
	flag.StringVar(&config.Queue.User, "queue-user", "fcuser", "queue user (Default: fcuser)")
	flag.StringVar(&config.Queue.Pass, "queue-pass", "fcpass", "queue pass (Default: fcpass)")
	flag.Parse()

	logger := log.New(os.Stdout, "producer", log.Lmicroseconds|log.LstdFlags|log.Lshortfile)

	qch, q, err := queue.Connect(config.Queue)
	if err != nil {
		log.Print(err)
		return
	}
	defer qch.Close()

	// open csv file
	f, err := os.Open(config.File)
	if err != nil {
		logger.Fatalf("failed to open file: %s", config.File)
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
		logger.Printf("stopped with signal: %s", s)
		cancel()
	}()

	// create pool
	for i := config.Parallel; i > 0; i-- {
		wg.Add(1)
		go worker(ctx, i, logger, qch, *q, wg, records)
	}

	// read file row by row
	for {
		record, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			logger.Printf("read from file: %s", err)
			return
		}
		records <- record
	}
	close(records)
	wg.Wait()
}

func worker(ctx context.Context, id int, log *log.Logger, qch *amqp.Channel, q amqp.Queue, wg *sync.WaitGroup, input <-chan []string) {
	defer func() {
		log.Printf("worker %d stopped", id)
		wg.Done()
	}()

	log.Printf("worker %d started", id)
	for {
		select {
		case <-ctx.Done():
			return
		case value, ok := <-input:
			if !ok {
				return
			}

			if err := validateRecord(value); err != nil {
				log.Print(err)
				continue
			}
			msg := queue.Message{Name: value[0], Email: value[1]}
			body, err := json.Marshal(msg)
			if err != nil {
				log.Printf("error: json encode: %s", err)
			}
			if err := qch.Publish("", q.Name, false, false, amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				ContentType:  "application/json",
				Body:         []byte(body),
			}); err != nil {
				log.Printf("error: publish to queue: %s", err)
				continue
			}
		}
	}
}

func validateRecord(value []string) error {
	if len(value) != 2 {
		return fmt.Errorf("error: unexpected value %v", value)
	}
	if value[0] == "" {
		return fmt.Errorf("error: name can not be empty")
	}
	if !emailPattern.Match([]byte(value[1])) {
		return fmt.Errorf("error: invalid email %s", value[1])
	}
	return nil
}
