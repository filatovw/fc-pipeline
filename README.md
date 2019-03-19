# Code Challenge:

## Develop two separate executables that will:

### 1st Program:
Process CSV files and send each row/data to a message queue*. An example CSV file (https://gist.github.com/sarpdag/07da1e0614124b28b1d12e8ac410fa1a)

### 2nd Program:
Consume the messages that are sent by the 1st program and insert them into a database table**. There can only be one record with the same email address. Imagine this program will be running on multiple servers or even on the same server with multiple processes at the same time.

* Can be any production-ready distributed system such as redis, rabbitmq, sqs, or even kafka.

** Can be any modern relational database but the solution is better to be independent of the database server.

Tips And Tricks:

- Pay attention to the details in the description.
- Try to create a project structure you would use in production.
- Think about edge cases, either cover by code(best) or write a todo comment or document the edge cases/or what could go wrong.
- Any programming language can be used.

# Solution

Read more about solution [here](./docs/solution.md)