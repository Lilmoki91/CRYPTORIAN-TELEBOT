package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "net/url"
    "os"
    "regexp"
    "strings"
    "time"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// --- KONSTAN & STRUCTS ---
const WELCOME_JINGLE_URL = "https://raw.githubusercontent.com/Lilmoki91/CRYPTORIAN-TELEBOT/main/assets/Selamat_datang.mp3"

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

// --- KEYBOARDS ---
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

// --- HELPERS ---
func escapeMarkdownV2(text string) string {
    re := regexp.MustCompile(`([_*\[\]()~>#+\-=|{}.!])`)
    return re.ReplaceAllString(text, `\$1`)
}

func sendDetailedGuide(bot *tgbotapi.BotAPI, chatID int64, guide Guide, messageIDs *map[int64][]int) {
    titleMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("📖 *%s*", escapeMarkdownV2(guide.Title)))
    titleMsg.ParseMode = tgbotapi.ModeMarkdownV2
    if sentMsg, err := bot.Send(titleMsg); err == nil {
        (*messageIDs)[chatID] = append((*messageIDs)[chatID], sentMsg.MessageID)
    }

    for _, step := range guide.Steps {
        var caption strings.Builder
        caption.WriteString(fmt.Sprintf("*%s*\n\n", escapeMarkdownV2(step.Title)))
        caption.WriteString(escapeMarkdownV2(step.Desc))

        if len(step.Images) == 0 {
            msg := tgbotapi.NewMessage(chatID, caption.String())
            msg.ParseMode = tgbotapi.ModeMarkdownV2
            if sentMsg, err := bot.Send(msg); err == nil {
                (*messageIDs)[chatID] = append((*messageIDs)[chatID], sentMsg.MessageID)
            }
        } else if len(step.Images) == 1 {
            photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(step.Images[0]))
            photo.Caption = caption.String()
            photo.ParseMode = tgbotapi.ModeMarkdownV2
            if sentMsg, err := bot.Send(photo); err == nil {
                (*messageIDs)[chatID] = append((*messageIDs)[chatID], sentMsg.MessageID)
            }
        }
    }
}

func sendInfographicGuide(bot *tgbotapi.BotAPI, chatID int64, guide InfographicGuide, messageIDs *map[int64][]int) {
    titleMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("*%s*", escapeMarkdownV2(guide.Title)))
    titleMsg.ParseMode = tgbotapi.ModeMarkdownV2
    if sentMsg, err := bot.Send(titleMsg); err == nil {
        (*messageIDs)[chatID] = append((*messageIDs)[chatID], sentMsg.MessageID)
    }
    if guide.ImageMain != "" {
        photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(guide.ImageMain))
        if sentMsg, err := bot.Send(photo); err == nil {
            (*messageIDs)[chatID] = append((*messageIDs)[chatID], sentMsg.MessageID)
        }
    }
}

