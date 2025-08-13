FROM golang:1.24.6-alpine AS builder
WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN go build -o /app/pvz -ldflags="-w -s" ./cmd/app

EXPOSE 16700

ENTRYPOINT [ "/app/pvz" ]
