// ===============
// --- MAIN.GO ---
// ===============

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

// --- HELPER FUNCTIONS ---

func escapeMarkdownV2(text string) string {
	re := regexp.MustCompile(`([_{}\[\]()~>#\+\-=|.!])`)
	return re.ReplaceAllString(text, `\$1`)
}

func sendDetailedGuide(bot *tgbotapi.BotAPI, chatID int64, guide Guide, messageIDs *map[int64][]int) {
	// Send main title
	titleMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("📖 *%s*", escapeMarkdownV2(guide.Title)))
	titleMsg.ParseMode = tgbotapi.ModeMarkdownV2
	if sentMsg, err := bot.Send(titleMsg); err == nil {
		(*messageIDs)[chatID] = append((*messageIDs)[chatID], sentMsg.MessageID)
	} else {
		log.Printf("Failed to send guide title to chat %d: %v", chatID, err)
	}

	// Send each step
	for _, step := range guide.Steps {
		var caption strings.Builder
		caption.WriteString(fmt.Sprintf("*%s*\n\n", escapeMarkdownV2(step.Title)))
		caption.WriteString(escapeMarkdownV2(step.Desc))

		if len(step.Images) == 0 {
			msg := tgbotapi.NewMessage(chatID, caption.String())
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			if sentMsg, err := bot.Send(msg); err == nil {
				(*messageIDs)[chatID] = append((*messageIDs)[chatID], sentMsg.MessageID)
			} else {
				log.Printf("Failed to send guide step (text) to chat %d: %v", chatID, err)
			}
		} else if len(step.Images) == 1 {
			photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(step.Images[0]))
			photo.Caption = caption.String()
			photo.ParseMode = tgbotapi.ModeMarkdownV2
			if sentMsg, err := bot.Send(photo); err == nil {
				(*messageIDs)[chatID] = append((*messageIDs)[chatID], sentMsg.MessageID)
			} else {
				log.Printf("Failed to send guide step (photo) to chat %d: %v", chatID, err)
			}
		} else {
			// Send as an album, splitting into chunks of 10 if necessary
			const chunkSize = 10
			totalImages := len(step.Images)
			for i := 0; i < totalImages; i += chunkSize {
				end := i + chunkSize
				if end > totalImages {
					end = totalImages
				}
				chunk := step.Images[i:end]
				var mediaGroup []interface{}
				for j, imgURL := range chunk {
					photo := tgbotapi.NewInputMediaPhoto(tgbotapi.FileURL(imgURL))
					if i == 0 && j == 0 {
						photo.Caption = caption.String()
						photo.ParseMode = tgbotapi.ModeMarkdownV2
					}
					mediaGroup = append(mediaGroup, photo)
				}
				album := tgbotapi.NewMediaGroup(chatID, mediaGroup)

				// FINAL FIX: Use bot.Request for albums as it returns an array of messages
				if resp, err := bot.Request(album); err == nil {
					var sentMessages []tgbotapi.Message
					if err := json.Unmarshal(resp.Result, &sentMessages); err == nil {
						for _, msg := range sentMessages {
							(*messageIDs)[chatID] = append((*messageIDs)[chatID], msg.MessageID)
						}
					}
				} else {
					log.Printf("Failed to send guide step (album chunk) to chat %d: %v", chatID, err)
				}
			}
		}
	}

	// Send important notes
	if len(guide.Important.Notes) > 0 {
		var notesBuilder strings.Builder
		notesBuilder.WriteString(fmt.Sprintf("*%s*\n\n", escapeMarkdownV2(guide.Important.Title)))
		for _, note := range guide.Important.Notes {
			notesBuilder.WriteString(fmt.Sprintf("• %s\n", escapeMarkdownV2(note)))
		}
		msg := tgbotapi.NewMessage(chatID, notesBuilder.String())
		msg.ParseMode = tgbotapi.ModeMarkdownV2
		if sentMsg, err := bot.Send(msg); err == nil {
			(*messageIDs)[chatID] = append((*messageIDs)[chatID], sentMsg.MessageID)
		} else {
			log.Printf("Failed to send guide notes to chat %d: %v", chatID, err)
		}
	}
}

