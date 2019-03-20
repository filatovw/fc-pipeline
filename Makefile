export GO111MODULE=on
DC = docker-compose

PHONY:prune
prune:
	docker system prune -f --volumes

PHONY:clean
clean:
	@rm -rf ./bin/*

PHONY:migrate-db
migrate-db:
	$(DC) up migrate-db

PHONY:drop-db
drop-db:
	$(DC) up drop-db

PHONY:infra
infra:
	$(DC) stop queue db 
	$(DC) rm -f queue db 
	$(DC) up -d queue db 

PHONY:start
start:
	$(DC) stop producer consumer
	$(DC) rm -f producer consumer
	$(DC) up --scale consumer=2 -d producer consumer

PHONY:logs
logs:
	$(DC) logs -f producer consumer

PHONY:stop
stop:
	$(DC) stop

PHONY:build
build:
	$(DC) build producer consumer

PHONY:push
push:
	$(DC) push producer consumer

PHONY:test
test:
	go test -v -race ./...

PHONY:check
check:
	golangci-lint run ./...

PHONY:build-producer
build-producer:
	go build -o ./bin/producer ./producer/...

PHONY:build-consumer
build-consumer:
	go build -o ./bin/consumer ./consumer/...