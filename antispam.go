package main

import (
	"fmt"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ===== DATA ADMIN =====
const (
	ADMIN_USER_ID  int64  = 7348614053
	ADMIN_USERNAME string = "JohanSetia"
	ADMIN_NAME     string = "Mr JOHAN (MACGYVER CODERMAN)"
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

// IsAdminID menyemak sama ada user ID adalah Admin
func IsAdminID(userID int64) bool {
	return userID == ADMIN_USER_ID
}

// CheckSpam akan memulangkan 'true' jika user disahkan spammer
// PENTING: Fungsi ini TIDAK akan mengesan spam untuk Admin
func CheckSpam(userID int64) bool {
	// Admin dikecualikan dari sistem anti-spam
	if IsAdminID(userID) {
		return false // Admin tidak akan dianggap spammer
	}
	
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
// PENTING: Fungsi ini TIDAK akan menjalankan ban untuk Admin
func ExecuteAutoBan(bot *tgbotapi.BotAPI, chatID int64, userID int64, username string) {
	// Langkah keselamatan: Jangan ban Admin
	if IsAdminID(userID) {
		logMsg := fmt.Sprintf("⚠️ PERHATIAN: Percubaan ban Admin dikesan! User: @%s (ID: %d) - TINDAKAN DIBATALKAN", username, userID)
		fmt.Println(logMsg)
		
		// Hantar notifikasi kepada Admin tentang percubaan ini
		adminAlert := fmt.Sprintf(
			"🛡️ **SISTEM KESELAMATAN**\n\n"+
			"Percubaan untuk menjalankan Auto-Ban ke atas Admin dikesan dan telah **DIBATALKAN**.\n\n"+
			"**Detail:**\n"+
			"👤 Username: @%s\n"+
			"🆔 User ID: `%d`\n"+
			"📋 Nama: %s\n"+
			"⏰ Masa: %s\n\n"+
			"_Sistem melindungi Admin daripada sekatan automatik._", 
			username, userID, ADMIN_NAME, time.Now().Format("2006-01-02 15:04:05"))
		
		msg := tgbotapi.NewMessage(ADMIN_USER_ID, adminAlert)
		msg.ParseMode = tgbotapi.ModeMarkdown
		bot.Send(msg)
		return
	}
	
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
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.DisableWebPagePreview = false
	bot.Send(msg)

	// 3. Laporkan kepada Admin (Mr JOHAN) supaya tahu ada 'pelanggan' baru nak bayar denda
	adminLog := fmt.Sprintf(
		"📢 **RADAR ALERT: AUTO-BAN**\n\n"+
		"👤 User: @%s\n"+
		"🆔 ID: `%d`\n"+
		"📋 Status: Menunggu Saman\n"+
		"⏰ Masa: %s", 
		username, userID, time.Now().Format("2006-01-02 15:04:05"))
	
	report := tgbotapi.NewMessage(ADMIN_USER_ID, adminLog)
	report.ParseMode = tgbotapi.ModeMarkdown
	bot.Send(report)
}

// UnbanUser - Fungsi untuk membuang sekatan (untuk kegunaan Admin)
func UnbanUser(bot *tgbotapi.BotAPI, adminID int64, targetID int64, chatID int64) error {
	// Pastikan hanya Admin boleh unban
	if !IsAdminID(adminID) {
		return fmt.Errorf("hanya Admin boleh menggunakan fungsi unban")
	}
	
	// Logik untuk unban dari GitHub akan ditambah di sini
	// (perlu diintegrasikan dengan fungsi dari terms.go)
	
	notisUnban := fmt.Sprintf(
		"✅ **NOTIS PENARIKAN SEKATAN**\n\n"+
		"Akaun anda (ID: `%d`) telah **DINYAHSEKAT** oleh Admin.\n\n"+
		"Anda kini boleh menggunakan bot semula. Sila taip /start untuk mula.", targetID)
	
	msg := tgbotapi.NewMessage(targetID, notisUnban)
	msg.ParseMode = tgbotapi.ModeMarkdown
	bot.Send(msg)
	
	return nil
}
