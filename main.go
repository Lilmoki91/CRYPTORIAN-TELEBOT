// --- BLOCK 01: SETUP & IMPORT ---
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

// --- STRUCTS UNTUK INFOGRAFIK ---
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

// --- HELPER: Escape untuk MarkdownV2 ---
func escapeMarkdownV2(text string) string {
    chars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
    for _, c := range chars {
        text = strings.ReplaceAll(text, c, "\\"+c)
    }
    return text
}

// --- BLOCK 02: KEYBOARDS ---
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

    // Kirim gambar utama (jika ada)
    if guide.ImageMain != "" {
        url := strings.TrimSpace(guide.ImageMain)
        photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(url))
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

        photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(strings.TrimSpace(step.Image)))
        photo.Caption = caption.String()
        photo.ParseMode = tgbotapi.ModeMarkdownV2

        if sent, err := bot.Send(photo); err == nil {
            (*messageIDs)[chatID] = append((*messageIDs)[chatID], sent.MessageID)
        }
    }
}

func sendDetailedGuide(bot *tgbotapi.BotAPI, chatID int64, guide Guide, messageIDs *map[int64][]int) {
    text := fmt.Sprintf("ℹ️ Panduan *%s* sedang dalam pembangunan.\n\nUntuk maklumat pantas, sila gunakan *📊 Infografik* atau hubungi admin.", escapeMarkdownV2(guide.Title))
    msg := tgbotapi.NewMessage(chatID, text)
    msg.ParseMode = tgbotapi.ModeMarkdownV2
    if sent, err := bot.Send(msg); err == nil {
        (*messageIDs)[chatID] = append((*messageIDs)[chatID], sent.MessageID)
    }
}

// --- BLOCK 04: MAIN FUNCTION ---
func main() {
    botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
    if botToken == "" {
        log.Fatal("Ralat: TELEGRAM_BOT_TOKEN tidak ditetapkan dalam environment variable.")
    }

    bot, err := tgbotapi.NewBotAPI(botToken)
    if err != nil {
        log.Panicf("Gagal sambung ke Telegram API: %v", err)
    }
    log.Printf("✅ Bot aktif: @%s", bot.Self.UserName)

    // Baca markdown.json
    log.Println("Membaca markdown.json...")
    jsonData, err := os.ReadFile("markdown.json")
    if err != nil {
        log.Fatalf("Gagal baca markdown.json: %v", err)
    }

    var guides map[string]json.RawMessage
    if err := json.Unmarshal(jsonData, &guides); err != nil {
        log.Fatalf("JSON tidak sah: %v", err)
    }
    log.Println("✅ markdown.json dimuatkan.")

    // Mulakan polling
    u := tgbotapi.NewUpdate(0)
    u.Timeout = 60
    updates := bot.GetUpdatesChan(u)

    messageIDsToDelete := make(map[int64][]int)

    for update := range updates {
        if update.CallbackQuery != nil {
            cq := update.CallbackQuery
            chatID := cq.Message.Chat.ID

            // ✅ Jawab callback (tanpa error)
            bot.Request(tgbotapi.NewCallback(cq.ID, ""))

            switch cq.Data {
            case "nav_main":
                msg := tgbotapi.NewEditMessageTextAndMarkup(chatID, cq.Message.MessageID, "👋 Selamat Datang! Sila pilih satu pilihan di bawah:", mainMenuKeyboard)
                bot.Request(msg)

            case "nav_guides":
                msg := tgbotapi.NewEditMessageTextAndMarkup(chatID, cq.Message.MessageID, "📚 *Panduan Kripto*\n\nPilih panduan yang anda mahu lihat:", guidesMenuKeyboard)
                msg.ParseMode = tgbotapi.ModeMarkdownV2
                bot.Request(msg)

            case "nav_links":
                msg := tgbotapi.NewEditMessageTextAndMarkup(chatID, cq.Message.MessageID, "🔗 *Pautan \\& Bantuan*\n\nPilih satu pautan di bawah:", linksMenuKeyboard)
                msg.ParseMode = tgbotapi.ModeMarkdownV2
                bot.Request(msg)

            case "get_guide_claim", "get_guide_wallet", "get_guide_cashout":
                bot.Request(tgbotapi.NewDeleteMessage(chatID, cq.Message.MessageID))
                var key string
                switch cq.Data {
                case "get_guide_claim":
                    key = "worldcoin_registration_guide"
                case "get_guide_wallet":
                    key = "hata_setup_guide"
                case "get_guide_cashout":
                    key = "cashout_guide"
                }
                if raw, ok := guides[key]; ok {
                    var g Guide
                    if err := json.Unmarshal(raw, &g); err != nil {
                        log.Printf("Gagal unmarshal %s: %v", key, err)
                        bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Gagal memuat panduan."))
                    } else {
                        sendDetailedGuide(bot, chatID, g, &messageIDsToDelete)
                    }
                } else {
                    bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Panduan tidak dijumpai."))
                }

            case "get_infographic":
                bot.Request(tgbotapi.NewDeleteMessage(chatID, cq.Message.MessageID))
                if raw, ok := guides["infographic_guide"]; ok {
                    var ig InfographicGuide
                    if err := json.Unmarshal(raw, &ig); err != nil {
                        log.Printf("Infografik unmarshal error: %v", err)
                        bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Gagal memproses infografik."))
                    } else {
                        sendInfographicGuide(bot, chatID, ig, &messageIDsToDelete)
                    }
                } else {
                    bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Infografik tidak dijumpai."))
                }

            case "action_reset":
                for _, id := range messageIDsToDelete[chatID] {
                    bot.Request(tgbotapi.NewDeleteMessage(chatID, id))
                }
                messageIDsToDelete[chatID] = nil
                msg := tgbotapi.NewMessage(chatID, "🔄 Sesi telah di\\_reset\\. Sila pilih satu pilihan di bawah:")
                msg.ParseMode = tgbotapi.ModeMarkdownV2
                msg.ReplyMarkup = mainMenuKeyboard
                if sent, err := bot.Send(msg); err == nil {
                    messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sent.MessageID)
                }
            }
            continue
        }

        // Handle /start
        if update.Message != nil {
            chatID := update.Message.Chat.ID
            messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], update.Message.MessageID)

            if update.Message.IsCommand() && update.Message.Command() == "start" {
                msg := tgbotapi.NewMessage(chatID, "👋 Selamat Datang! Sila pilih satu pilihan di bawah:")
                msg.ReplyMarkup = mainMenuKeyboard
                if sent, err := bot.Send(msg); err == nil {
                    messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sent.MessageID)
                }
            }
        }
    }
}
