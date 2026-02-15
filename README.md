# CRYPTORIAN-TELEBOT

Cryptorian-Telebot ialah bot Telegram ringkas yang dibangunkan menggunakan Go. Projek ini direka sebagai rangka asas yang mudah disesuaikan untuk menambah perintah Telegram, sambungan API (contoh: data pasaran kripto) dan penyimpanan (contoh: SQLite/Postgres).

> Bahasa: Bahasa Malaysia

## Ciri-ciri
- Integrasi dengan Telegram Bot API
- Sokongan arahan asas seperti `/start`, `/help`
- Rangka asas untuk menambah arahan dan pengendalian mesej
- Contoh sambungan ke API luaran dan pangkalan data (boleh dikembangkan)
- Ringkas, mudah dibaca dan sesuai untuk pembangunan pantas

## Keperluan
- Go (sila rujuk `go.mod` untuk versi yang disyorkan)
- Token Telegram Bot (dapatkan daripada [@BotFather](https://t.me/BotFather))
- (Opsional) Kunci API untuk perkhidmatan luaran
- (Opsional) Pangkalan data (SQLite/Postgres) jika diperlukan

## Persediaan (Setup)
1. Clone repositori:
```bash
git clone https://github.com/<username>/<repo>.git
cd <repo>
```

2. Tetapkan pembolehubah persekitaran untuk token Telegram:
```bash
export TELEGRAM_BOT_TOKEN="masukkan_token_anda_di_sini"
```

3. (Opsional) Tetapkan pembolehubah persekitaran tambahan:
```bash
export API_KEY="kunci_api_anda"
export DATABASE_URL="postgres://user:pass@host:port/dbname"
```

4. Muat turun dependensi:
```bash
go mod download
```

Jika projek menggunakan fail `go.mod`, anda boleh semak versi Go yang disyorkan dengan membuka fail tersebut dan mengemaskini persekitaran pembangunan anda jika perlu.

## Menjalankan Projek (Development)
Cara paling ringkas untuk menjalankan semasa pembangunan:
```bash
go run ./cmd/cryptorian-telebot
```
atau, jika titik masuk adalah `main.go` di akar projek:
```bash
go run main.go
```

## Membangun (Build)
Untuk bina binari:
```bash
go build -o cryptorian-telebot ./cmd/cryptorian-telebot
./cryptorian-telebot
```
(Ubah jalan `./cmd/cryptorian-telebot` jika titik masuk projek anda berbeza.)

## Docker (Pilihan)
Contoh ringkas `Dockerfile`:
```dockerfile
FROM golang:1.20-alpine

WORKDIR /app
COPY . .

RUN go mod download
RUN go build -o cryptorian-telebot ./cmd/cryptorian-telebot

ENV TELEGRAM_BOT_TOKEN=""

CMD ["./cryptorian-telebot"]
```

Untuk bina dan jalankan kontena:
```bash
docker build -t cryptorian-telebot .
docker run -e TELEGRAM_BOT_TOKEN="$TELEGRAM_BOT_TOKEN" cryptorian-telebot
```

## Struktur Projek (Contoh)
Susunan fail mungkin seperti berikut. Sesuaikan mengikut struktur sebenar repo:
```
.
├── cmd/
│   └── cryptorian-telebot/   # titik masuk aplikasi (main.go)
├── internal/                  # pakej dalaman
├── pkg/                       # pakej boleh kongsi
├── configs/                   # konfigurasi
├── scripts/                   # skrip bantu
├── go.mod
├── go.sum
└── README.md
```

## Konfigurasi & Keselamatan
- Simpan token dan kunci sensitif menggunakan pembolehubah persekitaran atau sistem pengurusan rahsia (Vault, GitHub Secrets, dll.).
- Jangan commit token atau kunci API ke dalam kawalan versi.
- Jika menggunakan pangkalan data, letakkan URL sambungan dalam `DATABASE_URL`.

## Logging & Debugging
- Log dijana ke stdout; gunakan `docker logs` atau `journalctl` di persekitaran pengeluaran.
- Untuk pembangunan, jalankan aplikasi dengan `go run` dan pantau output konsol.
- Tambah tahap logging yang sesuai (contoh: debug/info/error) mengikut keperluan.

## Menambah Ciri / Menyumbang
- Buka isu (issue) untuk cadangan ciri atau melaporkan pepijat.
- Buat cawangan (branch) untuk kerja ciri baharu dan hantar Pull Request.
- Sertakan penerangan jelas, langkah untuk menguji, dan contoh input/output jika boleh.
- Ikuti garis panduan gaya kod yang ada dalam repo (jika disediakan).

## Ujian
- Tambah ujian unit untuk fungsi yang penting.
- Jalankan ujian dengan:
```bash
go test ./...
```

## Lesen
Sila tambah fail `LICENSE` yang sesuai (contoh: MIT) jika anda mahu memberikan kebenaran penggunaan terbuka. Jika anda mahu saya sediakan templat lesen (contoh: MIT), beritahu dan saya akan tambahkan.

## Soalan Lazim (FAQ)
- Bagaimana nak dapatkan token bot?  
  \- Buka perbualan dengan [@BotFather](https://t.me/BotFather) di Telegram dan ikut arahan untuk buat bot baharu.
- Bagaimana menambah arahan baharu?  
  \- Tambah pengendali arahan di bahagian yang mengendalikan mesej/updates (lihat pakej `cmd/cryptorian-telebot` atau `internal` bergantung kepada struktur).

## Hubungi / Sokongan
Jika anda perlukan bantuan lanjut atau mahu README disesuaikan berdasarkan struktur sebenar repo dan `go.mod`, tampal kandungan `go.mod` di sini atau beri akses bacaan ke repo — saya akan kemaskini README mengikut struktur sebenar.

---

Terima kasih kerana menggunakan CRYPTORIAN-TELEBOT — semoga projek ini memudahkan pembangunan bot Telegram anda!

---

### ⚖️ TERMA & SYARAT PENGGUNAAN

Dengan menggunakan **CRYPTORIAN-TELEBOT**, anda dianggap telah membaca dan bersetuju dengan terma berikut:

**1. TUJUAN SISTEM**
• Bot ini adalah alat bantuan pengurusan rujukan (referral) dan tuntutan ganjaran Worldcoin secara sukarela.
• Kami bukan wakil rasmi syarikat Worldcoin atau mana-mana badan kewangan.

**2. BUKAN NASIHAT KEWANGAN**
• Segala maklumat dalam bot ini bersifat pendidikan sahaja.
• Kami tidak bertanggungjawab atas sebarang keputusan pelaburan atau kerugian nilai aset digital anda.

**3. INTEGRITI & ANTI-PENIPUAN**
• Sebarang cubaan memanipulasi sistem rujukan menggunakan akaun palsu (sybil attack) atau bot luaran akan disekat serta-merta.
• Ganjaran hanya akan diproses bagi penyertaan yang sah mengikut syarat sistem.

**4. KESELAMATAN DATA & AKAUN**
• Kami tidak akan pernah meminta kata laluan, private key, atau seed phrase anda.
• Anda bertanggungjawab sepenuhnya untuk menjaga keselamatan akaun Telegram dan aplikasi World App anda sendiri.

**5. HAD LIABILITI**
• Pembangun tidak bertanggungjawab atas kegagalan teknikal, ralat transaksi pada rangkaian blockchain, atau perubahan polisi pihak ketiga yang menjejaskan ganjaran.

---
**"Sertai dengan bijak, fahami risiko anda."**
*© 2026 Cryptorian World MY*

