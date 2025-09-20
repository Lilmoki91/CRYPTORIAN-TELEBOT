package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

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

// --- GLOBAL MESSAGE ID TRACKER ---
// Perekod ID mesej global untuk memastikan semua mesej boleh dipadam.
var messageIDsToDelete = make(map[int64][]int)

// --- HELPER FUNCTIONS ---

// Function to escape text for MarkdownV2
func escapeMarkdownV2(text string) string {
	re := regexp.MustCompile(`([_{}\[\]()~>#\+\-=|.!])`)
	return re.ReplaceAllString(text, `\$1`)
}

// FUNGSI BANTUAN BARU: Menghantar mesej dan merekod IDnya secara automatik
func sendAndTrack(bot *tgbotapi.BotAPI, chatID int64, config tgbotapi.Chattable) {
	if sentMessages, err := bot.Send(config); err == nil {
		// bot.Send boleh memulangkan satu mesej atau beberapa mesej (untuk album)
		// Kod ini mengendalikan kedua-dua kes
		switch v := sentMessages.(type) {
		case tgbotapi.Message:
			messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], v.MessageID)
		case []tgbotapi.Message:
			for _, msg := range v {
				messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], msg.MessageID)
			}
		}
	} else {
		log.Printf("Gagal menghantar atau menjejak mesej: %v", err)
	}
}

// FUNGSI DIKEMAS KINI: Kini menggunakan sendAndTrack
func sendDetailedGuide(bot *tgbotapi.BotAPI, chatID int64, guide Guide) {
	titleMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("📖 *%s*", escapeMarkdownV2(guide.Title)))
	titleMsg.ParseMode = tgbotapi.ModeMarkdownV2
	sendAndTrack(bot, chatID, titleMsg)

	for _, step := range guide.Steps {
		var caption strings.Builder
		caption.WriteString(fmt.Sprintf("*%s*\n\n", escapeMarkdownV2(step.Title)))
		caption.WriteString(escapeMarkdownV2(step.Desc))

		var config tgbotapi.Chattable
		if len(step.Images) == 0 {
			msg := tgbotapi.NewMessage(chatID, caption.String())
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			config = msg
		} else if len(step.Images) == 1 {
			photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(step.Images[0]))
			photo.Caption = caption.String()
			photo.ParseMode = tgbotapi.ModeMarkdownV2
			config = photo
		} else {
			var mediaGroup []interface{}
			for i, imgURL := range step.Images {
				photo := tgbotapi.NewInputMediaPhoto(tgbotapi.FileURL(imgURL))
				if i == 0 {
					photo.Caption = caption.String()
					photo.ParseMode = tgbotapi.ModeMarkdownV2
				}
				mediaGroup = append(mediaGroup, photo)
			}
			config = tgbotapi.NewMediaGroup(chatID, mediaGroup)
		}
		sendAndTrack(bot, chatID, config)
	}

	if len(guide.Important.Notes) > 0 {
		var notesBuilder strings.Builder
		notesBuilder.WriteString(fmt.Sprintf("*%s*\n\n", escapeMarkdownV2(guide.Important.Title)))
		for _, note := range guide.Important.Notes {
			notesBuilder.WriteString(fmt.Sprintf("• %s\n", escapeMarkdownV2(note)))
		}
		msg := tgbotapi.NewMessage(chatID, notesBuilder.String())
		msg.ParseMode = tgbotapi.ModeMarkdownV2
		sendAndTrack(bot, chatID, msg)
	}
}

