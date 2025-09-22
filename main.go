package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// --- STRUCTS TO MATCH MARKDOWN.JSON ---
type Step struct {
	Title  string   `json:"title"`
	Desc   string   `json:"desc"`
	Images []string `json:"images"`
}

type Important struct {
	Title string   `json:"title"`
	Notes []string `json:"notes"`
}

type Guide struct {
	Title     string    `json:"title"`
	Steps     []Step    `json:"steps"`
	Important Important `json:"important"`
}

// --- GLOBAL VARIABLES ---
var messageIDsToDelete = make(map[int64][]int)
var userLastActivity = make(map[int64]time.Time)

// --- HELPER FUNCTIONS ---

// Escape text untuk MarkdownV2
func escapeMarkdownV2(text string) string {
	re := regexp.MustCompile(`([_{}\[\]()~>#\+\-=|.!])`)
	return re.ReplaceAllString(text, `\$1`)
}

// Buat main menu keyboard
func createMainMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🚀 Claim WorldCoin"),
			tgbotapi.NewKeyboardButton("💰 Wallet HATA"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("💸 Cashout ke Bank"),
			tgbotapi.NewKeyboardButton("📢 Channel"),
			tgbotapi.NewKeyboardButton("👨‍💻 Admin"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🆔 Dapatkan Chat ID"),
			tgbotapi.NewKeyboardButton("🔄 Reset"),
			tgbotapi.NewKeyboardButton("🔙 Menu Utama"),
		),
	)

	keyboard.OneTimeKeyboard = false
	keyboard.ResizeKeyboard = true
	return keyboard
}

// Buat inline keyboard
func createInlineKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🚀 Claim", "claim_cmd"),
			tgbotapi.NewInlineKeyboardButtonData("💰 Wallet", "wallet_cmd"),
			tgbotapi.NewInlineKeyboardButtonData("💸 Cashout", "cashout_cmd"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("📢 Channel", "https://t.me/cucikripto"),
			tgbotapi.NewInlineKeyboardButtonURL("👨‍💻 Admin", "https://t.me/johansetia"),
		),
	)
}

// Hantar panduan penuh
func sendDetailedGuide(bot *tgbotapi.BotAPI, chatID int64, guide Guide) {
	// Title utama
	titleMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("📖 *%s*", escapeMarkdownV2(guide.Title)))
	titleMsg.ParseMode = tgbotapi.ModeMarkdownV2
	sentTitle, _ := bot.Send(titleMsg)
	messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentTitle.MessageID)

	// Langkah demi langkah
	for _, step := range guide.Steps {
		var caption strings.Builder
		caption.WriteString(fmt.Sprintf("*%s*\n\n", escapeMarkdownV2(step.Title)))
		caption.WriteString(escapeMarkdownV2(step.Desc))

		if len(step.Images) == 0 {
			msg := tgbotapi.NewMessage(chatID, caption.String())
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			sentMsg, _ := bot.Send(msg)
			messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)

		} else if len(step.Images) == 1 {
			photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(step.Images[0]))
			photo.Caption = caption.String()
			photo.ParseMode = tgbotapi.ModeMarkdownV2
			sentMsg, _ := bot.Send(photo)
			messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)

		} else {
			// Album multi-gambar (FIX)
			var mediaGroup []tgbotapi.InputMedia
			for i, imgURL := range step.Images {
				photo := tgbotapi.NewInputMediaPhoto(tgbotapi.FileURL(imgURL))
				if i == 0 {
					photo.Caption = caption.String()
					photo.ParseMode = tgbotapi.ModeMarkdownV2
				}
				mediaGroup = append(mediaGroup, photo)
			}
			album := tgbotapi.NewMediaGroup(chatID, mediaGroup)

			sentMsgs, err := bot.Send(album)
			if err == nil {
				for _, m := range sentMsgs {
					messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], m.MessageID)
				}
			}
		}
	}

	// Nota penting
	if len(guide.Important.Notes) > 0 {
		var notesBuilder strings.Builder
		notesBuilder.WriteString(fmt.Sprintf("*%s*\n\n", escapeMarkdownV2(guide.Important.Title)))
		for _, note := range guide.Important.Notes {
			notesBuilder.WriteString(fmt.Sprintf("• %s\n", escapeMarkdownV2(note)))
		}
		msg := tgbotapi.NewMessage(chatID, notesBuilder.String())
		msg.ParseMode = tgbotapi.ModeMarkdownV2
		sentMsg, _ := bot.Send(msg)
		messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
	}
}