func main() {
	// --- INITIALIZATION ---
	jsonData, err := os.ReadFile("markdown.json")
	if err != nil {
		log.Fatalf("Error reading markdown.json: %v", err)
	}
	var guides map[string]Guide
	if err := json.Unmarshal(jsonData, &guides); err != nil {
		log.Fatalf("JSON parsing error: %v", err)
	}

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatalf("TELEGRAM_BOT_TOKEN environment variable not set")
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatalf("Token error: %v", err)
	}
	log.Printf("Bot started: @%s", bot.Self.UserName)

	// --- KEYBOARDS DEFINITION ---
	initialKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("▶️ Papar Menu Utama", "show_main_menu"),
		),
	)
	mainMenuKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("💰 Claim WorldCoin"),
			tgbotapi.NewKeyboardButton("🛅 Wallet HATA"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🏧 Cashout ke Bank"),
			tgbotapi.NewKeyboardButton("📢 Channel"),
			tgbotapi.NewKeyboardButton("👨‍💻 Admin"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔙 Kembali Menu Utama"),
			tgbotapi.NewKeyboardButton("🔄 Reset (Padam Mesej)"),
		),
	)
	mainMenuKeyboard.ResizeKeyboard = true

	// --- BOT UPDATE LOOP ---
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)
	var messageIDsToDelete = make(map[int64][]int)
	for update := range updates {
		if update.CallbackQuery != nil {
			bot.Request(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
			if update.CallbackQuery.Data == "show_main_menu" {
				chatID := update.CallbackQuery.Message.Chat.ID
				text := "Sila tekan ⚡*_Butang Aksi Pantas_* di bawah untuk pautan 🔗link pantas, atau 🔡taip teks `/claim` `/wallet` `/cashout` untuk langkah panduan bergambar\\."
				msg := tgbotapi.NewMessage(chatID, text)
				msg.ParseMode = tgbotapi.ModeMarkdownV2
				msg.ReplyMarkup = mainMenuKeyboard
				if sentMsg, err := bot.Send(msg); err == nil {
					messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
				} else {
					log.Printf("Failed to send main menu message to chat %d: %v", chatID, err)
				}
				deleteMsg := tgbotapi.NewDeleteMessage(chatID, update.CallbackQuery.Message.MessageID)
				bot.Request(deleteMsg)
			}
			continue
		}
		if update.Message == nil || update.Message.Text == "" {
			continue
		}
		chatID := update.Message.Chat.ID
		messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], update.Message.MessageID)
		switch update.Message.Text {
		case "/start", "🔙 Kembali Menu Utama":
			msg := tgbotapi.NewMessage(chatID, "👋 Selamat Datang ke 🤖 *`CRYTORIAN-TELEBOT`*\\!\\. Tekan butang ▶️ di bawah untuk memaparkan menu\\.")
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			msg.ReplyMarkup = initialKeyboard
			if sentMsg, err := bot.Send(msg); err == nil {
				messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
			} else {
				log.Printf("Failed to send start message to chat %d: %v", chatID, err)
			}
		case "💰 Claim WorldCoin":
			text := "Untuk mendaftar *WorldCoin*, sila guna pautan di bawah:\n\n🔗 [Daftar WorldCoin Di Sini](https://worldcoin.org/join/4RH0OTE)\n\nUntuk panduan penuh bergambar, taip: `/claim`"
			msg := tgbotapi.NewMessage(chatID, text)
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			if sentMsg, err := bot.Send(msg); err == nil {
				messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
			} else {
				log.Printf("Failed to send 'Claim WorldCoin' message to chat %d: %v", chatID, err)
			}
		case "🛅 Wallet HATA":
			text := "Untuk mendaftar *HATA Wallet*, sila guna pautan di bawah:\n\n🔗 [Daftar HATA Di Sini](https://hata.io/signup?ref=HDX8778)\n\nUntuk panduan penuh bergambar, taip: `/wallet`"
			msg := tgbotapi.NewMessage(chatID, text)
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			if sentMsg, err := bot.Send(msg); err == nil {
				messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
			} else {
				log.Printf("Failed to send 'Wallet HATA' message to chat %d: %v", chatID, err)
			}
		case "🏧 Cashout ke Bank":
			text := "Panduan untuk menjual WorldCoin dan mengeluarkannya ke akaun bank anda boleh didapati melalui arahan di bawah\\.\n\ntaip: `/cashout`"
			msg := tgbotapi.NewMessage(chatID, text)
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			if sentMsg, err := bot.Send(msg); err == nil {
				messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
			} else {
				log.Printf("Failed to send 'Cashout' message to chat %d: %v", chatID, err)
			}
		case "📢 Channel":
			text := "Sertai saluran Telegram kami untuk info dan maklumat lanjutan\\.\n\n🔗 [Join Channel](https://t.me/cucikripto)"
			msg := tgbotapi.NewMessage(chatID, text)
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			if sentMsg, err := bot.Send(msg); err == nil {
				messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
			} else {
				log.Printf("Failed to send 'Channel' message to chat %d: %v", chatID, err)
			}
		case "👨‍💻 Admin":
			text := "🆘 *_Perlukan bantuan lanjut_*\\? 📞Hubungi admin secara terus ⌚ Waktu urusan *setiap hari*\\.\n\n🔗 [Hubungi Admin](https://t.me/johansetia)"
			msg := tgbotapi.NewMessage(chatID, text)
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			if sentMsg, err := bot.Send(msg); err == nil {
				messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
			} else {
				log.Printf("Failed to send 'Admin' message to chat %d: %v", chatID, err)
			}
		case "/claim":
			sendDetailedGuide(bot, chatID, guides["worldcoin_registration_guide"], &messageIDsToDelete)
		case "/wallet", "/wallet hata":
			sendDetailedGuide(bot, chatID, guides["hata_setup_guide"], &messageIDsToDelete)
		case "/cashout":
			sendDetailedGuide(bot, chatID, guides["cashout_guide"], &messageIDsToDelete)
		case "/reset mesej", "🔄 Reset (Padam Mesej)":
			for _, id := range messageIDsToDelete[chatID] {
				bot.Request(tgbotapi.NewDeleteMessage(chatID, id))
			}
			messageIDsToDelete[chatID] = nil
			msg := tgbotapi.NewMessage(chatID, "🔄 Sesi telah direset semula\\. Tekan ▶️ *_butang menu_* di bawah untuk mula sekali lagi\\.")
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			msg.ReplyMarkup = initialKeyboard
			if sentMsg, err := bot.Send(msg); err == nil {
				messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
			} else {
				log.Printf("Failed to send reset confirmation to chat %d: %v", chatID, err)
			}
		default:
			msg := tgbotapi.NewMessage(chatID, "❌ Arahan tidak dikenali pasti\\. Sila gunakan butang atau taip arahan yang sah: `/claim`, `/wallet`, `/cashout`\\.")
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			if sentMsg, err := bot.Send(msg); err == nil {
				messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
			} else {
				log.Printf("Failed to send default message to chat %d: %v", chatID, err)
			}
		}
	}
}