func main() {
	// --- INITIALIZATION ---
	jsonData, err := os.ReadFile("markdown.json")
	if err != nil { log.Panic("Error reading markdown.json: ", err) }
	var guides map[string]Guide
	if err := json.Unmarshal(jsonData, &guides); err != nil { log.Panic("JSON parsing error: ", err) }

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" { log.Panic("TELEGRAM_BOT_TOKEN environment variable not set") }

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil { log.Panic("Token error: ", err) }
	log.Printf("Bot started: @%s", bot.Self.UserName)

	// --- KEYBOARDS DEFINITION ---
	initialKeyboard := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("▶️ Papar Menu Utama", "show_main_menu")))
	mainMenuKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("🚀 Claim WorldCoin"), tgbotapi.NewKeyboardButton("💰 Wallet HATA")),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("💸 Cashout ke Bank"), tgbotapi.NewKeyboardButton("📢 Channel"), tgbotapi.NewKeyboardButton("👨‍💻 Admin")),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("🔙 Kembali Menu Utama"), tgbotapi.NewKeyboardButton("🔄 Reset (Padam Mesej)")),
	)

	// --- BOT UPDATE LOOP ---
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.CallbackQuery != nil {
			bot.Request(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
			chatID := update.CallbackQuery.Message.Chat.ID
			if update.CallbackQuery.Data == "show_main_menu" {
				text := "Sila guna *Butang Aksi Pantas* di bawah untuk pautan pantas, atau guna arahan seperti `/claim` untuk panduan langkah demi langkah\\."
				msg := tgbotapi.NewMessage(chatID, text)
				msg.ParseMode = tgbotapi.ModeMarkdownV2
				msg.ReplyMarkup = mainMenuKeyboard
				sendAndTrack(bot, chatID, msg)
				bot.Request(tgbotapi.NewDeleteMessage(chatID, update.CallbackQuery.Message.MessageID))
			}
			continue
		}

		if update.Message == nil || update.Message.Text == "" { continue }
		
		chatID := update.Message.Chat.ID
		messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], update.Message.MessageID)

		switch update.Message.Text {
		case "/start", "🔙 Kembali Menu Utama":
			msg := tgbotapi.NewMessage(chatID, "👋 Selamat Datang\\! Tekan butang di bawah untuk memaparkan menu\\.")
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			msg.ReplyMarkup = initialKeyboard
			sendAndTrack(bot, chatID, msg)

		case "🚀 Claim WorldCoin":
			msg := tgbotapi.NewMessage(chatID, "Untuk mendaftar *WorldCoin*, sila guna pautan di bawah:\n\n🔗 [Daftar WorldCoin Di Sini](https://worldcoin.org/join/4RH0OTE)\n\nUntuk panduan penuh bergambar, guna arahan: `/claim`")
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			sendAndTrack(bot, chatID, msg)
			
		case "💰 Wallet HATA":
			msg := tgbotapi.NewMessage(chatID, "Untuk mendaftar *HATA Wallet*, sila guna pautan di bawah:\n\n🔗 [Daftar HATA Di Sini](https://hata.io/signup?ref=HDX8778)\n\nUntuk panduan penuh bergambar, guna arahan: `/wallet`")
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			sendAndTrack(bot, chatID, msg)

		case "💸 Cashout ke Bank":
			msg := tgbotapi.NewMessage(chatID, "Panduan untuk menjual WorldCoin dan mengeluarkannya ke akaun bank anda boleh didapati melalui arahan di bawah\\.\n\nTaip arahan: `/cashout`")
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			sendAndTrack(bot, chatID, msg)

		case "📢 Channel":
			msg := tgbotapi.NewMessage(chatID, "Sertai saluran Telegram kami untuk berita dan maklumat terkini\\.\n\n🔗 [Join Channel](https://t.me/cucikripto)")
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			sendAndTrack(bot, chatID, msg)
		
		case "👨‍💻 Admin":
			msg := tgbotapi.NewMessage(chatID, "Perlukan bantuan lanjut\\? Hubungi admin secara terus\\.\n\n🔗 [Hubungi Admin](https://t.me/johansetia)")
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			sendAndTrack(bot, chatID, msg)

		case "/claim":
			sendDetailedGuide(bot, chatID, guides["worldcoin_registration_guide"])
			
		case "/wallet", "/wallet hata":
			sendDetailedGuide(bot, chatID, guides["hata_setup_guide"])
			
		case "/cashout":
			sendDetailedGuide(bot, chatID, guides["cashout_guide"])

		case "/reset mesej", "🔄 Reset (Padam Mesej)":
			// Padam semua mesej yang telah direkodkan untuk chat ini
			if ids, ok := messageIDsToDelete[chatID]; ok && len(ids) > 0 {
				deleteConfig := tgbotapi.DeleteMessagesConfig{
					ChatID:     chatID,
					MessageIDs: ids,
				}
				bot.Request(deleteConfig)
				messageIDsToDelete[chatID] = nil // Kosongkan senarai
			}
			// Hantar mesej start yang baru (tanpa menjejaknya kerana sesi baru bermula)
			startMsg := tgbotapi.NewMessage(chatID, "🔄 Sesi telah direset semula\\. Tekan butang di bawah untuk mula sekali lagi\\.")
			startMsg.ParseMode = tgbotapi.ModeMarkdownV2
			startMsg.ReplyMarkup = initialKeyboard
			bot.Send(startMsg)

		default:
			msg := tgbotapi.NewMessage(chatID, "❌ Arahan tidak dikenali\\. Sila guna butang atau arahan yang sah\\.")
			sendAndTrack(bot, chatID, msg)
		}
	}
}
