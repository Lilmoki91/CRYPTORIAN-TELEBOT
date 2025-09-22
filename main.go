// main.go

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
var messageIDsToDelete = make(map[int64][]int) // Stores message IDs to delete per chat
var userLastActivity = make(map[int64]time.Time)

// --- HELPER FUNCTIONS ---

// Escape text untuk MarkdownV2 mengikut spesifikasi Telegram
func escapeMarkdownV2(text string) string {
	// Karakter yang perlu di-escape dalam MarkdownV2: _ * [ ] ( ) ~ ` > # + - = | { } . !
	// Menggunakan regex untuk mencari dan menggantikan
	re := regexp.MustCompile(`([_*\[\]()~` + "`" + `>#+\-=|{}.!])`)
	return re.ReplaceAllString(text, `\$1`)
}

// Buat main menu keyboard dengan settings yang betul
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

// Buat inline keyboard untuk easy navigation
func createInlineKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🚀 Claim", "claim_cmd"),
			tgbotapi.NewInlineKeyboardButtonData("💰 Wallet", "wallet_cmd"),
			tgbotapi.NewInlineKeyboardButtonData("💸 Cashout", "cashout_cmd"),
		),
		tgbotapi.NewInlineKeyboardRow(
			// URL dibersihkan daripada ruang tambahan
			tgbotapi.NewInlineKeyboardButtonURL("📢 Channel", "https://t.me/cucikripto"),
			tgbotapi.NewInlineKeyboardButtonURL("👨‍💻 Admin", "https://t.me/johansetia"),
		),
	)
}

// Hantar panduan penuh dari markdown.json
func sendDetailedGuide(bot *tgbotapi.BotAPI, chatID int64, guide Guide) {
	// Title utama
	titleMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("📖 *%s*", escapeMarkdownV2(guide.Title)))
	titleMsg.ParseMode = tgbotapi.ModeMarkdownV2
	sentTitle, err := bot.Send(titleMsg)
	if err == nil {
		messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentTitle.MessageID)
	} else {
		log.Printf("Error sending title message: %v", err)
	}

	// Langkah demi langkah
	for _, step := range guide.Steps {
		var caption strings.Builder
		caption.WriteString(fmt.Sprintf("*%s*\n\n", escapeMarkdownV2(step.Title)))
		caption.WriteString(escapeMarkdownV2(step.Desc))

		if len(step.Images) == 0 {
			// Hantar teks sahaja
			msg := tgbotapi.NewMessage(chatID, caption.String())
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			sentMsg, err := bot.Send(msg)
			if err == nil {
				messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
			} else {
				log.Printf("Error sending text step message: %v", err)
			}

		} else if len(step.Images) == 1 {
			// Hantar satu gambar dengan caption
			photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(step.Images[0]))
			photo.Caption = caption.String()
			photo.ParseMode = tgbotapi.ModeMarkdownV2
			sentMsg, err := bot.Send(photo)
			if err == nil {
				messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
			} else {
				log.Printf("Error sending photo message: %v", err)
			}

		} else {
			// Hantar album multi-gambar
			// InputMedia tidak menyokong ParseMode secara individu dalam v5
			// Caption hanya boleh diletakkan pada item pertama
			var mediaGroup []interface{}
			for i, imgURL := range step.Images {
				media := tgbotapi.NewInputMediaPhoto(tgbotapi.FileURL(imgURL))
				// Hanya item pertama yang boleh mempunyai caption
				if i == 0 {
					media.Caption = caption.String()
					// Media itu sendiri tidak memerlukan ParseMode, ia digunakan pada mesej induk
				}
				mediaGroup = append(mediaGroup, media)
			}
			album := tgbotapi.NewMediaGroup(chatID, mediaGroup)

			// Hantar album dan dapatkan slice mesej
			messages, err := bot.Send(album)
			if err == nil {
				// Tambahkan MessageID setiap mesej dalam album
				for _, msg := range messages {
					messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], msg.MessageID)
				}
			} else {
				log.Printf("Error sending media group: %v", err)
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
		sentMsg, err := bot.Send(msg)
		if err == nil {
			messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
		} else {
			log.Printf("Error sending important notes message: %v", err)
		}
	}
}

