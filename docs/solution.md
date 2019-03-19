# Solution

Services:

* Producer - reads CSV rows, process them and pushes into queue
* Consumer - reads from queue and stores into database

Both services are written in Golang.
I thought about Python with Celery and RabbitMQ, but this is a vendor lock which I wanted to avoid.
Besides that Golang works faster for data transferring, but Python is extremely good to data processing which is not the case.

Queue:

* RabbitMQ. This queue provides can store messages on a disk and it supports message acknowledgement. 

I tried NATS Streaming, but it doesn't support such functionality and it is possible only with external storage which makes it harder to support.


Database:

* Postgres

There could be any other database providing UNIQUE index.


TODO:

[ ] Add PG driver
[x] DB migration tool
[ ] move code from `/cmd` folder to separate folders
[ ] configure apps from environment variables for passing credentials