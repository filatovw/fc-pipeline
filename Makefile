DC = docker-compose
PHONY:prune
prune:
	docker system prune -f --volumes

PHONY:migrate
migrate:
	$(DC) up migration

PHONY:infra
infra:
	$(DC) stop queue db 
	$(DC) rm -f queue db 
	$(DC) up -d queue db 

PHONY:restart
restart:
	$(DC) stop queue db 
	$(DC) rm -f queue db 
	$(DC) up -d queue db 

PHONY:logs
logs:
	$(DC) logs -f producer consumer

PHONY:stop
stop:
	$(DC) stop

PHONY: test
test:
	go test -v -race ./...

PHONY:check
check:
	golangci-lint run ./...

PHONY:build
build:
	$(DC) build producer consumer

PHONY:push
push:
	$(DC) push producer consumer