PHONY:build-producer
build-producer:
	go build -o ./bin/producer ./cmd/producer/...

PHONY:start-producer
start-producer:
	./bin/producer -file ./data/data_example.csv -parallel 2 -queue-addr amqp://rabbitmq:rabbitmq@0.0.0.0:5672/


PHONY:build-consumer
build-consumer:
	go build -o ./bin/consumer ./cmd/consumer/...

PHONY:start-consumer
start-consumer:
	./bin/consumer -queue-addr amqp://rabbitmq:rabbitmq@0.0.0.0:5672/