FROM golang:latest

WORKDIR /app

COPY src/go.mod .
COPY src/go.sum .

RUN go mod download