func main() {
	// --- LOAD DATA ---
	jsonData, err := os.ReadFile("markdown.json")
	if err != nil {
		log.Panicf("Error reading markdown.json: %v", err)
	}
	var guides map[string]Guide
	if err := json.Unmarshal(jsonData, &guides); err != nil {
		log.Panicf("JSON parsing error: %v", err)
	}

	// --- TOKEN ---
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Panic("TELEGRAM_BOT_TOKEN environment variable not set")
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panicf("Token error: %v", err)
	}

	bot.Debug = false
	log.Printf("Bot started: @%s", bot.Self.UserName)

	// --- KEYBOARDS ---
	initialKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("▶️ Papar Menu Utama", "show_main_menu"),
		),
	)
	mainMenuKeyboard := createMainMenuKeyboard()

	// --- LOOP UPDATE ---
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		// Skip old messages
		if update.Message != nil && update.Message.Time().Before(time.Now().Add(-2*time.Minute)) {
			continue
		}

		// Inline callback
		if update.CallbackQuery != nil {
			bot.Request(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
			if update.CallbackQuery.Data == "show_main_menu" {
				chatID := update.CallbackQuery.Message.Chat.ID
				text := "🤖 *SMART AASA BOT* - Panduan Lengkap Kripto\n\n" +
					"📋 *Gunakan butang di bawah atau ketik command:*\n" +
					"🚀 /claim - Panduan Claim Worldcoin\n" +
					"💰 /wallet - Setup Wallet HATA\n" +
					"💸 /cashout - Cashout ke Bank\n" +
					"🆔 /chatid - Dapatkan Chat ID\n\n" +
					"🔗 *Link Penting:*\n" +
					"📢 Channel: @cucikripto\n" +
					"👨‍💻 Admin: @johansetia"

				msg := tgbotapi.NewMessage(chatID, text)
				msg.ParseMode = tgbotapi.ModeMarkdownV2
				msg.ReplyMarkup = mainMenuKeyboard
				sentMsg, _ := bot.Send(msg)
				messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)

				// Hapus butang asal
				deleteMsg := tgbotapi.NewDeleteMessage(chatID, update.CallbackQuery.Message.MessageID)
				bot.Request(deleteMsg)
			}
			continue
		}

		if update.Message == nil || update.Message.Text == "" {
			continue
		}

		chatID := update.Message.Chat.ID
		userLastActivity[chatID] = time.Now()

		// Security check untuk suspicious links
		if strings.Contains(strings.ToLower(update.Message.Text), "http") ||
			strings.Contains(strings.ToLower(update.Message.Text), "www.") ||
			strings.Contains(strings.ToLower(update.Message.Text), ".ba.ba") {

			warning := "⚠️ *PERINGATAN KESELAMATAN*\n\n" +
				"Hindari link yang mencurigakan!\n\n" +
				"✅ *Gunakan hanya link official:*\n" +
				"• https://worldcoin.org\n" +
				"• https://hata.io\n\n" +
				"❌ *Jangan klik link random* dari unknown sources"

			msg := tgbotapi.NewMessage(chatID, warning)
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			bot.Send(msg)
		}

		switch update.Message.Text {
		case "/start", "🔙 Menu Utama":
			welcomeText := "👋 *Selamat Datang di SMART AASA BOT!*\n\n" +
				"Saya di sini untuk bantu anda dengan:\n" +
				"• 🚀 Claim Worldcoin percuma\n" +
				"• 💰 Setup Wallet HATA\n" +
				"• 💸 Cashout ke bank lokal\n\n" +
				"Tekan butang di bawah untuk mula!"

			msg := tgbotapi.NewMessage(chatID, welcomeText)
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			msg.ReplyMarkup = initialKeyboard
			sentMsg, _ := bot.Send(msg)
			messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)

		case "🚀 Claim WorldCoin":
			text := "🔐 *VERIFIKASI WORLDCOIN*\n\n" +
				"Untuk claim Worldcoin, sila:\n\n" +
				"1. ✅ Download *World App* dari official store\n" +
				"2. ✅ Daftar dengan dokumen yang valid\n" +
				"3. ✅ Selesaikan verifikasi wajah\n\n" +
				"🌐 *Link Official:* [Klik sini](https://worldcoin.org/join/4RH0OTE)\n\n" +
				"📝 *Panduan Lengkap:* ketik `/claim`"

			msg := tgbotapi.NewMessage(chatID, text)
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			msg.ReplyMarkup = createInlineKeyboard()
			sentMsg, _ := bot.Send(msg)
			messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)

		case "💰 Wallet HATA":
			text := "📱 *HATA WALLET*\n\n" +
				"Daftar wallet HATA untuk terima transfer:\n\n" +
				"🌐 *Link Official:* [Klik sini](https://hata.io/signup?ref=HDX8778)\n\n" +
				"📝 *Panduan Lengkap:* ketik `/wallet`\n\n" +
				"💡 *Tips:* Simpan private key dengan aman!"

			msg := tgbotapi.NewMessage(chatID, text)
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			msg.ReplyMarkup = createInlineKeyboard()
			sentMsg, _ := bot.Send(msg)
			messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)

		case "💸 Cashout ke Bank":
			text := "🏦 *CASHOUT KE BANK*\n\n" +
				"Tukar kripto ke ringgit Malaysia:\n\n" +
				"📝 *Panduan Lengkap:* ketik `/cashout`\n\n" +
				"⏱ *Proses:* 1-3 hari bekerja\n" +
				"💸 *Minimum withdrawal:* RM50"

			msg := tgbotapi.NewMessage(chatID, text)
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			msg.ReplyMarkup = createInlineKeyboard()
			sentMsg, _ := bot.Send(msg)
			messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)

		case "📢 Channel":
			text := "📢 *JOIN CHANNEL KAMI*\n\n" +
				"Dapatkan info terkini tentang kripto:\n\n" +
				"👉 [@cucikripto](https://t.me/cucikripto)\n\n" +
				"✨ *Content:*\n" +
				"• News & updates\n" +
				"• Tips & tricks\n" +
				"• Announcements"

			msg := tgbotapi.NewMessage(chatID, text)
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			sentMsg, _ := bot.Send(msg)
			messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)

		case "👨‍💻 Admin":
			text := "👨‍💻 *SUPPORT ADMIN*\n\n" +
				"Hubungi admin untuk bantuan:\n\n" +
				"👉 [@johansetia](https://t.me/johansetia)\n\n" +
				"🕒 *Waktu Response:*\n" +
				"• Weekdays: 9AM-6PM\n" +
				"• Weekends: 10AM-4PM\n\n" +
				"📞 *Untuk urgent:* PM terus admin"

			msg := tgbotapi.NewMessage(chatID, text)
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			sentMsg, _ := bot.Send(msg)
			messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)

		case "🆔 Dapatkan Chat ID":
			userInfo := fmt.Sprintf(
				"📋 *INFO AKUN ANDA*\n\n"+
					"🆔 *Chat ID:* `%d`\n"+
					"👤 *Nama:* %s %s\n"+
					"🔖 *Username:* @%s\n"+
					"📞 *User ID:* %d\n"+
					"💬 *Tipe Chat:* %s\n\n"+
					"⚠️ *Jangan share info ini dengan orang lain!*",
				chatID,
				escapeMarkdownV2(update.Message.From.FirstName),
				escapeMarkdownV2(update.Message.From.LastName),
				update.Message.From.UserName,
				update.Message.From.ID,
				update.Message.Chat.Type,
			)

			msg := tgbotapi.NewMessage(chatID, userInfo)
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			sentMsg, _ := bot.Send(msg)
			messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)

		case "/menu", "📋 Menu":
			text := "🤖 *SMART AASA BOT* - Panduan Lengkap Kripto\n\n" +
				"📋 *Pilih menu di bawah:*"

			msg := tgbotapi.NewMessage(chatID, text)
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			msg.ReplyMarkup = mainMenuKeyboard
			sentMsg, _ := bot.Send(msg)
			messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)

		// Commands
		case "/claim":
			sendDetailedGuide(bot, chatID, guides["worldcoin_registration_guide"])

		case "/wallet", "/wallet hata":
			sendDetailedGuide(bot, chatID, guides["hata_setup_guide"])

		case "/cashout":
			sendDetailedGuide(bot, chatID, guides["cashout_guide"])

		case "/chatid":
			userInfo := fmt.Sprintf(
				"🆔 *Chat ID Anda:* `%d`\n\n"+
					"💡 *Untuk support/admin purposes*",
				chatID,
			)

			msg := tgbotapi.NewMessage(chatID, userInfo)
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			sentMsg, _ := bot.Send(msg)
			messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)

		case "/reset", "🔄 Reset":
			// Padam semua mesej bot
			deletedCount := 0
			for _, id := range messageIDsToDelete[chatID] {
				_, err := bot.Request(tgbotapi.NewDeleteMessage(chatID, id))
				if err == nil {
					deletedCount++
				}
			}
			messageIDsToDelete[chatID] = nil

			// Hantar fresh start
			text := fmt.Sprintf("🔄 *Reset Berjaya!*\n%d mesej telah dipadam.\n\nTekan butang di bawah untuk mula semula.", deletedCount)
			msg := tgbotapi.NewMessage(chatID, text)
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			msg.ReplyMarkup = initialKeyboard
			sentMsg, _ := bot.Send(msg)
