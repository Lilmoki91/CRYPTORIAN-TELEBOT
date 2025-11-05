# Gunakan imej Go yang betul untuk langkah pembinaan
FROM golang:1.21-alpine AS builder

# Tetapkan direktori kerja di dalam kontena pembinaan
WORKDIR /app

# Salin fail go.mod dan go.sum dan muat turun kebergantungan
COPY go.mod go.sum ./
RUN go mod download

# Salin semua fail sumber yang lain (main.go dan markdown.json)
COPY . .

# Bina aplikasi Go dan hasilkan fail boleh laksana bernama 'bot'
# CGO_ENABLED=0 adalah penting untuk pembinaan statik yang berfungsi di Alpine
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bot main.go

# --- Langkah Akhir (Imej Paling Kecil) ---
# Gunakan imej Alpine yang kecil untuk imej akhir
FROM alpine:latest

# Tetapkan direktori kerja di dalam imej akhir
WORKDIR /

# Pasang sijil CA untuk sambungan HTTPS berfungsi
RUN apk --no-cache add ca-certificates

# Salin fail boleh laksana 'bot' dari langkah pembinaan
COPY --from=builder /bot /

# Salin fail konfigurasi kritikal (markdown.json) dari langkah pembinaan
# Fail ini kini disalin ke root '/' kontena, di mana bot perlu mencarinya.
COPY --from=builder /app/markdown.json /markdown.json

# Tentukan arahan lalai untuk menjalankan aplikasi
# Bot anda akan dijalankan apabila kontena dimulakan
ENTRYPOINT ["/bot"]