// --- MAIN ---
func main() {
    botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
    webhookURL := os.Getenv("WEBHOOK_URL")
    port := os.Getenv("PORT")
    if port == "" { port = "7860" }

    if botToken == "" {
        log.Fatal("TELEGRAM_BOT_TOKEN mesti ditetapkan")
    }

    // 1. Initialize Bot
    bot, err := tgbotapi.NewBotAPI(botToken)
    if err != nil {
        log.Panic(err)
    }
    log.Printf("Bot Aktif: @%s", bot.Self.UserName)

    // 2. Load Data JSON
    jsonData, err := os.ReadFile("markdown.json")
    if err != nil {
        log.Fatalf("Fail JSON hilang: %v", err)
    }
    var guides map[string]json.RawMessage
    json.Unmarshal(jsonData, &guides)

    // 3. Setup Updates (Satu sahaja punca updates)
    var updates tgbotapi.UpdatesChannel
    if webhookURL != "" {
        parsedURL, _ := url.Parse(webhookURL)
        wh, _ := tgbotapi.NewWebhook(webhookURL)
        bot.Request(wh)
        path := parsedURL.Path
        if path == "" { path = "/" + bot.Token }
        updates = bot.ListenForWebhook(path)
        go http.ListenAndServe(":"+port, nil)
        log.Println("Mode: Webhook aktif")
    } else {
        // Health check untuk Hugging Face
        go func() {
            http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "Bot is running!") })
            http.ListenAndServe(":"+port, nil)
        }()
        u := tgbotapi.NewUpdate(0)
        u.Timeout = 60
        updates = bot.GetUpdatesChan(u)
        log.Println("Mode: Polling aktif")
    }

    messageIDsToDelete := make(map[int64][]int)

    // 4. SATU SAHAJA LOOP UTAMA UNTUK SEMUA MESEJ
    for update := range updates {
        // Logik Callback (Butang Inline)
        if update.CallbackQuery != nil {
            bot.Request(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
            chatID := update.CallbackQuery.Message.Chat.ID
            bot.Request(tgbotapi.NewDeleteMessage(chatID, update.CallbackQuery.Message.MessageID))

            switch update.CallbackQuery.Data {
            case "get_guide_claim":
                var g Guide
                json.Unmarshal(guides["worldcoin_registration_guide"], &g)
                sendDetailedGuide(bot, chatID, g, &messageIDsToDelete)
            case "get_guide_wallet":
                var g Guide
                json.Unmarshal(guides["hata_setup_guide"], &g)
                sendDetailedGuide(bot, chatID, g, &messageIDsToDelete)
            case "get_guide_cashout":
                var g Guide
                json.Unmarshal(guides["cashout_guide"], &g)
                sendDetailedGuide(bot, chatID, g, &messageIDsToDelete)
            case "close_menu":
                continue
            }
            continue
        }

        // Logik Mesej Teks
        if update.Message != nil && update.Message.Text != "" {
            chatID := update.Message.Chat.ID
            messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], update.Message.MessageID)

            switch update.Message.Text {
            case "/start", "🔙 Kembali Menu Utama":
                // Jingle
                audio := tgbotapi.NewAudio(chatID, tgbotapi.FileURL(WELCOME_JINGLE_URL))
                audio.Caption = "🎶 Selamat datang ke Cryptorian\\!"
                audio.ParseMode = tgbotapi.ModeMarkdownV2
                if sent, err := bot.Send(audio); err == nil {
                    messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sent.MessageID)
                }

                msg := tgbotapi.NewMessage(chatID, "👋 Selamat Datang! Sila pilih menu:")
                msg.ReplyMarkup = mainMenuReplyKeyboard
                if sent, err := bot.Send(msg); err == nil {
                    messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sent.MessageID)
                }

            case "📚 Panduan Kripto":
                msg := tgbotapi.NewMessage(chatID, "*📚 Panduan Kripto*")
                msg.ParseMode = tgbotapi.ModeMarkdownV2
                msg.ReplyMarkup = guidesInlineKeyboard
                if sent, err := bot.Send(msg); err == nil {
                    messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sent.MessageID)
                }

            case "🔗 Pautan & 🆘 Bantuan":
                msg := tgbotapi.NewMessage(chatID, "*🔗 Pautan & Bantuan*")
                msg.ParseMode = tgbotapi.ModeMarkdownV2
                msg.ReplyMarkup = linksInlineKeyboard
                if sent, err := bot.Send(msg); err == nil {
                    messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sent.MessageID)
                }

            case "📊 Infografik":
                var g InfographicGuide
                json.Unmarshal(guides["infographic_guide"], &g)
                sendInfographicGuide(bot, chatID, g, &messageIDsToDelete)

            case "♻️ Reset Mesej":
                for _, id := range messageIDsToDelete[chatID] {
                    bot.Request(tgbotapi.NewDeleteMessage(chatID, id))
                }
                messageIDsToDelete[chatID] = nil
                bot.Send(tgbotapi.NewMessage(chatID, "🔄 Sesi direset."))
            }
        }
    }
}
