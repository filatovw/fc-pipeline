# Solution

Requirements:

- docker
- docker-compose >= 3.6
- make-tool

Services:

* Producer - reads CSV rows, process them and pushes into queue.
* Consumer - reads from queue and stores into database. 
Consumer service works until someone stops it explicitly.


Both services are written in Golang.
I thought about Python with Celery and RabbitMQ, but this is a vendor lock which I wanted to avoid.
Besides that Golang works faster for data transferring, but Python is extremely good to data processing which is not the case.

Queue:

* RabbitMQ. This queue can store messages on a disk and it supports message acknowledgement. 

I tried NATS Streaming, but it doesn't provide exactly once delivery and it is possible only with external storage which makes it hard to maintain.

Database:

* Postgres

There could be any other database providing UNIQUE constraint.

I prefer to execute Queue and Database with `Docker`.


# Usage

Simple way:

1) start infrastructure
2) apply db migration 
3) start producer and consumers

Start infrastructure services:

    make infra

Apply database migration:

    make migrate-db

Start producer and consumers:

    make start

Now you can monitor state of execution using logs:

    make logs

# Maintenance

Stop infrastructure and all services:

    make stop

Build services:

    make build

Push images dockerhub:

    make push

Drop database:

    make drop-db

Delete old volumes and unused images:

    make prune

# Development

Install [golangcli-lint](https://github.com/golangci/golangci-lint#install)

Check code:

    make check

Run tests:

    make test

Build producer binary on a local machine:

    make build-producer

Build consumer binary on a local machine:

    make build-consumer

Delete binaries:

    make clean

If you need to clean up data please do it manually:

    rm -rf ./volumes/*

Once you've built binaries you can execute them manually

Producer:

    ./bin/producer --help
        Usage of ./bin/producer:
        -file string
                path to CSV file
        -parallel int
                number of workers
        -queue-addr string
                env: FC_PRODUCER_QUEUE_ADDR. Address of queue [default: 0.0.0.0:5672]
        -queue-pass string
                env: FC_PRODUCER_QUEUE_PASS. Queue pass [default: fcpass]
        -queue-user string
                env: FC_PRODUCER_QUEUE_USER. Queue user [default: fcuser]

Consumer:

    ./bin/consumer --help
        Usage of ./bin/consumer:
        -db-host string
                env: FC_CONSUMER_DB_HOST. Database host [default: 0.0.0.0]
        -db-pass string
                env: FC_CONSUMER_DB_PASS. Database pass [default: fcpass]
        -db-port int
                env: FC_CONSUMER_DB_PORT. Database port [default: 5432]
        -db-user string
                env: FC_CONSUMER_DB_USER. Database user [default: fcuser]
        -parallel int
                number of workers
        -queue-addr string
                env: FC_CONSUMER_QUEUE_ADDR. Address of queue [default: 0.0.0.0:5672]
        -queue-pass string
                env: FC_CONSUMER_QUEUE_PASS. Queue pass [default: fcpass]
        -queue-user string
                env: FC_CONSUMER_QUEUE_USER. Queue user [default: fcuser]
            
# Repository structure

`/bin` - compiled binaries

`/consumer` - consumer service. All code is here including `Dockerfile`

`/consumer/migrations` - sql migrations

`/docs` - documentation for this solution

`/env` - environment files. There is only one environment, but in reality they should be stored somewhere else separated by environment (dev,test, stage, prod, etc)

`/libs` - common code. I prefer not to rely on 3d-party tools too much and use included batteries whereas possible

`/producer` - procuder service. All code is here including `Dockerfile`

`/producer/data` - csv files. I added randomly generated files with 1000 and 10000 records

`/volumes` - place for queue and database data


# Corner cases

- Queue is unreachable -> producer: show number of processed rows.

For now: you delete first N rows and start producer with this new file or just process it from the beginning. 
Database will not allow us to insert records with emails that already exists.

Better: accept offset in command line arguments and start to process from this offset.

- Queue is unreachable -> consumer: stop and exit

For now: fix queue issue then restart consumers manually

Better: monitor state of queue and manage consumers automatically

- Database is unreachable -> consumer stop and exit

For now: fix database and restart consumers manually

Better: monitor state of database and manage consumers automatically 

- Producer stopped to work with panic -> exit code 2

For now: read error message and fix the code. Replay file again from the beginning

Better: store offset in external storage (eg. put it into Redis once per second)

- Consumer stopped to work with panic -> exit code 2

For now: read error message and fix the code. Then start consumers again