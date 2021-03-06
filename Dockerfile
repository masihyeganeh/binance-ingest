FROM golang:1.15.8

RUN mkdir /app

ADD . /app

WORKDIR /app

RUN go build -o ingest cmd/main.go

ENV symbols=BNBBTC

CMD ["/app/ingest"]
