FROM golang:1.18

WORKDIR /app

COPY src/go.mod ./
COPY src/go.sum ./
RUN go mod download

COPY src/main.go ./

EXPOSE 8000