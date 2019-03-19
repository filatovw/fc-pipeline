PHONY:build-producer
build-producer:
	go build -o ./bin/producer ./cmd/producer/...

PHONY:start-producer
start-producer:
	./bin/producer -file ./data/data_10000.csv -parallel 10 -queue-addr 0.0.0.0:5672 -queue-user=fcuser -queue-pass=fcpass


PHONY:build-consumer
build-consumer:
	go build -o ./bin/consumer ./cmd/consumer/...

PHONY:start-consumer
start-consumer:
	./bin/consumer