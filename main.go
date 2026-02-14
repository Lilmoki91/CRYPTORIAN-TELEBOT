package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// --- KONSTAN AUDIO ---
const WELCOME_JINGLE_URL = "https://raw.githubusercontent.com/Lilmoki91/CRYPTORIAN-TELEBOT/main/assets/Selamat_datang.mp3"

// --- SEMUA STRUCTS ---
type Guide struct {
	Title     string    `json:"title"`
	Steps     []Step    `json:"steps"`
	Important Important `json:"important"`
}

type Step struct {
	Title  string   `json:"title"`
	Desc   string   `json:"desc"`
	Images []string `json:"images"`
}

type Important struct {
	Title string   `json:"title"`
	Notes []string `json:"notes"`
}

type InfographicStep struct {
	Step    string   `json:"step"`
	Image   string   `json:"image"`
	Details []string `json:"details"`
	Arrow   string   `json:"arrow"`
}

type InfographicGuide struct {
	Title     string            `json:"title"`
	ImageMain string            `json:"image_main"`
	Steps     []InfographicStep `json:"steps"`
}

// --- CACHE & GLOBAL VARS ---
var (
	worldcoinGuide   Guide
	hataGuide        Guide
	cashoutGuide     Guide
	infographicGuide InfographicGuide
	guidesLoaded     = false
	loadGuidesMutex  sync.Mutex
)

// --- DEFINISI KEYBOARD ---
var mainMenuReplyKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("📚 Panduan Kripto"),
		tgbotapi.NewKeyboardButton("🔗 Pautan & 🆘 Bantuan"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("📊 Infografik"),
		tgbotapi.NewKeyboardButton("♻️ Reset Mesej"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("🔙 Kembali Menu Utama"),
	),
)

var guidesInlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🌏 Claim Worldcoin", "get_guide_claim"),
		tgbotapi.NewInlineKeyboardButtonData("🛄 Wallet HATA", "get_guide_wallet"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🏧 Proses Cashout", "get_guide_cashout"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("« Tutup Menu Ini", "close_menu"),
	),
)

var linksInlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonURL("🌏 Claim Worldcoin", "https://worldcoin.org/join/4RH0OTE"),
		tgbotapi.NewInlineKeyboardButtonURL("🛄 Wallet HATA", "https://hata.io/signup?ref=186300"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonURL("📢 Channel Telegram", "https://t.me/cucikripto"),
		tgbotapi.NewInlineKeyboardButtonURL("🆘 Hubungi Admin", "https://t.me/johansetia"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("« Tutup Menu Ini", "close_menu"),
	),
)

// --- FUNGSI BANTUAN ---

// escapeMarkdown Ringkas untuk ModeMarkdown biasa
func escapeMarkdown(text string) string {
	if text == "" {
		return ""
	}
	// Dalam Markdown biasa, kita hanya perlu berhati-hati dengan simbol yang tidak berpasangan.
	// Kita kekalkan teks seadanya supaya nampak bersih.
	return text
}

func loadGuides() error {
	loadGuidesMutex.Lock()
	defer loadGuidesMutex.Unlock()

	if guidesLoaded {
		return nil
	}

	jsonData, err := os.ReadFile("markdown.json")
	if err != nil {
		return fmt.Errorf("gagal membaca markdown.json: %v", err)
	}

	var rawGuides map[string]json.RawMessage
	if err := json.Unmarshal(jsonData, &rawGuides); err != nil {
		return fmt.Errorf("gagal memproses JSON: %v", err)
	}

	json.Unmarshal(rawGuides["worldcoin_registration_guide"], &worldcoinGuide)
	json.Unmarshal(rawGuides["hata_setup_guide"], &hataGuide)
	json.Unmarshal(rawGuides["cashout_guide"], &cashoutGuide)
	json.Unmarshal(rawGuides["infographic_guide"], &infographicGuide)

	guidesLoaded = true
	log.Println("✓ Semua panduan berjaya dimuatkan ke cache")
	return nil
}

