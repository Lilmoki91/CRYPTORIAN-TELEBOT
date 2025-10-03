// ==============================
// --- main.go menu & submenu ---
// ==============================

// --- BLOCK 01 ---
// --- SETUP & IMPORT ---

package main

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
    "strings"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// --- STRUCTS UNTUK PANDUAN BIASA ---
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

// --- STRUCTS BARU UNTUK INFOGRAFIK ---
type InfographicStep struct {
    Step    string   `json:"step"`
    Image   string   `json:"image"`
    Details []string `json:"details"`
    Arrow   string   `json:"arrow,omitempty"`
}

type InfographicGuide struct {
    Title     string            `json:"title"`
    ImageMain string            `json:"image_main"`
    Steps     []InfographicStep `json:"steps"`
}

// --- HELPER: Escape Markdown V2 ---
func escapeMarkdownV2(text string) string {
    // Karakter khusus MarkdownV2: _ * [ ] ( ) ~ ` > # + - = | { } . !
    toEscape := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
    for _, char := range toEscape {
        text = strings.ReplaceAll(text, char, "\\"+char)
    }
    return text
}

// --- BLOCK 02 ---
// --- DEFINISI PAPAN KEKUNCI (KEYBOARD) ---

var mainMenuKeyboard = tgbotapi.NewInlineKeyboardMarkup(
    tgbotapi.NewInlineKeyboardRow(
        tgbotapi.NewInlineKeyboardButtonData("📚 Panduan Kripto", "nav_guides"),
        tgbotapi.NewInlineKeyboardButtonData("🔗 Pautan & Bantuan", "nav_links"),
    ),
    tgbotapi.NewInlineKeyboardRow(
        tgbotapi.NewInlineKeyboardButtonData("📊 Infografik", "get_infographic"),
        tgbotapi.NewInlineKeyboardButtonData("♻️ Reset Mesej", "action_reset"),
    ),
)

var guidesMenuKeyboard = tgbotapi.NewInlineKeyboardMarkup(
    tgbotapi.NewInlineKeyboardRow(
        tgbotapi.NewInlineKeyboardButtonData("Claim Worldcoin", "get_guide_claim"),
        tgbotapi.NewInlineKeyboardButtonData("Wallet HATA", "get_guide_wallet"),
    ),
    tgbotapi.NewInlineKeyboardRow(
        tgbotapi.NewInlineKeyboardButtonData("Proses Cashout", "get_guide_cashout"),
    ),
    tgbotapi.NewInlineKeyboardRow(
        tgbotapi.NewInlineKeyboardButtonData("« Kembali ke Menu Utama", "nav_main"),
    ),
)

var linksMenuKeyboard = tgbotapi.NewInlineKeyboardMarkup(
    tgbotapi.NewInlineKeyboardRow(
        tgbotapi.NewInlineKeyboardButtonURL("📢 Channel Telegram", "https://t.me/cucikripto"),
        tgbotapi.NewInlineKeyboardButtonURL("🆘 Hubungi Admin", "https://t.me/johansetia"),
    ),
    tgbotapi.NewInlineKeyboardRow(
        tgbotapi.NewInlineKeyboardButtonData("« Kembali ke Menu Utama", "nav_main"),
    ),
)

// --- BLOCK 03: FUNGSI PENGIRIMAN ---

func sendInfographicGuide(bot *tgbotapi.BotAPI, chatID int64, guide InfographicGuide, messageIDs *map[int64][]int) {
    // Kirim judul
    titleMsg := tgbotapi.NewMessage(chatID, escapeMarkdownV2(guide.Title))
    titleMsg.ParseMode = tgbotapi.ModeMarkdownV2
    if sent, err := bot.Send(titleMsg); err == nil {
        (*messageIDs)[chatID] = append((*messageIDs)[chatID], sent.MessageID)
    }

    // Kirim gambar utama
    if guide.ImageMain != "" {
        cleanURL := strings.TrimSpace(guide.ImageMain)
        photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(cleanURL))
        if sent, err := bot.Send(photo); err == nil {
            (*messageIDs)[chatID] = append((*messageIDs)[chatID], sent.MessageID)
        }
    }

    // Kirim setiap langkah
    for _, step := range guide.Steps {
        var caption strings.Builder
        caption.WriteString(fmt.Sprintf("*%s*\n\n", escapeMarkdownV2(step.Step)))
        for _, detail := range step.Details {
            caption.WriteString(fmt.Sprintf("`\\-` %s\n", escapeMarkdownV2(detail)))
        }
        if step.Arrow != "" {
            caption.WriteString(fmt.Sprintf("\n%s", escapeMarkdownV2(step.Arrow)))
        }

        cleanImageURL := strings.TrimSpace(step.Image)
        photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(cleanImageURL))
        photo.Caption = caption.String()
        photo.ParseMode = tgbotapi.ModeMarkdownV2

        if sent, err := bot.Send(photo); err == nil {
            (*messageIDs)[chatID] = append((*messageIDs)[chatID], sent.MessageID)
        }
    }
}

