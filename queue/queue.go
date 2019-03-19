package queue

import (
	"github.com/streadway/amqp"
)

type Message struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func Declare(ch *amqp.Channel, name string) (amqp.Queue, error) {
	return ch.QueueDeclare(
		name,  // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
}
