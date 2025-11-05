# Tahap 1: Pembina
FROM golang:1.21-alpine AS builder

# Tetapkan direktori kerja
WORKDIR /app

# Salin fail dependensi (go.mod, go.sum)
COPY go.mod go.sum ./
RUN go mod download

# Salin kod bot dan fail konfigurasi
COPY main.go markdown.json ./

# Bina binari Go
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o cryptorian-telebot .

# Tahap 2: Kontena Akhir
FROM alpine:latest

# Tambah sijil CA
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Salin binari yang dibina dari tahap pembina
COPY --from=builder /app/cryptorian-telebot .

# CIPTA FOLDER DAN SALIN FAIL KONFIGURASI KRITIKAL
# Kod anda memerlukan markdown.json di dalam folder /config/
RUN mkdir -p /config/
COPY --from=builder /app/markdown.json /config/markdown.json 

EXPOSE 8080

CMD ["./cryptorian-telebot"]