func addMessageID(messageIDs *map[int64][]int, mu *sync.Mutex, chatID int64, messageID int) {
	if messageIDs == nil || mu == nil {
		return
	}
	mu.Lock()
	defer mu.Unlock()

	if _, exists := (*messageIDs)[chatID]; !exists {
		(*messageIDs)[chatID] = []int{}
	}
	(*messageIDs)[chatID] = append((*messageIDs)[chatID], messageID)
}

func sendDetailedGuide(bot *tgbotapi.BotAPI, chatID int64, guide Guide, messageIDs *map[int64][]int, mu *sync.Mutex) {
	titleMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("📖 *%s*", guide.Title))
	titleMsg.ParseMode = tgbotapi.ModeMarkdown
	if sentMsg, err := bot.Send(titleMsg); err == nil {
		addMessageID(messageIDs, mu, chatID, sentMsg.MessageID)
	}

	for _, step := range guide.Steps {
		var caption strings.Builder
		caption.WriteString(fmt.Sprintf("*%s*\n\n", step.Title))
		caption.WriteString(step.Desc)

		if len(step.Images) == 0 {
			msg := tgbotapi.NewMessage(chatID, caption.String())
			msg.ParseMode = tgbotapi.ModeMarkdown
			if sentMsg, err := bot.Send(msg); err == nil {
				addMessageID(messageIDs, mu, chatID, sentMsg.MessageID)
			}
		} else if len(step.Images) == 1 {
			photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(step.Images[0]))
			photo.Caption = caption.String()
			photo.ParseMode = tgbotapi.ModeMarkdown
			if sentMsg, err := bot.Send(photo); err == nil {
				addMessageID(messageIDs, mu, chatID, sentMsg.MessageID)
			}
		} else {
			mediaGroup := []interface{}{}
			for i, imgURL := range step.Images {
				photo := tgbotapi.NewInputMediaPhoto(tgbotapi.FileURL(imgURL))
				if i == 0 {
					photo.Caption = caption.String()
					photo.ParseMode = tgbotapi.ModeMarkdown
				}
				mediaGroup = append(mediaGroup, photo)
			}
			if len(mediaGroup) > 0 {
				album := tgbotapi.NewMediaGroup(chatID, mediaGroup)
				if sentMessages, err := bot.SendMediaGroup(album); err == nil {
					for _, msg := range sentMessages {
						addMessageID(messageIDs, mu, chatID, msg.MessageID)
					}
				}
			}
		}
	}

	if len(guide.Important.Notes) > 0 {
		var notesBuilder strings.Builder
		notesBuilder.WriteString(fmt.Sprintf("\n*%s*\n", guide.Important.Title))
		for _, note := range guide.Important.Notes {
			notesBuilder.WriteString(fmt.Sprintf("- %s\n", note))
		}
		msg := tgbotapi.NewMessage(chatID, notesBuilder.String())
		msg.ParseMode = tgbotapi.ModeMarkdown

		if sentMsg, err := bot.Send(msg); err == nil {
			addMessageID(messageIDs, mu, chatID, sentMsg.MessageID)
		}
	}
}

func sendInfographicGuide(bot *tgbotapi.BotAPI, chatID int64, guide InfographicGuide, messageIDs *map[int64][]int, mu *sync.Mutex) {
	titleMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("*%s*", guide.Title))
	titleMsg.ParseMode = tgbotapi.ModeMarkdown
	if sentMsg, err := bot.Send(titleMsg); err == nil {
		addMessageID(messageIDs, mu, chatID, sentMsg.MessageID)
	}

	if guide.ImageMain != "" {
		photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(guide.ImageMain))
		if sentMsg, err := bot.Send(photo); err == nil {
			addMessageID(messageIDs, mu, chatID, sentMsg.MessageID)
		}
	}

	for _, step := range guide.Steps {
		var caption strings.Builder
		caption.WriteString(fmt.Sprintf("*%s*\n\n", step.Step))
		for _, detail := range step.Details {
			caption.WriteString(fmt.Sprintf("- %s\n", detail))
		}
		if step.Arrow != "" {
			caption.WriteString(fmt.Sprintf("\n%s\n", step.Arrow))
		}
		photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(step.Image))
		photo.Caption = caption.String()
		photo.ParseMode = tgbotapi.ModeMarkdown
		if sentMsg, err := bot.Send(photo); err == nil {
			addMessageID(messageIDs, mu, chatID, sentMsg.MessageID)
		}
	}
}

