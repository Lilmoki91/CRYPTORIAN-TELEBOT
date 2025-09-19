# syntax=docker/dockerfile:1
FROM golang:1.22 AS builder

WORKDIR /app
COPY . .

RUN go mod tidy
RUN go build -o telebot main.go

FROM gcr.io/distroless/base-debian12
COPY --from=builder /app/telebot /telebot
COPY config/ /config/
CMD ["/telebot"]
