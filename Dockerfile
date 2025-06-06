# syntax=docker/dockerfile:1


FROM golang:1.24.3-alpine
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download


COPY . .
RUN go build -o bot ./cmd/bot/main.go


CMD ["./bot"]