// --- FUNGSI UTAMA (MAIN) ---
func main() {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("❌ TELEGRAM_BOT_TOKEN mesti ditetapkan")
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panicf("❌ Gagal inisialisasi bot: %v", err)
	}
	log.Printf("✅ Bot Cryptorian dimulakan: @%s", bot.Self.UserName)

	if err := loadGuides(); err != nil {
		log.Fatalf("❌ %v", err)
	}

	go func() {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8000"
		}
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "✅ Cryptorian Bot is Running Live!")
		})
		log.Printf("🚀 Health Check Server bermula di port: %s", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Printf("⚠️ Gagal memulakan HTTP server: %v", err)
		}
	}()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	var (
		messageIDsToDelete = make(map[int64][]int)
		mu                 sync.Mutex
	)

	for update := range updates {
		var userID int64
		var chatID int64
		var username string

		if update.Message != nil {
			userID = update.Message.From.ID
			chatID = update.Message.Chat.ID
			username = update.Message.From.UserName
		} else if update.CallbackQuery != nil {
			userID = update.CallbackQuery.From.ID
			chatID = update.CallbackQuery.Message.Chat.ID
			username = update.CallbackQuery.From.UserName
		} else {
			continue
		}

		if IsBanned(userID) {
			continue
		}

		if update.CallbackQuery != nil {
			callback := update.CallbackQuery

			if callback.Data == "setuju_tnc" {
				err := SaveAgreementToGithub(userID, username)
				responseText := "✅ Persetujuan direkodkan! Sila taip /start untuk mula."
				if err != nil {
					log.Printf("Ralat Github: %v", err)
					responseText = "❌ Ralat teknikal, sila cuba lagi."
				}
				bot.Send(tgbotapi.NewEditMessageText(chatID, callback.Message.MessageID, responseText))
				bot.Request(tgbotapi.NewCallback(callback.ID, ""))
				continue
			}

			if callback.Data == "tolak_tnc" {
				pesanKeluar := "🚫 *AKSES DITOLAK*\n\nAnda tidak bersetuju dengan Terma & Syarat kami. Akses anda kepada sistem Cryptorian ditutup serta-merta.\n\nSila padam (delete) bot ini."
				editMsg := tgbotapi.NewEditMessageText(chatID, callback.Message.MessageID, pesanKeluar)
				editMsg.ParseMode = tgbotapi.ModeMarkdown
				bot.Send(editMsg)
				bot.Request(tgbotapi.NewCallback(callback.ID, "Akses Ditolak"))
				continue
			}

			if HasAgreed(userID) || IsAdmin(userID) {
				switch callback.Data {
				case "close_menu":
					bot.Request(tgbotapi.NewDeleteMessage(chatID, callback.Message.MessageID))
				case "get_guide_claim":
					sendDetailedGuide(bot, chatID, worldcoinGuide, &messageIDsToDelete, &mu)
				case "get_guide_wallet":
					sendDetailedGuide(bot, chatID, hataGuide, &messageIDsToDelete, &mu)
				case "get_guide_cashout":
					sendDetailedGuide(bot, chatID, cashoutGuide, &messageIDsToDelete, &mu)
				}
			}
			bot.Request(tgbotapi.NewCallback(callback.ID, ""))
			continue
		}

		if update.Message == nil {
			continue
		}
		addMessageID(&messageIDsToDelete, &mu, chatID, update.Message.MessageID)

		if update.Message.Command() == "ban" && IsAdmin(userID) {
			args := update.Message.CommandArguments()
			if args != "" {
				targetID, _ := strconv.ParseInt(args, 10, 64)
				err := BanUser(targetID, "Sekatan Admin Mr JOHAN")
				if err == nil {
					bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("🚫 User %d telah disekat!", targetID)))
				} else {
					bot.Send(tgbotapi.NewMessage(chatID, "❌ Gagal menyekat user."))
				}
			} else {
				bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Format salah: /ban [ID]"))
			}
			continue
		}

		isAllowed := IsAdmin(userID) || HasAgreed(userID)

		if !isAllowed && update.Message.Command() != "start" {
			msg := tgbotapi.NewMessage(chatID, "⚠️ Akses dihadkan. Sila taip /start dan setuju Terma & Syarat dahulu.")
			sentMsg, _ := bot.Send(msg)
			addMessageID(&messageIDsToDelete, &mu, chatID, sentMsg.MessageID)
			continue
		}

		switch update.Message.Text {
		case "/start", "🔙 Kembali Menu Utama":
			if isAllowed {
				audio := tgbotapi.NewAudio(chatID, tgbotapi.FileURL(WELCOME_JINGLE_URL))
				audio.Caption = "🎶 Selamat datang ke Cryptorian!"
				audio.ParseMode = tgbotapi.ModeMarkdown
				sentAudio, _ := bot.Send(audio)
				addMessageID(&messageIDsToDelete, &mu, chatID, sentAudio.MessageID)

				text := "👋 Selamat Datang ke 🤖 Cryptorian-Telebot!\n\nSila pilih satu pilihan dari menu utama di bawah."
				msg := tgbotapi.NewMessage(chatID, text)
				msg.ReplyMarkup = mainMenuReplyKeyboard
				msg.ParseMode = tgbotapi.ModeMarkdown
				sentMsg, _ := bot.Send(msg)
				addMessageID(&messageIDsToDelete, &mu, chatID, sentMsg.MessageID)
			} else {
				txt, err := BuildTermsUI()
				if err != nil {
					bot.Send(tgbotapi.NewMessage(chatID, "❌ Ralat memuatkan terma."))
					continue
				}
				msg := tgbotapi.NewMessage(chatID, txt)
				msg.ParseMode = tgbotapi.ModeMarkdown
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Setuju ✅", "setuju_tnc"),
						tgbotapi.NewInlineKeyboardButtonData("Tidak Setuju ❌", "tolak_tnc"),
					),
				)
				sentMsg, _ := bot.Send(msg)
				addMessageID(&messageIDsToDelete, &mu, chatID, sentMsg.MessageID)
			}

		case "📚 Panduan Kripto":
			if isAllowed {
				text := "*📚 Panduan Kripto*\n\nPilih satu panduan dari sub-menu di bawah:"
				msg := tgbotapi.NewMessage(chatID, text)
				msg.ParseMode = tgbotapi.ModeMarkdown
				msg.ReplyMarkup = guidesInlineKeyboard
				sentMsg, _ := bot.Send(msg)
				addMessageID(&messageIDsToDelete, &mu, chatID, sentMsg.MessageID)
			}

		case "🔗 Pautan & 🆘 Bantuan":
			if isAllowed {
				text := "*🔗 Pautan & 🆘 Bantuan*\n\nPilih pautan rasmi kami:"
				msg := tgbotapi.NewMessage(chatID, text)
				msg.ParseMode = tgbotapi.ModeMarkdown
				msg.ReplyMarkup = linksInlineKeyboard
				sentMsg, _ := bot.Send(msg)
				addMessageID(&messageIDsToDelete, &mu, chatID, sentMsg.MessageID)
			}

		case "📊 Infografik":
			if isAllowed {
				sendInfographicGuide(bot, chatID, infographicGuide, &messageIDsToDelete, &mu)
			}

		case "♻️ Reset Mesej":
			if isAllowed {
				mu.Lock()
				if ids, exists := messageIDsToDelete[chatID]; exists {
					for _, id := range ids {
						bot.Request(tgbotapi.NewDeleteMessage(chatID, id))
					}
					delete(messageIDsToDelete, chatID)
				}
				mu.Unlock()

				msg := tgbotapi.NewMessage(chatID, "🔄 *Sesi Direset*\n\nSila pilih menu:")
				msg.ParseMode = tgbotapi.ModeMarkdown
				msg.ReplyMarkup = mainMenuReplyKeyboard
				sentMsg, _ := bot.Send(msg)
				addMessageID(&messageIDsToDelete, &mu, chatID, sentMsg.MessageID)
			}
		}
	}
}
