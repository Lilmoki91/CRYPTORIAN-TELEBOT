package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// --- KONSTAN AUDIO ---

// Pautan telah dikemas kini kepada URL RAW GitHub agar Telegram boleh mengakses fail audio secara langsung.
const WELCOME_JINGLE_URL = "https://raw.githubusercontent.com/Lilmoki91/CRYPTORIAN-TELEBOT/main/assets/Selamat_datang.mp3"

// Had saiz imej yang dibenarkan untuk dimuat turun (10 MB)
const maxImageSize = 10 << 20 // 10 MB

// --- SEMUA STRUCTS YANG DIPERLUKAN ---
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

// --- DEFINISI PAPAN KEKUNCI (KEYBOARD) ---
var mainMenuReplyKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("📚 Panduan Kripto"), tgbotapi.NewKeyboardButton("🔗 Pautan & 🆘 Bantuan")),
	tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("📊 Infografik"), tgbotapi.NewKeyboardButton("♻️ Reset Mesej")),
	tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("🔙 Kembali Menu Utama")),
)
var guidesInlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🌏 Claim Worldcoin", "get_guide_claim"), tgbotapi.NewInlineKeyboardButtonData("🛄 Wallet HATA", "get_guide_wallet")),
	tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🏧 Proses Cashout", "get_guide_cashout")),
	tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("« Tutup Menu Ini", "close_menu")),
)
var linksInlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
	tgbotapi.NewInlineKeyboardButtonURL("🌏 Claim Worldcoin", "https://worldcoin.org/join/4RH0OTE"),
	tgbotapi.NewInlineKeyboardButtonURL("🛄 Wallet HATA", "https://hata.io/signup?ref=HDX8778"),
	),
	tgbotapi.NewInlineKeyboardRow(
	tgbotapi.NewInlineKeyboardButtonURL("📢 Channel Telegram", "https://t.me/cucikripto"),
	tgbotapi.NewInlineKeyboardButtonURL("🆘 Hubungi Admin", "https://t.me/johansetia"),
	),
	tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("« Tutup Menu Ini", "close_menu")),
)
// --- FUNGSI-FUNGSI BANTUAN ---

