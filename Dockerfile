FROM golang:1.22-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /app
COPY . .
# Kita senaraikan fail untuk debug kalau gagal
RUN ls -l
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o telebot main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/telebot .

# Guna wildcard (*) supaya kalau fail tak ada, build TAKKAN gagal
COPY --from=builder /app/markdown.jso* ./
COPY --from=builder /app/selamat_datang.mp* ./

ENV GODEBUG=netdns=go
ENV PORT=7860
EXPOSE 7860
CMD ["./telebot"]
