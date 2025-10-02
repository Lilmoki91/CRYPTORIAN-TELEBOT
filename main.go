
// --- BLOCK 01
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
    Arrow   string   `json:"arrow"`
}
type InfographicGuide struct {
    Title     string            `json:"title"`
    ImageMain string            `json:"image_main"`
    Steps     []InfographicStep `json:"steps"`
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

// --- ⚠️ BLOCK 03 (HELPER FUNCTION) ⚠️
// (Letakkan fungsi escapeMarkdownV2 di sini jika anda mahu gunakannya)

func sendInfographicGuide(bot *tgbotapi.BotAPI, chatID int64, guide InfographicGuide, messageIDs *map[int64][]int) {
    titleMsg := tgbotapi.NewMessage(chatID, guide.Title)
    if sentMsg, err := bot.Send(titleMsg); err == nil {
        (*messageIDs)[chatID] = append((*messageIDs)[chatID], sentMsg.MessageID)
    }
    if guide.ImageMain != "" {
        photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(guide.ImageMain))
        if sentMsg, err := bot.Send(photo); err == nil {
            (*messageIDs)[chatID] = append((*messageIDs)[chatID], sentMsg.MessageID)
        }
    }
    for _, step := range guide.Steps {
        var caption strings.Builder
        caption.WriteString(fmt.Sprintf("*%s*\n\n", step.Step))
        for _, detail := range step.Details {
            caption.WriteString(fmt.Sprintf("`-` %s\n", detail))
        }
        if step.Arrow != "" {
            caption.WriteString(fmt.Sprintf("\n%s\n", step.Arrow))
        }
        photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(step.Image))
        photo.Caption = caption.String()
        photo.ParseMode = tgbotapi.ModeMarkdown
        if sentMsg, err := bot.Send(photo); err == nil {
            (*messageIDs)[chatID] = append((*messageIDs)[chatID], sentMsg.MessageID)
        }
    }
}

// ⚠️ Untuk tujuan ujian, kita juga akan cipta fungsi sendDetailedGuide yang ringkas
func sendDetailedGuide(bot *tgbotapi.BotAPI, chatID int64, guide Guide, messageIDs *map[int64][]int) {
    // Hantar mesej ringkas sebagai ganti panduan penuh
    text := fmt.Sprintf("Panduan untuk *%s* akan dipaparkan di sini.", guide.Title)
    msg := tgbotapi.NewMessage(chatID, text)
    msg.ParseMode = tgbotapi.ModeMarkdown
    if sentMsg, err := bot.Send(msg); err == nil {
        (*messageIDs)[chatID] = append((*messageIDs)[chatID], sentMsg.MessageID)
    }
}

// --- BLOCK 04
// --- MARKDOWN.JSON LOADER (MAIN LOGIC)

func main() {
    botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
    if botToken == "" {
        log.Fatal("TELEGRAM_BOT_TOKEN mesti ditetapkan")
    }

    bot, err := tgbotapi.NewBotAPI(botToken)
    if err != nil {
        log.Panic(err)
    }
    log.Printf("Bot Ujian Menu dimulakan: @%s", bot.Self.UserName)

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

    var messageIDsToDelete = make(map[int64][]int)

    for update := range updates {
        if update.CallbackQuery != nil {
            bot.Request(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
            chatID := update.CallbackQuery.Message.Chat.ID
            messageID := update.CallbackQuery.Message.MessageID

            switch update.CallbackQuery.Data {
            case "nav_main":
                text := "👋 Selamat Datang! Sila pilih satu pilihan di bawah:"
                editMsg := tgbotapi.NewEditMessageTextAndMarkup(chatID, messageID, text, mainMenuKeyboard)
                bot.Request(editMsg)
            case "nav_guides":
                text := "📚 *Panduan Kripto*\n\nPilih panduan yang anda mahu lihat:"
                editMsg := tgbotapi.NewEditMessageTextAndMarkup(chatID, messageID, text, guidesMenuKeyboard)
                editMsg.ParseMode = tgbotapi.ModeMarkdown
                bot.Request(editMsg)
            case "nav_links":
                text := "🔗 *Pautan & Bantuan*\n\nPilih satu pautan di bawah:"
                editMsg := tgbotapi.NewEditMessageTextAndMarkup(chatID, messageID, text, linksMenuKeyboard)
                editMsg.ParseMode = tgbotapi.ModeMarkdown
                bot.Request(editMsg)
            
            case "get_guide_claim":
                bot.Request(tgbotapi.NewDeleteMessage(chatID, messageID))
                var guideData Guide
                json.Unmarshal(guides["worldcoin_registration_guide"], &guideData)
                sendDetailedGuide(bot, chatID, guideData, &messageIDsToDelete)
            
            case "get_guide_wallet":
                bot.Request(tgbotapi.NewDeleteMessage(chatID, messageID))
                var guideData Guide
                json.Unmarshal(guides["hata_setup_guide"], &guideData)
                sendDetailedGuide(bot, chatID, guideData, &messageIDsToDelete)

            case "get_guide_cashout":
                bot.Request(tgbotapi.NewDeleteMessage(chatID, messageID))
                var guideData Guide
                json.Unmarshal(guides["cashout_guide"], &guideData)
                sendDetailedGuide(bot, chatID, guideData, &messageIDsToDelete)

            case "get_infographic":
                bot.Request(tgbotapi.NewDeleteMessage(chatID, messageID))
                var infographicData InfographicGuide
                if err := json.Unmarshal(guides["infographic_guide"], &infographicData); err == nil {
                    sendInfographicGuide(bot, chatID, infographicData, &messageIDsToDelete)
                } else {
                    bot.Send(tgbotapi.NewMessage(chatID, "Gagal memproses data infografik."))
                    log.Printf("Ralat unmarshal infografik: %v", err)
                }
                
            case "action_reset":
                for _, id := range messageIDsToDelete[chatID] {
                    bot.Request(tgbotapi.NewDeleteMessage(chatID, id))
                }
                messageIDsToDelete[chatID] = nil
                startText := "🔄 Sesi telah direset. Sila pilih satu pilihan di bawah:"
                msg := tgbotapi.NewMessage(chatID, startText)
                msg.ReplyMarkup = mainMenuKeyboard
                if sentMsg, err := bot.Send(msg); err == nil {
                    messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
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
                if sentMsg, err := bot.Send(msg); err == nil {
                    messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
                }
            }
        }
    }
}


