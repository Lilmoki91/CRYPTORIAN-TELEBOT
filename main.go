				
// ===================================
// --- main.go main menu and sub menu
// ===================================


package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// --- DEFINISI PAPAN KEKUNCI (KEYBOARD) ---

// 1. Menu Utama (di bawah papan kekunci)
var mainMenuReplyKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("📚 Panduan Kripto"),
		tgbotapi.NewKeyboardButton("🔗 Pautan & Bantuan"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("📊 Infografik"),
		tgbotapi.NewKeyboardButton("♻️ Reset Mesej"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("🔙 Kembali Menu Utama"), // Butang untuk panggil semula mesej alu-aluan
	),
)

// 2. Sub-Menu untuk Panduan (pada mesej)
var guidesInlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Claim Worldcoin", "get_guide_claim"),
		tgbotapi.NewInlineKeyboardButtonData("Wallet HATA", "get_guide_wallet"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Proses Cashout", "get_guide_cashout"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("« Tutup Menu Ini", "close_menu"),
	),
)

// 3. Sub-Menu untuk Pautan (pada mesej)
var linksInlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonURL("📢 Channel Telegram", "https://t.me/cucikripto"),
		tgbotapi.NewInlineKeyboardButtonURL("🆘 Hubungi Admin", "https://t.me/johansetia"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("« Tutup Menu Ini", "close_menu"),
	),
)

func main() {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN mesti ditetapkan")
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("Bot Hibrid UI dimulakan: @%s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	var messageIDsToDelete = make(map[int64][]int)

	for update := range updates {
		// --- PENGENDALIAN BUTANG INLINE (SUB-MENU) ---
		if update.CallbackQuery != nil {
			bot.Request(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
			chatID := update.CallbackQuery.Message.Chat.ID
			messageID := update.CallbackQuery.Message.MessageID

			switch update.CallbackQuery.Data {
			// Butang 'Tutup' atau 'Kembali' pada sub-menu akan memadam mesej sub-menu itu
			case "close_menu":
				deleteMsg := tgbotapi.NewDeleteMessage(chatID, messageID)
				bot.Request(deleteMsg)
			
			// Tindakan untuk butang sub-menu
			case "get_guide_claim":
				bot.Send(tgbotapi.NewMessage(chatID, "✅ Anda telah menekan butang 'Claim Worldcoin'. Fungsi panduan akan dijalankan di sini."))
			
			case "get_guide_wallet":
				bot.Send(tgbotapi.NewMessage(chatID, "✅ Anda telah menekan butang 'Wallet HATA'. Fungsi panduan akan dijalankan di sini."))

			case "get_guide_cashout":
				bot.Send(tgbotapi.NewMessage(chatID, "✅ Anda telah menekan butang 'Proses Cashout'. Fungsi panduan akan dijalankan di sini."))
			}
			continue
		}

		// --- PENGENDALIAN TEKS & BUTANG UTAMA (REPLY KEYBOARD) ---
		if update.Message != nil && update.Message.Text != "" {
			chatID := update.Message.Chat.ID
			messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], update.Message.MessageID)

			switch update.Message.Text {
			case "/start", "🔙 Kembali Menu Utama":
				text := "👋 Selamat Datang! Sila pilih satu pilihan dari menu utama di bawah."
				msg := tgbotapi.NewMessage(chatID, text)
				msg.ReplyMarkup = mainMenuReplyKeyboard
				if sentMsg, err := bot.Send(msg); err == nil {
					messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
				}

			case "📚 Panduan Kripto":
				text := "📚 *Panduan Kripto*\n\nPilih satu panduan dari sub-menu di bawah:"
				msg := tgbotapi.NewMessage(chatID, text)
				msg.ParseMode = tgbotapi.ModeMarkdown
				msg.ReplyMarkup = guidesInlineKeyboard
				if sentMsg, err := bot.Send(msg); err == nil {
					messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
				}

			case "🔗 Pautan & Bantuan":
				text := "🔗 *Pautan & Bantuan*\n\nPilih satu pautan dari sub-menu di bawah:"
				msg := tgbotapi.NewMessage(chatID, text)
				msg.ParseMode = tgbotapi.ModeMarkdown
				msg.ReplyMarkup = linksInlineKeyboard
				if sentMsg, err := bot.Send(msg); err == nil {
					messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
				}

			case "📊 Infografik":
				bot.Send(tgbotapi.NewMessage(chatID, "✅ Anda telah menekan butang 'Infografik'. Fungsi infografik akan dijalankan di sini."))
				
			case "♻️ Reset Mesej":
				for _, id := range messageIDsToDelete[chatID] {
					bot.Request(tgbotapi.NewDeleteMessage(chatID, id))
				}
				messageIDsToDelete[chatID] = nil
				
				startText := "🔄 Sesi telah direset. Sila pilih satu pilihan dari menu utama di bawah."
				msg := tgbotapi.NewMessage(chatID, startText)
				msg.ReplyMarkup = mainMenuReplyKeyboard
				if sentMsg, err := bot.Send(msg); err == nil {
					messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
				}
			
			default:
				// (Anda boleh letak mesej untuk arahan tidak sah di sini)
			}
		}
	}
}
		