func sendDetailedGuide(bot *tgbotapi.BotAPI, chatID int64, guide Guide, messageIDs *map[int64][]int) {
    text := fmt.Sprintf("Panduan untuk *%s* sedang dalam pembangunan.\n\nℹ️ Untuk maklumat segera, sila gunakan *Infografik* atau hubungi admin.", escapeMarkdownV2(guide.Title))
    msg := tgbotapi.NewMessage(chatID, text)
    msg.ParseMode = tgbotapi.ModeMarkdownV2
    if sent, err := bot.Send(msg); err == nil {
        (*messageIDs)[chatID] = append((*messageIDs)[chatID], sent.MessageID)
    }
}

// --- BLOCK 04: MAIN ---

func main() {
    botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
    if botToken == "" {
        log.Fatal("TELEGRAM_BOT_TOKEN mesti ditetapkan")
    }

    bot, err := tgbotapi.NewBotAPI(botToken)
    if err != nil {
        log.Panic(err)
    }
    bot.Debug = false
    log.Printf("Bot dimulakan: @%s", bot.Self.UserName)

    log.Println("Memuatkan panduan dari markdown.json...")
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

    messageIDsToDelete := make(map[int64][]int)

    for update := range updates {
        if update.CallbackQuery != nil {
            cq := update.CallbackQuery
            bot.AnswerCallbackQuery(tgbotapi.NewCallback(cq.ID, ""))

            chatID := cq.Message.Chat.ID
            // Opsional: hapus pesan lama sebelum kirim baru
            // bot.Request(tgbotapi.NewDeleteMessage(chatID, cq.Message.MessageID))

            switch cq.Data {
            case "nav_main":
                text := "👋 Selamat Datang! Sila pilih satu pilihan di bawah:"
                editMsg := tgbotapi.NewEditMessageTextAndMarkup(chatID, cq.Message.MessageID, text, mainMenuKeyboard)
                bot.Request(editMsg)

            case "nav_guides":
                text := "📚 *Panduan Kripto*\n\nPilih panduan yang anda mahu lihat:"
                editMsg := tgbotapi.NewEditMessageTextAndMarkup(chatID, cq.Message.MessageID, text, guidesMenuKeyboard)
                editMsg.ParseMode = tgbotapi.ModeMarkdownV2
                bot.Request(editMsg)

            case "nav_links":
                text := "🔗 *Pautan \\& Bantuan*\n\nPilih satu pautan di bawah:"
                editMsg := tgbotapi.NewEditMessageTextAndMarkup(chatID, cq.Message.MessageID, text, linksMenuKeyboard)
                editMsg.ParseMode = tgbotapi.ModeMarkdownV2
                bot.Request(editMsg)

            case "get_guide_claim", "get_guide_wallet", "get_guide_cashout":
                // Hapus pesan lama
                bot.Request(tgbotapi.NewDeleteMessage(chatID, cq.Message.MessageID))
                var guideData Guide
                var key string
                switch cq.Data {
                case "get_guide_claim":
                    key = "worldcoin_registration_guide"
                case "get_guide_wallet":
                    key = "hata_setup_guide"
                case "get_guide_cashout":
                    key = "cashout_guide"
                }
                if raw, exists := guides[key]; exists {
                    if err := json.Unmarshal(raw, &guideData); err != nil {
                        log.Printf("Gagal unmarshal %s: %v", key, err)
                        bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Gagal memuat panduan. Sila cuba lagi."))
                        return
                    }
                    sendDetailedGuide(bot, chatID, guideData, &messageIDsToDelete)
                } else {
                    bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Panduan tidak dijumpai."))
                }

            case "get_infographic":
                bot.Request(tgbotapi.NewDeleteMessage(chatID, cq.Message.MessageID))
                var infographicData InfographicGuide
                if raw, exists := guides["infographic_guide"]; exists {
                    if err := json.Unmarshal(raw, &infographicData); err != nil {
                        log.Printf("Ralat unmarshal infografik: %v", err)
                        bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Gagal memproses infografik."))
                    } else {
                        sendInfographicGuide(bot, chatID, infographicData, &messageIDsToDelete)
                    }
                } else {
                    bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Data infografik tidak dijumpai."))
                }

            case "action_reset":
                for _, id := range messageIDsToDelete[chatID] {
                    bot.Request(tgbotapi.NewDeleteMessage(chatID, id))
                }
                messageIDsToDelete[chatID] = nil
                startText := "🔄 Sesi telah di\\_reset\\. Sila pilih satu pilihan di bawah:"
                msg := tgbotapi.NewMessage(chatID, startText)
                msg.ParseMode = tgbotapi.ModeMarkdownV2
                msg.ReplyMarkup = mainMenuKeyboard
                if sent, err := bot.Send(msg); err == nil {
                    messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sent.MessageID)
                }
            }
            continue
        }

        if update.Message != nil {
            chatID := update.Message.Chat.ID
            messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], update.Message.MessageID)

            if update.Message.IsCommand() && update.Message.Command() == "start" {
                text := "👋 Selamat Datang! Sila pilih satu pilihan di bawah:"
                msg := tgbotapi.NewMessage(chatID, text)
                msg.ReplyMarkup = mainMenuKeyboard
                if sent, err := bot.Send(msg); err == nil {
                    messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sent.MessageID)
                }
            }
        }
    }
}
