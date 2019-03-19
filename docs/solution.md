# Solution

Services:

* Producer - reads CSV rows, process them and pushes into queue
* Consumer - reads from queue and stores into database

Both services are written in Golang.

Queue:

* NATS Streaming

Database:

* Postgres