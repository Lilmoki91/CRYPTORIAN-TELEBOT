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

// --- GLOBAL (untuk reset mesej) ---
var messageIDsToDelete = make(map[int64][]int)

// --- HELPER FUNCTIONS ---

// Escape text untuk MarkdownV2
func escapeMarkdownV2(text string) string {
	re := regexp.MustCompile(`([_{}\[\]()~>#\+\-=|.!])`)
	return re.ReplaceAllString(text, `\$1`)
}

// Hantar panduan penuh dari markdown.json
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
			// Album multi-gambar
			var mediaGroup []interface{}
			for i, imgURL := range step.Images {
				photo := tgbotapi.NewInputMediaPhoto(tgbotapi.FileURL(imgURL))
				if i == 0 {
					photo.Caption = caption.String()
					photo.ParseMode = tgbotapi.ModeMarkdownV2
				}
				mediaGroup = append(mediaGroup, photo)
			}
			album := tgbotapi.NewMediaGroup(chatID, mediaGroup)

			// FIX: bot.Send(album) return single Message, bukan slice
			sentMsg, err := bot.Send(album)
			if err == nil {
				messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
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
	log.Printf("Bot started: @%s", bot.Self.UserName)

	// --- KEYBOARDS ---
	initialKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("▶️ Papar Menu Utama", "show_main_menu"),
		),
	)

	mainMenuKeyboard := tgbotapi.NewReplyKeyboard(
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
			tgbotapi.NewKeyboardButton("🔙 Kembali Menu Utama"),
			tgbotapi.NewKeyboardButton("🔄 Reset (Padam Mesej)"),
		),
	)

	// --- LOOP UPDATE ---
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		// Inline callback
		if update.CallbackQuery != nil {
			bot.Request(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
			if update.CallbackQuery.Data == "show_main_menu" {
				chatID := update.CallbackQuery.Message.Chat.ID
				text := "Sila guna *Butang Aksi Pantas* di bawah atau taip `/claim`, `/wallet`, `/cashout`."
				msg := tgbotapi.NewMessage(chatID, text)
				msg.ParseMode = tgbotapi.ModeMarkdownV2
				msg.ReplyMarkup = mainMenuKeyboard
				sentMsg, _ := bot.Send(msg)
				messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)

				// Padam butang asal
				deleteMsg := tgbotapi.NewDeleteMessage(chatID, update.CallbackQuery.Message.MessageID)
				bot.Request(deleteMsg)
			}
			continue
		}

		if update.Message == nil || update.Message.Text == "" {
			continue
		}

		chatID := update.Message.Chat.ID

		switch update.Message.Text {
		case "/start", "🔙 Kembali Menu Utama":
			msg := tgbotapi.NewMessage(chatID, "👋 Selamat Datang! Tekan butang di bawah untuk mula.")
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			msg.ReplyMarkup = initialKeyboard
			sentMsg, _ := bot.Send(msg)
			messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)

		case "🚀 Claim WorldCoin":
			text := "Daftar *WorldCoin*: [Klik sini](https://worldcoin.org/join/4RH0OTE)\n\nPanduan penuh: `/claim`"
			msg := tgbotapi.NewMessage(chatID, text)
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			sentMsg, _ := bot.Send(msg)
			messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)

		case "💰 Wallet HATA":
			text := "Daftar *HATA Wallet*: [Klik sini](https://hata.io/signup?ref=HDX8778)\n\nPanduan penuh: `/wallet`"
			msg := tgbotapi.NewMessage(chatID, text)
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			sentMsg, _ := bot.Send(msg)
			messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)

		case "💸 Cashout ke Bank":
			text := "Panduan cashout ada pada arahan: `/cashout`"
			msg := tgbotapi.NewMessage(chatID, text)
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			sentMsg, _ := bot.Send(msg)
			messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)

		case "📢 Channel":
			text := "Join channel kami: [Klik sini](https://t.me/cucikripto)"
			msg := tgbotapi.NewMessage(chatID, text)
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			sentMsg, _ := bot.Send(msg)
			messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)

		case "👨‍💻 Admin":
			text := "Hubungi admin: [Klik sini](https://t.me/johansetia)"
			msg := tgbotapi.NewMessage(chatID, text)
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			sentMsg, _ := bot.Send(msg)
			messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)

		// Commands
		case "/claim":
			sendDetailedGuide(bot, chatID, guides["worldcoin_registration_guide"])

		case "/wallet", "/wallet hata":
			sendDetailedGuide(bot, chatID, guides["hata_setup_guide"])

		case "/cashout":
			sendDetailedGuide(bot, chatID, guides["cashout_guide"])

		case "/reset mesej", "🔄 Reset (Padam Mesej)":
			// Padam semua mesej bot
			for _, id := range messageIDsToDelete[chatID] {
				bot.Request(tgbotapi.NewDeleteMessage(chatID, id))
			}
			messageIDsToDelete[chatID] = nil

			// Hantar fresh start
			msg := tgbotapi.NewMessage(chatID, "🔄 Reset! Tekan butang di bawah untuk mula semula.")
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			msg.ReplyMarkup = initialKeyboard
			sentMsg, _ := bot.Send(msg)
			messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)

		default:
			msg := tgbotapi.NewMessage(chatID, "❌ Arahan tak dikenali. Sila guna butang atau command yang betul.")
			sentMsg, _ := bot.Send(msg)
			messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
		}
	}
}
