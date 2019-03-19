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

	"github.com/filatovw/fc-pipeline/queue"

	"github.com/streadway/amqp"
)

type Config struct {
	Parallel int
	File     string
	Queue    Queue
}

type Queue struct {
	Addr string
	User string
	Pass string
}

var emailPattern = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

type Publisher interface {
	Publish(string, []byte) error
}

func main() {
	// read parameters
	config := Config{Queue: Queue{}}
	flag.IntVar(&config.Parallel, "parallel", runtime.NumCPU()*2, "number of workers")
	flag.StringVar(&config.File, "file", "", "path to CSV file")
	flag.StringVar(&config.Queue.Addr, "queue-addr", "queue", "address of queue")
	flag.StringVar(&config.Queue.User, "queue-user", "", "queue user")
	flag.StringVar(&config.Queue.Pass, "queue-pass", "", "queue password")
	flag.Parse()

	logger := log.New(os.Stdout, "producer", log.Lmicroseconds|log.LstdFlags|log.Llongfile)

	connection, err := amqp.Dial(config.Queue.Addr)
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

	// open csv file
	f, err := os.Open(config.File)
	if err != nil {
		logger.Fatalf("failed to open file: %s", config.File)
	}
	defer f.Close()

	records := make(chan []string)

	// csv reader
	r := csv.NewReader(f)

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
		go worker(ctx, i, logger, ch, q, wg, records)
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

func worker(ctx context.Context, id int, log *log.Logger, channel *amqp.Channel, q amqp.Queue, wg *sync.WaitGroup, input <-chan []string) {
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

			msg, err := createMessage(value)
			if err != nil {
				log.Print(err)
				continue
			}
			body, err := json.Marshal(*msg)
			if err != nil {
				log.Printf("error: json encode: %s", err)
			}
			if err := channel.Publish("", q.Name, false, false, amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				ContentType:  "application/json",
				Body:         []byte(body),
			}); err != nil {
				log.Printf("error: publish to queue: %s", err)
				continue
			}
			log.Printf("w: %d, msg: %v", id, msg)
		}
	}
}

func createMessage(value []string) (*queue.Message, error) {
	if len(value) != 2 {
		return nil, fmt.Errorf("error: unexpected value %v", value)
	}
	if value[0] == "" {
		return nil, fmt.Errorf("error: name can not be empty")
	}
	if !emailPattern.Match([]byte(value[1])) {
		return nil, fmt.Errorf("error: invalid email %s", value[1])
	}
	return &queue.Message{
		Name:  value[0],
		Email: value[1],
	}, nil
}
