FROM golang:latest

WORKDIR /go-test-api

COPY ./ /go-test-api

RUN go mod download

EXPOSE 8081

ENTRYPOINT go run main.go