// escapeMarkdownV2 mengelak aksara khas MarkdownV2
func escapeMarkdownV2(text string) string {
	// Aksara khas: _, *, [, ], (, ), ~, `, >, #, +, -, =, |, {, }, ., !
	re := regexp.MustCompile(`([_*
\[\]()~>#+\-=|{}.!])`)
	// Tanda backslash perlu di-escape dua kali dalam rentetan Go
	return re.ReplaceAllString(text, `\\$1`)
}
// fetchImage memuat turun imej dari URL dan mengembalikan tgbotapi.FileBytes untuk upload.
// Ia memeriksa status, Content-Type mesti bermula dengan "image/", dan mengehadkan saiz.
func fetchImage(url string) (tgbotapi.FileBytes, error) {
	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return tgbotapi.FileBytes{}, fmt.Errorf("gagal buat request: %w", err)
	}
	// Tetapkan User-Agent ringkas supaya sesetengah server tak reject permintaan
	req.Header.Set("User-Agent", "CryptorianTelebot/1.0 (+https://github.com/Lilmoki91/CRYPTORIAN-TELEBOT)")

	resp, err := client.Do(req)
	if err != nil {
		return tgbotapi.FileBytes{}, fmt.Errorf("gagal fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return tgbotapi.FileBytes{}, fmt.Errorf("status bukan 200: %d", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if ct == "" || !strings.HasPrefix(ct, "image/") {
		return tgbotapi.FileBytes{}, fmt.Errorf("bukan content-type imej: %s", ct)
	}

	// Hadkan pembacaan kepada maxImageSize+1 supaya kita boleh detect oversize
	limitedReader := io.LimitReader(resp.Body, maxImageSize+1)
	buf := &bytes.Buffer{}
	n, err := io.Copy(buf, limitedReader)
	if err != nil {
		return tgbotapi.FileBytes{}, fmt.Errorf("gagal baca badan response: %w", err)
	}
	if n > int64(maxImageSize) {
		return tgbotapi.FileBytes{}, fmt.Errorf("imej melebihi saiz maksima (%d bytes)", maxImageSize)
	}

	// Nama fail mudah; Telegram tidak bergantung sepenuhnya pada nama ini untuk jenis
	name := "image"
	// cuba derive extension dari content-type (pilihan, bukan kritikal)
	if strings.HasPrefix(ct, "image/") {
		ext := strings.TrimPrefix(ct, "image/")
		// beberapa content-type tenggang seperti jpeg -> jpg
		if ext == "jpeg" {
			ext = "jpg"
		}
		name = "image." + ext
	}

	return tgbotapi.FileBytes{Name: name, Bytes: buf.Bytes()}, nil
}
// sendDetailedGuide menghantar panduan berstruktur langkah demi langkah
func sendDetailedGuide(bot *tgbotapi.BotAPI, chatID int64, guide Guide, messageIDs *map[int64][]int) {
	// Guna escapeMarkdownV2 untuk memastikan Title dilayan sebagai MarkdownV2
	titleMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("📖 *%s*", escapeMarkdownV2(guide.Title)))
	titleMsg.ParseMode = tgbotapi.ModeMarkdownV2
	if sentMsg, err := bot.Send(titleMsg); err == nil {
		(*messageIDs)[chatID] = append((*messageIDs)[chatID], sentMsg.MessageID)
	}

	for _, step := range guide.Steps {
		var caption strings.Builder
		// Escape Title dan Desc untuk MarkdownV2
		caption.WriteString(fmt.Sprintf("*%s*\n\n", escapeMarkdownV2(step.Title)))
		caption.WriteString(escapeMarkdownV2(step.Desc))

		if len(step.Images) == 0 {
			msg := tgbotapi.NewMessage(chatID, caption.String())
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			if sentMsg, err := bot.Send(msg); err == nil {
				(*messageIDs)[chatID] = append((*messageIDs)[chatID], sentMsg.MessageID)
			}
		} else if len(step.Images) == 1 {
			// Muat turun imej dulu, kemudian upload
			fileBytes, err := fetchImage(step.Images[0])
			if err != nil {
				log.Printf("Gagal fetch imej tunggal %s: %v", step.Images[0], err)
				// Jika fetch gagal, hantar mesej teks sebagai fallback
				msg := tgbotapi.NewMessage(chatID, caption.String())
				msg.ParseMode = tgbotapi.ModeMarkdownV2
				if sentMsg, err := bot.Send(msg); err == nil {
					(*messageIDs)[chatID] = append((*messageIDs)[chatID], sentMsg.MessageID)
				}
				continue
			}
			photo := tgbotapi.NewPhoto(chatID, fileBytes)
			photo.Caption = caption.String()
			photo.ParseMode = tgbotapi.ModeMarkdownV2
			if sentMsg, err := bot.Send(photo); err == nil {
				(*messageIDs)[chatID] = append((*messageIDs)[chatID], sentMsg.MessageID)
			} else {
				log.Printf("Gagal menghantar foto tunggal: %v", err)
			}
		} else {
			// Menggunakan MediaGroup untuk album foto, muat turun semua imej dahulu
			mediaGroup := []interface{}{}
			for i, imgURL := range step.Images {
				fileBytes, err := fetchImage(imgURL)
				if err != nil {
					log.Printf("Gagal fetch image %s: %v", imgURL, err)
					continue // skip imej yang gagal
				}
				photo := tgbotapi.NewInputMediaPhoto(fileBytes)
				if i == 0 {
					photo.Caption = caption.String()
					photo.ParseMode = tgbotapi.ModeMarkdownV2
				}
				mediaGroup = append(mediaGroup, photo)
			}
			if len(mediaGroup) == 0 {
				// Tiada imej berjaya dimuat turun, hantar teks sahaja
				msg := tgbotapi.NewMessage(chatID, caption.String())
				msg.ParseMode = tgbotapi.ModeMarkdownV2
				if sentMsg, err := bot.Send(msg); err == nil {
					(*messageIDs)[chatID] = append((*messageIDs)[chatID], sentMsg.MessageID)
				}
			} else {
				album := tgbotapi.NewMediaGroup(chatID, mediaGroup)
				if resp, err := bot.Request(album); err == nil {
					// Memproses respons untuk mendapatkan MessageID yang dihantar
					var sentMessages []tgbotapi.Message
					if err := json.Unmarshal(resp.Result, &sentMessages); err == nil {
						for _, msg := range sentMessages {
							(*messageIDs)[chatID] = append((*messageIDs)[chatID], msg.MessageID)
						}
					} else {
						log.Printf("Gagal unmarshal respons MediaGroup: %v", err)
					}
				} else {
					log.Printf("Gagal menghantar MediaGroup: %v", err)
				}
			}
		}
	}

	if len(guide.Important.Notes) > 0 {
		var notesBuilder strings.Builder
		// Escape Title dan Notes untuk MarkdownV2
		notesBuilder.WriteString(fmt.Sprintf("\n*%s*\n", escapeMarkdownV2(guide.Important.Title)))
		for _, note := range guide.Important.Notes {
			// Menggunakan \\n\- untuk bullet point dalam MarkdownV2
			notesBuilder.WriteString(fmt.Sprintf("\- %s\n", escapeMarkdownV2(note)))
		}
		msg := tgbotapi.NewMessage(chatID, notesBuilder.String())
		msg.ParseMode = tgbotapi.ModeMarkdownV2
		if sentMsg, err := bot.Send(msg); err == nil {
			(*messageIDs)[chatID] = append((*messageIDs)[chatID], sentMsg.MessageID)
		}
	}
}
// sendInfographicGuide menghantar panduan dalam bentuk infografik
func sendInfographicGuide(bot *tgbotapi.BotAPI, chatID int64, guide InfographicGuide, messageIDs *map[int64][]int) {
	// Menggunakan MarkdownV2 dan escape Title
	titleMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("*%s*", escapeMarkdownV2(guide.Title)))
	titleMsg.ParseMode = tgbotapi.ModeMarkdownV2
	if sentMsg, err := bot.Send(titleMsg); err == nil {
		(*messageIDs)[chatID] = append((*messageIDs)[chatID], sentMsg.MessageID)
	}

	if guide.ImageMain != "" {
		if fileBytes, err := fetchImage(guide.ImageMain); err == nil {
			photo := tgbotapi.NewPhoto(chatID, fileBytes)
			if sentMsg, err := bot.Send(photo); err == nil {
				(*messageIDs)[chatID] = append((*messageIDs)[chatID], sentMsg.MessageID)
			}
		} else {
			log.Printf("Gagal fetch ImageMain %s: %v", guide.ImageMain, err)
		}
	}

	for _, step := range guide.Steps {
		var caption strings.Builder
		// Escape Step title
		caption.WriteString(fmt.Sprintf("*%s*\n\n", escapeMarkdownV2(step.Step)))
		for _, detail := range step.Details {
			// Escape details dan guna \\n\- untuk bullet
			caption.WriteString(fmt.Sprintf("\- %s\n", escapeMarkdownV2(detail)))
		}
		if step.Arrow != "" {
			// Escape Arrow
			caption.WriteString(fmt.Sprintf("\n%s\n", escapeMarkdownV2(step.Arrow)))
		}

		if fileBytes, err := fetchImage(step.Image); err == nil {
			photo := tgbotapi.NewPhoto(chatID, fileBytes)
			photo.Caption = caption.String()
			photo.ParseMode = tgbotapi.ModeMarkdownV2
			if sentMsg, err := bot.Send(photo); err == nil {
				(*messageIDs)[chatID] = append((*messageIDs)[chatID], sentMsg.MessageID)
			} else {
				log.Printf("Gagal menghantar foto infografik: %v", err)
			}
		} else {
			log.Printf("Gagal fetch infographic step image %s: %v", step.Image, err)
			// fallback: hantar teks sahaja
			msg := tgbotapi.NewMessage(chatID, caption.String())
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			if sentMsg, err := bot.Send(msg); err == nil {
				(*messageIDs)[chatID] = append((*messageIDs)[chatID], sentMsg.MessageID)
			}
		}
	}
}

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

	jsonData, err := os.ReadFile("markdown.json")
	if err != nil {
		log.Fatalf("Gagal membaca markdown.json: %v", err)
	}

	var guides map[string]json.RawMessage
	if err := json.Unmarshal(jsonData, &guides); err != nil {
		log.Fatalf("Gagal memproses JSON: %v", err)
	}
	log.Println("Panduan berjaya dimuatkan.")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	var messageIDsToDelete = make(map[int64][]int)

	for update := range updates {
		if update.CallbackQuery != nil {
			// Akui callback query untuk menghilangkan status loading
			bot.Request(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
			chatID := update.CallbackQuery.Message.Chat.ID
			messageID := update.CallbackQuery.Message.MessageID

			// Hapus mesej butang yang baru ditekan
			bot.Request(tgbotapi.NewDeleteMessage(chatID, messageID))

			switch update.CallbackQuery.Data {
			case "close_menu":
				continue
			case "get_guide_claim":
				var guideData Guide
				if err := json.Unmarshal(guides["worldcoin_registration_guide"], &guideData); err == nil {
					sendDetailedGuide(bot, chatID, guideData, &messageIDsToDelete)
				} else {
					log.Printf("Ralat unmarshaling worldcoin_registration_guide: %v", err)
					bot.Send(tgbotapi.NewMessage(chatID, "🚫 Gagal memuatkan panduan Worldcoin\. Sila cuba lagi\."))
				}
			case "get_guide_wallet":
				var guideData Guide
				if err := json.Unmarshal(guides["hata_setup_guide"], &guideData); err == nil {
					sendDetailedGuide(bot, chatID, guideData, &messageIDsToDelete)
				} else {
					log.Printf("Ralat unmarshaling hata_setup_guide: %v", err)
					bot.Send(tgbotapi.NewMessage(chatID, "🚫 Gagal memuatkan panduan Wallet HATA\. Sila cuba lagi\."))
				}
			case "get_guide_cashout":
				var guideData Guide
				if err := json.Unmarshal(guides["cashout_guide"], &guideData); err == nil {
					sendDetailedGuide(bot, chatID, guideData, &messageIDsToDelete)
				} else {
					log.Printf("Ralat unmarshaling cashout_guide: %v", err)
					bot.Send(tgbotapi.NewMessage(chatID, "🚫 Gagal memuatkan panduan Cashout\. Sila cuba lagi\."))
				}
			}
			continue
		}

		if update.Message != nil && update.Message.Text != "" {
			chatID := update.Message.Chat.ID
			// Tambah mesej masuk ke senarai untuk pembersihan
			messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], update.Message.MessageID)

			switch update.Message.Text {
			case "/start", "🔙 Kembali Menu Utama":
				// --- LAGU SELAMAT DATANG (WELCOME JINGLE) ---
				audio := tgbotapi.NewAudio(chatID, tgbotapi.FileURL(WELCOME_JINGLE_URL))
				audio.Caption = "🎶 Selamat datang ke Cryptorian\!"
				audio.ParseMode = tgbotapi.ModeMarkdownV2
				// Tetapkan DisableNotification agar pengguna tidak menerima notifikasi kedua
				audio.DisableNotification = true

				if sentAudio, err := bot.Send(audio); err == nil {
					// Tambah ID mesej audio untuk dipadam kemudian
					messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentAudio.MessageID)
				} else {
					log.Printf("Gagal menghantar audio: %v", err)
					// Hantar mesej ralat ringkas tanpa menghalang mesej seterusnya
					errMsg := "⚠️ Gagal memainkan jingle\. Pastikan URL audio adalah sah dan boleh diakses awam\."
					errorMsg := tgbotapi.NewMessage(chatID, errMsg)
					errorMsg.ParseMode = tgbotapi.ModeMarkdownV2
					if sentError, err := bot.Send(errorMsg); err == nil {
						messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentError.MessageID)
					}
				}

				// --- MESEJ MENU UTAMA ---
				text := "👋 Selamat Datang ke 🤖 Cryptorian\-Telebot\!\, Sila pilih satu pilihan dari menu utama di bawah\." 
				msg := tgbotapi.NewMessage(chatID, text)
				msg.ReplyMarkup = mainMenuReplyKeyboard
				msg.ParseMode = tgbotapi.ModeMarkdownV2
				if sentMsg, err := bot.Send(msg); err == nil {
					messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
				}
			case "📚 Panduan Kripto":
				// Manual escape for V2 compliance on static text
				text := "*📚 Panduan Kripto*\n\nPilih satu panduan dari sub\-menu di bawah:"
				msg := tgbotapi.NewMessage(chatID, text)
				msg.ParseMode = tgbotapi.ModeMarkdownV2 // Menggunakan MarkdownV2
				msg.ReplyMarkup = guidesInlineKeyboard
				if sentMsg, err := bot.Send(msg); err == nil {
					messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
				}
			case "🔗 Pautan & 🆘 Bantuan":
				// Manual escape for V2 compliance on static text
				text := "*🔗 Pautan & 🆘 Bantuan*\n\nPilih satu pautan dari sub\-menu di bawah:"
				msg := tgbotapi.NewMessage(chatID, text)
				msg.ParseMode = tgbotapi.ModeMarkdownV2 // Menggunakan MarkdownV2
				msg.ReplyMarkup = linksInlineKeyboard
				if sentMsg, err := bot.Send(msg); err == nil {
					messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
				}
			case "📊 Infografik":
				var infographicData InfographicGuide
				if err := json.Unmarshal(guides["infographic_guide"], &infographicData); err == nil {
					sendInfographicGuide(bot, chatID, infographicData, &messageIDsToDelete)
				} else {
					log.Printf("Ralat unmarshaling infographic_guide: %v", err)
					bot.Send(tgbotapi.NewMessage(chatID, "🚫 Gagal memuatkan data infografik\. Sila cuba lagi\."))
				}
			case "♻️ Reset Mesej":
				for _, id := range messageIDsToDelete[chatID] {
					bot.Request(tgbotapi.NewDeleteMessage(chatID, id))
				}
				messageIDsToDelete[chatID] = nil
				startText := "🔄 Sesi telah direset\. Sila pilih satu pilihan dari menu utama di bawah\." 
				msg := tgbotapi.NewMessage(chatID, startText)
				msg.ReplyMarkup = mainMenuReplyKeyboard
				msg.ParseMode = tgbotapi.ModeMarkdownV2 // Menggunakan MarkdownV2
				if sentMsg, err := bot.Send(msg); err == nil {
					messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
				}
			}
		}
	}
}