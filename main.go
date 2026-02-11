package main

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
    "regexp"
    "strings"
    "sync"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// --- KONSTAN AUDIO ---
// Pautan telah dikemas kini kepada URL RAW GitHub agar Telegram boleh mengakses fail audio secara langsung.
const WELCOME_JINGLE_URL = "https://raw.githubusercontent.com/Lilmoki91/CRYPTORIAN-TELEBOT/main/assets/Selamat_datang.mp3"

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

// --- CACHE UNTUK PANDUAN ---
var (
    worldcoinGuide  Guide
    hataGuide       Guide
    cashoutGuide    Guide
    infographicGuide InfographicGuide
    guidesLoaded    = false
    loadGuidesMutex sync.Mutex
)

// --- DEFINISI PAPAN KEKUNCI (KEYBOARD) ---
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
        tgbotapi.NewInlineKeyboardButtonURL("🛄 Wallet HATA", "https://hata.io/signup?ref=HDX8778"),
    ),
    tgbotapi.NewInlineKeyboardRow(
        tgbotapi.NewInlineKeyboardButtonURL("📢 Channel Telegram", "https://t.me/cucikripto"),
        tgbotapi.NewInlineKeyboardButtonURL("🆘 Hubungi Admin", "https://t.me/johansetia"),
    ),
    tgbotapi.NewInlineKeyboardRow(
        tgbotapi.NewInlineKeyboardButtonData("« Tutup Menu Ini", "close_menu"),
    ),
)

// --- FUNGSI-FUNGSI BANTUAN ---

// escapeMarkdownV2 mengelak aksara khas MarkdownV2
func escapeMarkdownV2(text string) string {
    if text == "" {
        return ""
    }
    // Aksara khas: _, *, [, ], (, ), ~, `, >, #, +, -, =, |, {, }, ., !
    re := regexp.MustCompile(`([_*\[\]()~>\#\+\-=|{}.!])`)
    return re.ReplaceAllString(text, `\$1`)
}

// loadGuides memuatkan semua panduan dari fail JSON sekali sahaja
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

    // Unmarshal setiap panduan
    if err := json.Unmarshal(rawGuides["worldcoin_registration_guide"], &worldcoinGuide); err != nil {
        return fmt.Errorf("gagal memuatkan worldcoin_registration_guide: %v", err)
    }
    if err := json.Unmarshal(rawGuides["hata_setup_guide"], &hataGuide); err != nil {
        return fmt.Errorf("gagal memuatkan hata_setup_guide: %v", err)
    }
    if err := json.Unmarshal(rawGuides["cashout_guide"], &cashoutGuide); err != nil {
        return fmt.Errorf("gagal memuatkan cashout_guide: %v", err)
    }
    if err := json.Unmarshal(rawGuides["infographic_guide"], &infographicGuide); err != nil {
        return fmt.Errorf("gagal memuatkan infographic_guide: %v", err)
    }

    guidesLoaded = true
    log.Println("✓ Semua panduan berjaya dimuatkan ke cache")
    return nil
}

// addMessageID menambah ID mesej ke senarai dengan selamat (thread-safe)
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

// sendDetailedGuide menghantar panduan berstruktur langkah demi langkah
func sendDetailedGuide(bot *tgbotapi.BotAPI, chatID int64, guide Guide, messageIDs *map[int64][]int, mu *sync.Mutex) {
    // Hantar tajuk panduan
    titleMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("📖 *%s*", escapeMarkdownV2(guide.Title)))
    titleMsg.ParseMode = tgbotapi.ModeMarkdownV2
    if sentMsg, err := bot.Send(titleMsg); err == nil {
        addMessageID(messageIDs, mu, chatID, sentMsg.MessageID)
    }

    // Hantar setiap langkah
    for _, step := range guide.Steps {
        var caption strings.Builder
        caption.WriteString(fmt.Sprintf("*%s*\n\n", escapeMarkdownV2(step.Title)))
        caption.WriteString(escapeMarkdownV2(step.Desc))

        if len(step.Images) == 0 {
            msg := tgbotapi.NewMessage(chatID, caption.String())
            msg.ParseMode = tgbotapi.ModeMarkdownV2
            if sentMsg, err := bot.Send(msg); err == nil {
                addMessageID(messageIDs, mu, chatID, sentMsg.MessageID)
            }
        } else if len(step.Images) == 1 {
            photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(step.Images[0]))
            photo.Caption = caption.String()
            photo.ParseMode = tgbotapi.ModeMarkdownV2
            if sentMsg, err := bot.Send(photo); err == nil {
                addMessageID(messageIDs, mu, chatID, sentMsg.MessageID)
            }
        } else {
            // Menggunakan MediaGroup untuk album foto
            mediaGroup := []interface{}{}
            for i, imgURL := range step.Images {
                photo := tgbotapi.NewInputMediaPhoto(tgbotapi.FileURL(imgURL))
                if i == 0 {
                    photo.Caption = caption.String()
                    photo.ParseMode = tgbotapi.ModeMarkdownV2
                }
                mediaGroup = append(mediaGroup, photo)
            }
            
            if len(mediaGroup) > 0 {
                album := tgbotapi.NewMediaGroup(chatID, mediaGroup)
                if sentMessages, err := bot.SendMediaGroup(album); err == nil {
                    for _, msg := range sentMessages {
                        addMessageID(messageIDs, mu, chatID, msg.MessageID)
                    }
                } else {
                    log.Printf("Gagal menghantar MediaGroup: %v", err)
                }
            }
        }
    }

    // Hantar nota penting
    if len(guide.Important.Notes) > 0 {
        var notesBuilder strings.Builder
        notesBuilder.WriteString(fmt.Sprintf("\n*%s*\n", escapeMarkdownV2(guide.Important.Title)))
        for _, note := range guide.Important.Notes {
            notesBuilder.WriteString(fmt.Sprintf("\\- %s\n", escapeMarkdownV2(note)))
        }

        msg := tgbotapi.NewMessage(chatID, notesBuilder.String())
        msg.ParseMode = tgbotapi.ModeMarkdownV2
        if sentMsg, err := bot.Send(msg); err == nil {
            addMessageID(messageIDs, mu, chatID, sentMsg.MessageID)
        }
    }
}

