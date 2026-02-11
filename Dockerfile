# syntax=docker/dockerfile:1

FROM golang:1.22 AS builder


WORKDIR /app

COPY . .


RUN go mod tidy

RUN go build -ldflags="-w -s" -o telebot main.go


FROM gcr.io/distroless/base-debian12

WORKDIR /

COPY --from=builder /app/telebot /telebot

COPY config/ /config/


# Expose port 8080 (Cloud Run default)

EXPOSE 8080


# Run as non-root user for security (user 65532 = nonroot di distroless)

USER 65532:65532


CMD ["/telebot"]

