FROM golang:1.12 as builder
COPY . /src
WORKDIR /src
RUN go mod download
RUN go build -o /usr/bin/consumer ./consumer

FROM debian:stretch
WORKDIR /usr/bin/
COPY --from=builder /usr/bin/consumer .
CMD ["/usr/bin/consumer"]