// sendInfographicGuide menghantar panduan dalam bentuk infografik
func sendInfographicGuide(bot *tgbotapi.BotAPI, chatID int64, guide InfographicGuide, messageIDs *map[int64][]int, mu *sync.Mutex) {
    // Hantar tajuk
    titleMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("*%s*", escapeMarkdownV2(guide.Title)))
    titleMsg.ParseMode = tgbotapi.ModeMarkdownV2
    if sentMsg, err := bot.Send(titleMsg); err == nil {
        addMessageID(messageIDs, mu, chatID, sentMsg.MessageID)
    }

    // Hantar gambar utama
    if guide.ImageMain != "" {
        photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(guide.ImageMain))
        if sentMsg, err := bot.Send(photo); err == nil {
            addMessageID(messageIDs, mu, chatID, sentMsg.MessageID)
        }
    }

    // Hantar setiap langkah infografik
    for _, step := range guide.Steps {
        var caption strings.Builder
        caption.WriteString(fmt.Sprintf("*%s*\n\n", escapeMarkdownV2(step.Step)))
        
        for _, detail := range step.Details {
            caption.WriteString(fmt.Sprintf("\\- %s\n", escapeMarkdownV2(detail)))
        }
        
        if step.Arrow != "" {
            caption.WriteString(fmt.Sprintf("\n%s\n", escapeMarkdownV2(step.Arrow)))
        }

        photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(step.Image))
        photo.Caption = caption.String()
        photo.ParseMode = tgbotapi.ModeMarkdownV2
        if sentMsg, err := bot.Send(photo); err == nil {
            addMessageID(messageIDs, mu, chatID, sentMsg.MessageID)
        }
    }
}

