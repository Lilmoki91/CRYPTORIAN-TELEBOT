package main

import (
	"fmt"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Struktur untuk menjejak aktiviti setiap user
type UserActivity struct {
	LastAction time.Time
	Count      int
}

var (
	spamMap = make(map[int64]*UserActivity)
	spamMu  sync.Mutex
	// Had limit: 5 tekanan dalam masa 3 saat
	threshold      = 5
	timeWindow     = 3 * time.Second
)

// CheckSpam akan memulangkan 'true' jika user disahkan spammer
func CheckSpam(userID int64) bool {
	spamMu.Lock()
	defer spamMu.Unlock()

	now := time.Now()
	activity, exists := spamMap[userID]

	if !exists {
		spamMap[userID] = &UserActivity{LastAction: now, Count: 1}
		return false
	}

	// Jika masih dalam tingkap masa (3 saat)
	if now.Sub(activity.LastAction) < timeWindow {
		activity.Count++
	} else {
		// Reset jika sudah lama tidak tekan
		activity.Count = 1
	}

	activity.LastAction = now

	// Jika melebihi had, aktifkan hukuman
	if activity.Count > threshold {
		return true
	}

	return false
}

// ExecuteAutoBan menjalankan hukuman dan menghantar notis denda
func ExecuteAutoBan(bot *tgbotapi.BotAPI, chatID int64, userID int64, username string) {
	// 1. Simpan rekod sekatan ke GitHub (Audit Log)
	reason := "AUTO-BAN: Melakukan kesalahan spamming butang/mesej"
	BanUser(userID, reason)

	// 2. Bina mesej notis sekatan dan denda
	// Gunakan Markdown Standard yang serasi dengan main.go
	notisSaman := fmt.Sprintf(
		"🚫 **AKAUN ANDA TELAH DISEKAT**\n\n"+
			"Sistem mengesan aktiviti spam yang melampau dari akaun anda.\n\n"+
			"**Tindakan:** Sekatan Kekal (Permanent Ban)\n\n"+
			"Untuk membuka semula sekatan ini, anda wajib:\n"+
			"1. Mengemukakan rayuan kepada Admin.\n"+
			"2. Menjelaskan denda kesalahan (Bayaran) jika ingin unlock.\n\n"+
			"👉 **Hubungi Admin untuk Rayuan:** [KLIK DI SINI](https://t.me/johansetia)\n\n"+
			"_Sila sertakan ID anda (%d) semasa membuat rayuan._", userID)

	msg := tgbotapi.NewMessage(chatID, notisSaman)
	msg.ParseMode = tgbotapi.ModeMarkdown // Guna Markdown Standard
	msg.DisableWebPagePreview = false      // Supaya link t.me nampak cantik
	bot.Send(msg)

	// 3. Laporkan kepada King (Admin) supaya King tahu ada 'pelanggan' baru nak bayar denda
	adminLog := fmt.Sprintf("📢 **RADAR ALERT: AUTO-BAN**\n\nUser: @%s\nID: `%d`\nStatus: Menunggu Saman", username, userID)
	report := tgbotapi.NewMessage(ADMIN_ID, adminLog)
	report.ParseMode = tgbotapi.ModeMarkdown
	bot.Send(report)
}
