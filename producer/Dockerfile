FROM golang:1.12 as builder
COPY . /src
WORKDIR /src
RUN go mod download
RUN go build -o /usr/bin/producer ./producer

FROM debian:stretch
WORKDIR /usr/bin/
COPY --from=builder /usr/bin/producer .
VOLUME ["/data"]
CMD ["/usr/bin/producer"]