func main() {
    // Dapatkan token dari environment variable
    botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
    if botToken == "" {
        log.Fatal("❌ TELEGRAM_BOT_TOKEN mesti ditetapkan")
    }

    // Inisialisasi bot
    bot, err := tgbotapi.NewBotAPI(botToken)
    if err != nil {
        log.Panicf("❌ Gagal inisialisasi bot: %v", err)
    }
    log.Printf("✅ Bot Hibrid UI dimulakan: @%s", bot.Self.UserName)

    // Muatkan panduan ke cache
    if err := loadGuides(); err != nil {
        log.Fatalf("❌ %v", err)
    }

    // Setup update channel
    u := tgbotapi.NewUpdate(0)
    u.Timeout = 60
    updates := bot.GetUpdatesChan(u)

    // Map untuk menyimpan ID mesej dan mutex untuk keselamatan konkurensi
    var (
        messageIDsToDelete = make(map[int64][]int)
        mu                 sync.Mutex
    )

    // Loop utama untuk memproses updates
    for update := range updates {
        // --- PROSES CALLBACK QUERY ---
        if update.CallbackQuery != nil {
            // Akui callback query
            bot.Request(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
            
            // Semak jika mesej wujud
            if update.CallbackQuery.Message == nil {
                continue
            }
            
            chatID := update.CallbackQuery.Message.Chat.ID
            messageID := update.CallbackQuery.Message.MessageID

            // Hapus mesej butang yang ditekan
            bot.Request(tgbotapi.NewDeleteMessage(chatID, messageID))

            // Proses callback data
            switch update.CallbackQuery.Data {
            case "close_menu":
                continue
            case "get_guide_claim":
                sendDetailedGuide(bot, chatID, worldcoinGuide, &messageIDsToDelete, &mu)
            case "get_guide_wallet":
                sendDetailedGuide(bot, chatID, hataGuide, &messageIDsToDelete, &mu)
            case "get_guide_cashout":
                sendDetailedGuide(bot, chatID, cashoutGuide, &messageIDsToDelete, &mu)
            }
            continue
        }

        // --- PROSES MESEJ TEKS ---
        if update.Message != nil && update.Message.Text != "" {
            chatID := update.Message.Chat.ID
            
            // Tambah mesej pengguna ke senarai untuk pembersihan
            addMessageID(&messageIDsToDelete, &mu, chatID, update.Message.MessageID)

            // Proses arahan berdasarkan teks mesej
            switch update.Message.Text {
            case "/start", "🔙 Kembali Menu Utama":
                // Hantar audio selamat datang
                audio := tgbotapi.NewAudio(chatID, tgbotapi.FileURL(WELCOME_JINGLE_URL))
                audio.Caption = "🎶 Selamat datang ke Cryptorian\\!"
                audio.ParseMode = tgbotapi.ModeMarkdownV2
                audio.DisableNotification = true

                if sentAudio, err := bot.Send(audio); err == nil {
                    addMessageID(&messageIDsToDelete, &mu, chatID, sentAudio.MessageID)
                } else {
                    log.Printf("Gagal menghantar audio: %v", err)
                    errMsg := "⚠️ Gagal memainkan jingle\\. Pastikan URL audio adalah sah dan boleh diakses awam\\."
                    errorMsg := tgbotapi.NewMessage(chatID, errMsg)
                    errorMsg.ParseMode = tgbotapi.ModeMarkdownV2
                    if sentError, err := bot.Send(errorMsg); err == nil {
                        addMessageID(&messageIDsToDelete, &mu, chatID, sentError.MessageID)
                    }
                }

                // Hantar menu utama
                text := "👋 Selamat Datang ke 🤖 Cryptorian\\-Telebot\\!\\, Sila pilih satu pilihan dari menu utama di bawah\\."
                msg := tgbotapi.NewMessage(chatID, text)
                msg.ReplyMarkup = mainMenuReplyKeyboard
                msg.ParseMode = tgbotapi.ModeMarkdownV2
                if sentMsg, err := bot.Send(msg); err == nil {
                    addMessageID(&messageIDsToDelete, &mu, chatID, sentMsg.MessageID)
                }

            case "📚 Panduan Kripto":
                text := "*📚 Panduan Kripto*\n\nPilih satu panduan dari sub\\-menu di bawah:"
                msg := tgbotapi.NewMessage(chatID, text)
                msg.ParseMode = tgbotapi.ModeMarkdownV2
                msg.ReplyMarkup = guidesInlineKeyboard
                if sentMsg, err := bot.Send(msg); err == nil {
                    addMessageID(&messageIDsToDelete, &mu, chatID, sentMsg.MessageID)
                }

            case "🔗 Pautan & 🆘 Bantuan":
                text := "*🔗 Pautan & 🆘 Bantuan*\n\nPilih satu pautan dari sub\\-menu di bawah:"
                msg := tgbotapi.NewMessage(chatID, text)
                msg.ParseMode = tgbotapi.ModeMarkdownV2
                msg.ReplyMarkup = linksInlineKeyboard
                if sentMsg, err := bot.Send(msg); err == nil {
                    addMessageID(&messageIDsToDelete, &mu, chatID, sentMsg.MessageID)
                }

            case "📊 Infografik":
                sendInfographicGuide(bot, chatID, infographicGuide, &messageIDsToDelete, &mu)

            case "♻️ Reset Mesej":
                // Padam semua mesej dalam sesi
                mu.Lock()
                if ids, exists := messageIDsToDelete[chatID]; exists {
                    for _, id := range ids {
                        bot.Request(tgbotapi.NewDeleteMessage(chatID, id))
                    }
                    delete(messageIDsToDelete, chatID)
                }
                mu.Unlock()

                // Hantar mesej reset
                startText := "🔄 Sesi telah direset\\. Sila pilih satu pilihan dari menu utama di bawah\\."
                msg := tgbotapi.NewMessage(chatID, startText)
                msg.ReplyMarkup = mainMenuReplyKeyboard
                msg.ParseMode = tgbotapi.ModeMarkdownV2
                if sentMsg, err := bot.Send(msg); err == nil {
                    addMessageID(&messageIDsToDelete, &mu, chatID, sentMsg.MessageID)
                }
            }
        }
    }
}
