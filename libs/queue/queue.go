package queue

import (
	"github.com/filatovw/fc-pipeline/libs/config"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type Message struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func declare(ch *amqp.Channel, name string) (amqp.Queue, error) {
	return ch.QueueDeclare(
		name,  // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
}

func Connect(cfg config.Queue) (*amqp.Channel, *amqp.Queue, error) {
	qconn, err := amqp.Dial(cfg.ConnectionString())
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to establish connection to Queue service")
	}
	qch, err := qconn.Channel()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create channel")
	}

	q, err := declare(qch, "csv2db")
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to declare queue")
	}
	return qch, &q, nil
}
