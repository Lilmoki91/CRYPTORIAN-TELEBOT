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
var messageIDsToDelete = make(map[int64][]int)

// --- HELPER FUNCTIONS ---
func escapeMarkdownV2(text string) string {
	re := regexp.MustCompile(`([_{}\[\]()~>#\+\-=|.!])`)
	return re.ReplaceAllString(text, `\$1`)
}

// FUNGSI YANG TELAH DIPERBAIKI SEPENUHNYA
func sendAndTrack(bot *tgbotapi.BotAPI, chatID int64, config tgbotapi.Chattable) {
	// Kita periksa jenis 'config' (input), bukan output
	switch conf := config.(type) {
	case tgbotapi.MediaGroupConfig:
		// Jika ia album, bot.Send akan pulangkan beberapa mesej
		if sentMessages, err := bot.Send(conf); err == nil {
			for _, msg := range sentMessages {
				messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], msg.MessageID)
			}
		} else {
			log.Printf("Gagal menghantar album: %v", err)
		}
	default:
		// Untuk semua jenis lain (Teks, Gambar tunggal), ia akan pulangkan satu mesej
		if sentMsg, err := bot.Send(conf); err == nil {
			messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
		} else {
			log.Printf("Gagal menghantar mesej: %v", err)
		}
	}
}


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
	jsonData, err := os.ReadFile("markdown.json")
	if err != nil { log.Panic("Error reading markdown.json: ", err) }
	var guides map[string]Guide
	if err := json.Unmarshal(jsonData, &guides); err != nil { log.Panic("JSON parsing error: ", err) }
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" { log.Panic("TELEGRAM_BOT_TOKEN environment variable not set") }
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil { log.Panic("Token error ", err) }
	log.Printf("Bot started: @%s", bot.Self.UserName)
	initialKeyboard := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("▶️ Papar Menu Utama", "show_main_menu")))
	mainMenuKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("🚀 Claim WorldCoin"), tgbotapi.NewKeyboardButton("💰 Wallet HATA")),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("💸 Cashout ke Bank"), tgbotapi.NewKeyboardButton("📢 Channel"), tgbotapi.NewKeyboardButton("👨‍💻 Admin")),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("🔙 Kembali Menu Utama"), tgbotapi.NewKeyboardButton("🔄 Reset (Padam Mesej)")),
	)
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
		case "🚀 Claim WorldCoin", "💰 Wallet HATA", "💸 Cashout ke Bank", "📢 Channel", "👨‍💻 Admin":
			var text string
			switch update.Message.Text {
			case "🚀 Claim WorldCoin":
				text = "Untuk mendaftar *WorldCoin*, sila guna pautan di bawah:\n\n🔗 [Daftar WorldCoin Di Sini](https://worldcoin.org/join/4RH0OTE)\n\nUntuk panduan penuh bergambar, guna arahan: `/claim`"
			case "💰 Wallet HATA":
				text = "Untuk mendaftar *HATA Wallet*, sila guna pautan di bawah:\n\n🔗 [Daftar HATA Di Sini](https://hata.io/signup?ref=HDX8778)\n\nUntuk panduan penuh bergambar, guna arahan: `/wallet`"
			case "💸 Cashout ke Bank":
				text = "Panduan untuk menjual WorldCoin dan mengeluarkannya ke akaun bank anda boleh didapati melalui arahan di bawah\\.\n\nTaip arahan: `/cashout`"
			case "📢 Channel":
				text = "Sertai saluran Telegram kami untuk berita dan maklumat terkini\\.\n\n🔗 [Join Channel](https://t.me/cucikripto)"
			case "👨‍💻 Admin":
				text = "Perlukan bantuan lanjut\\? Hubungi admin secara terus\\.\n\n🔗 [Hubungi Admin](https://t.me/johansetia)"
			}
			msg := tgbotapi.NewMessage(chatID, text)
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			sendAndTrack(bot, chatID, msg)
		case "/claim":
			sendDetailedGuide(bot, chatID, guides["worldcoin_registration_guide"])
		case "/wallet", "/wallet hata":
			sendDetailedGuide(bot, chatID, guides["hata_setup_guide"])
		case "/cashout":
			sendDetailedGuide(bot, chatID, guides["cashout_guide"])
		case "/reset mesej", "🔄 Reset (Padam Mesej)":
			if ids, ok := messageIDsToDelete[chatID]; ok && len(ids) > 0 {
				for _, id := range ids {
					bot.Request(tgbotapi.NewDeleteMessage(chatID, id))
				}
				messageIDsToDelete[chatID] = nil
			}
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