func main() {
	// --- LOAD DATA ---
	jsonData, err := os.ReadFile("markdown.json")
	if err != nil {
		log.Panic("Error reading markdown.json: ", err)
	}
	var guides map[string]Guide
	if err := json.Unmarshal(jsonData, &guides); err != nil {
		log.Panic("JSON parsing error: ", err)
	}

	// --- TOKEN ---
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Panic("TELEGRAM_BOT_TOKEN environment variable not set")
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic("Token error: ", err)
	}

	bot.Debug = false // Matikan debug mode di production
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
		// Skip old messages (anti-spam) - Lebih selamat menggunakan update.UpdateID
		if update.Message != nil && update.Message.Time().Before(time.Now().Add(-2*time.Minute)) {
			continue
		}

		// Inline callback
		if update.CallbackQuery != nil {
			// Acknowledge the callback query
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
			if err := bot.Request(callback); err != nil {
				log.Printf("Error sending callback: %v", err)
			}

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
				sentMsg, err := bot.Send(msg)
				if err == nil {
					messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
				} else {
					log.Printf("Error sending main menu message: %v", err)
				}

				// Padam butang asal (pesan yang mengandungi inline keyboard)
				deleteMsg := tgbotapi.NewDeleteMessage(chatID, update.CallbackQuery.Message.MessageID)
				if _, err := bot.Request(deleteMsg); err != nil {
					log.Printf("Error deleting original inline keyboard message: %v", err)
				}
			}
			continue // Proses update seterusnya
		}

		// Jika bukan callback, pastikan update.Message ada
		if update.Message == nil || update.Message.Text == "" {
			continue
		}

		chatID := update.Message.Chat.ID
		userLastActivity[chatID] = time.Now()

		// Security check untuk suspicious links
		lowerText := strings.ToLower(update.Message.Text)
		if strings.Contains(lowerText, "http") ||
			strings.Contains(lowerText, "www.") ||
			strings.Contains(lowerText, ".ba.ba") {

			warning := "⚠️ *PERINGATAN KESELAMATAN*\n\n" +
				"Hindari link yang mencurigakan!\n\n" +
				"✅ *Gunakan hanya link official:*\n" +
				"• https://worldcoin.org\n" +
				"• https://hata.io\n\n" +
				"❌ *Jangan klik link random* dari unknown sources"

			msg := tgbotapi.NewMessage(chatID, warning)
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Error sending security warning: %v", err)
			}
		}

		// Tangani perintah dan mesej teks
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
			sentMsg, err := bot.Send(msg)
			if err == nil {
				messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
			} else {
				log.Printf("Error sending /start message: %v", err)
			}

		case "🚀 Claim WorldCoin":
			// URL dibersihkan daripada ruang tambahan
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
			sentMsg, err := bot.Send(msg)
			if err == nil {
				messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
			} else {
				log.Printf("Error sending Claim WorldCoin message: %v", err)
			}

		case "💰 Wallet HATA":
			// URL dibersihkan daripada ruang tambahan
			text := "📱 *HATA WALLET*\n\n" +
				"Daftar wallet HATA untuk terima transfer:\n\n" +
				"🌐 *Link Official:* [Klik sini](https://hata.io/signup?ref=HDX8778)\n\n" +
				"📝 *Panduan Lengkap:* ketik `/wallet`\n\n" +
				"💡 *Tips:* Simpan private key dengan aman!"

			msg := tgbotapi.NewMessage(chatID, text)
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			msg.ReplyMarkup = createInlineKeyboard()
			sentMsg, err := bot.Send(msg)
			if err == nil {
				messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
			} else {
				log.Printf("Error sending Wallet HATA message: %v", err)
			}

		case "💸 Cashout ke Bank":
			text := "🏦 *CASHOUT KE BANK*\n\n" +
				"Tukar kripto ke ringgit Malaysia:\n\n" +
				"📝 *Panduan Lengkap:* ketik `/cashout`\n\n" +
				"⏱ *Proses:* 1-3 hari bekerja\n" +
				"💸 *Minimum withdrawal:* RM50"

			msg := tgbotapi.NewMessage(chatID, text)
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			msg.ReplyMarkup = createInlineKeyboard()
			sentMsg, err := bot.Send(msg)
			if err == nil {
				messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
			} else {
				log.Printf("Error sending Cashout message: %v", err)
			}

		case "📢 Channel":
			// URL dibersihkan daripada ruang tambahan
			text := "📢 *JOIN CHANNEL KAMI*\n\n" +
				"Dapatkan info terkini tentang kripto:\n\n" +
				"👉 [@cucikripto](https://t.me/cucikripto)\n\n" +
				"✨ *Content:*\n" +
				"• News & updates\n" +
				"• Tips & tricks\n" +
				"• Announcements"

			msg := tgbotapi.NewMessage(chatID, text)
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			sentMsg, err := bot.Send(msg)
			if err == nil {
				messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
			} else {
				log.Printf("Error sending Channel message: %v", err)
			}

		case "👨‍💻 Admin":
			// URL dibersihkan daripada ruang tambahan
			text := "👨‍💻 *SUPPORT ADMIN*\n\n" +
				"Hubungi admin untuk bantuan:\n\n" +
				"👉 [@johansetia](https://t.me/johansetia)\n\n" +
				"🕒 *Waktu Response:*\n" +
				"• Weekdays: 9AM-6PM\n" +
				"• Weekends: 10AM-4PM\n\n" +
				"📞 *Untuk urgent:* PM terus admin"

			msg := tgbotapi.NewMessage(chatID, text)
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			sentMsg, err := bot.Send(msg)
			if err == nil {
				messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
			} else {
				log.Printf("Error sending Admin message: %v", err)
			}

		case "🆔 Dapatkan Chat ID":
			user := update.Message.From
			var firstName, lastName, username string
			if user != nil {
				firstName = escapeMarkdownV2(user.FirstName)
				lastName = escapeMarkdownV2(user.LastName)
				username = user.UserName // Username tidak perlu di-escape untuk MarkdownV2
			}

			userInfo := fmt.Sprintf(
				"📋 *INFO AKUN ANDA*\n\n"+
					"🆔 *Chat ID:* `%d`\n"+
					"👤 *Nama:* %s %s\n"+
					"🔖 *Username:* @%s\n"+
					"📞 *User ID:* %d\n"+
					"💬 *Tipe Chat:* %s\n\n"+
					"⚠️ *Jangan share info ini dengan orang lain!*",
				chatID,
				firstName, lastName,
				username,
				user.ID,
				update.Message.Chat.Type,
			)

			msg := tgbotapi.NewMessage(chatID, userInfo)
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			sentMsg, err := bot.Send(msg)
			if err == nil {
				messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
			} else {
				log.Printf("Error sending Chat ID message: %v", err)
			}

		case "/menu", "📋 Menu":
			text := "🤖 *SMART AASA BOT* - Panduan Lengkap Kripto\n\n" +
				"📋 *Pilih menu di bawah:*"

			msg := tgbotapi.NewMessage(chatID, text)
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			msg.ReplyMarkup = mainMenuKeyboard
			sentMsg, err := bot.Send(msg)
			if err == nil {
				messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
			} else {
				log.Printf("Error sending /menu message: %v", err)
			}

		// Commands
		case "/claim":
			guide, ok := guides["worldcoin_registration_guide"]
			if ok {
				sendDetailedGuide(bot, chatID, guide)
			} else {
				log.Printf("Guide 'worldcoin_registration_guide' not found in JSON")
				msg := tgbotapi.NewMessage(chatID, "❌ Panduan claim tidak dijumpai.")
				bot.Send(msg)
			}

		case "/wallet", "/wallet hata":
			guide, ok := guides["hata_setup_guide"]
			if ok {
				sendDetailedGuide(bot, chatID, guide)
			} else {
				log.Printf("Guide 'hata_setup_guide' not found in JSON")
				msg := tgbotapi.NewMessage(chatID, "❌ Panduan wallet tidak dijumpai.")
				bot.Send(msg)
			}

		case "/cashout":
			guide, ok := guides["cashout_guide"]
			if ok {
				sendDetailedGuide(bot, chatID, guide)
			} else {
				log.Printf("Guide 'cashout_guide' not found in JSON")
				msg := tgbotapi.NewMessage(chatID, "❌ Panduan cashout tidak dijumpai.")
				bot.Send(msg)
			}

		case "/chatid":
			userInfo := fmt.Sprintf(
				"🆔 *Chat ID Anda:* `%d`\n\n"+
					"💡 *Untuk support/admin purposes*",
				chatID,
			)

			msg := tgbotapi.NewMessage(chatID, userInfo)
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			sentMsg, err := bot.Send(msg)
			if err == nil {
				messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
			} else {
				log.Printf("Error sending /chatid message: %v", err)
			}

		case "/reset", "🔄 Reset":
			// Padam semua mesej bot yang disimpan
			deletedCount := 0
			for _, id := range messageIDsToDelete[chatID] {
				_, err := bot.Request(tgbotapi.NewDeleteMessage(chatID, id))
				if err == nil {
					deletedCount++
				} else {
					// Log ralat jika perlu, tetapi jangan hentikan proses
					// log.Printf("Error deleting message %d: %v", id, err)
				}
			}
			// Kosongkan senarai untuk chat ini
			messageIDsToDelete[chatID] = nil

			// Hantar fresh start
			text := fmt.Sprintf("🔄 *Reset Berjaya!*\n%d mesej telah dipadam.\n\nTekan butang di bawah untuk mula semula.", deletedCount)
			msg := tgbotapi.NewMessage(chatID, text)
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			msg.ReplyMarkup = initialKeyboard
			sentMsg, err := bot.Send(msg)
			if err == nil {
				// Simpan ID mesej reset ini untuk padam seterusnya jika perlu
				messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
			} else {
				log.Printf("Error sending reset confirmation message: %v", err)
			}

		default:
			unknownText := "❌ Arahan tak dikenali. Sila guna butang atau command yang betul.\n\nKetik /menu untuk papar menu."
			msg := tgbotapi.NewMessage(chatID, unknownText)
			sentMsg, err := bot.Send(msg)
			if err == nil {
				messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
			} else {
				log.Printf("Error sending unknown command message: %v", err)
			}
		}
	}
}
