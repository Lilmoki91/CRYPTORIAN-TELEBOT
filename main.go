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

func escapeMarkdownV2(text string) string {
    re := regexp.MustCompile(`([_{}\[\]()~>#\+\-=|.!])`)
    return re.ReplaceAllString(text, `\$1`)
}

// ✅ versi fixed (tak guna cast ke interface)
func sendDetailedGuide(bot *tgbotapi.BotAPI, chatID int64, guide Guide, store map[int64][]int) {
    // title
    titleMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("📖 *%s*", escapeMarkdownV2(guide.Title)))
    titleMsg.ParseMode = tgbotapi.ModeMarkdownV2
    sent, _ := bot.Send(titleMsg)
    store[chatID] = append(store[chatID], sent.MessageID)

    // steps
    for _, step := range guide.Steps {
        var caption strings.Builder
        caption.WriteString(fmt.Sprintf("*%s*\n\n", escapeMarkdownV2(step.Title)))
        caption.WriteString(escapeMarkdownV2(step.Desc))

        if len(step.Images) == 0 {
            msg := tgbotapi.NewMessage(chatID, caption.String())
            msg.ParseMode = tgbotapi.ModeMarkdownV2
            sent, _ := bot.Send(msg)
            store[chatID] = append(store[chatID], sent.MessageID)
        } else if len(step.Images) == 1 {
            photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(step.Images[0]))
            photo.Caption = caption.String()
            photo.ParseMode = tgbotapi.ModeMarkdownV2
            sent, _ := bot.Send(photo)
            store[chatID] = append(store[chatID], sent.MessageID)
        } else {
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

            // ✅ FIX: simpan message pertama sahaja
            sent, _ := bot.Send(album)
            store[chatID] = append(store[chatID], sent.MessageID)
        }
    }

    // notes
    if len(guide.Important.Notes) > 0 {
        var notesBuilder strings.Builder
        notesBuilder.WriteString(fmt.Sprintf("*%s*\n\n", escapeMarkdownV2(guide.Important.Title)))
        for _, note := range guide.Important.Notes {
            notesBuilder.WriteString(fmt.Sprintf("• %s\n", escapeMarkdownV2(note)))
        }
        msg := tgbotapi.NewMessage(chatID, notesBuilder.String())
        msg.ParseMode = tgbotapi.ModeMarkdownV2
        sent, _ := bot.Send(msg)
        store[chatID] = append(store[chatID], sent.MessageID)
    }
}

func main() {
    jsonData, err := os.ReadFile("markdown.json")
    if err != nil {
        log.Panic("Error reading markdown.json: ", err)
    }
    var guides map[string]Guide
    if err := json.Unmarshal(jsonData, &guides); err != nil {
        log.Panic("JSON parsing error: ", err)
    }

    botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
    if botToken == "" {
        log.Panic("TELEGRAM_BOT_TOKEN environment variable not set")
    }

    bot, err := tgbotapi.NewBotAPI(botToken)
    if err != nil {
        log.Panic("Token error: ", err)
    }
    log.Printf("Bot started: @%s", bot.Self.UserName)

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

    u := tgbotapi.NewUpdate(0)
    u.Timeout = 60
    updates := bot.GetUpdatesChan(u)

    var messageIDsToDelete = make(map[int64][]int)

    for update := range updates {
        if update.CallbackQuery != nil {
            bot.Request(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))

            if update.CallbackQuery.Data == "show_main_menu" {
                chatID := update.CallbackQuery.Message.Chat.ID
                text := "Sila guna *Butang Aksi Pantas* di bawah..."
                msg := tgbotapi.NewMessage(chatID, text)
                msg.ParseMode = tgbotapi.ModeMarkdownV2
                msg.ReplyMarkup = mainMenuKeyboard
                sentMsg, _ := bot.Send(msg)
                messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)

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
            msg := tgbotapi.NewMessage(chatID, "👋 Selamat Datang! Tekan butang di bawah untuk memaparkan menu.")
            msg.ParseMode = tgbotapi.ModeMarkdownV2
            msg.ReplyMarkup = initialKeyboard
            sentMsg, _ := bot.Send(msg)
            messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)

        case "🚀 Claim WorldCoin":
            text := "Untuk mendaftar *WorldCoin*..."
            msg := tgbotapi.NewMessage(chatID, text)
            msg.ParseMode = tgbotapi.ModeMarkdownV2
            sentMsg, _ := bot.Send(msg)
            messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)

        case "💰 Wallet HATA":
            text := "Untuk mendaftar *HATA Wallet*..."
            msg := tgbotapi.NewMessage(chatID, text)
            msg.ParseMode = tgbotapi.ModeMarkdownV2
            sentMsg, _ := bot.Send(msg)
            messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)

        case "💸 Cashout ke Bank":
            text := "Panduan untuk menjual WorldCoin..."
            msg := tgbotapi.NewMessage(chatID, text)
            msg.ParseMode = tgbotapi.ModeMarkdownV2
            sentMsg, _ := bot.Send(msg)
            messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)

        case "📢 Channel":
            text := "Sertai saluran Telegram kami..."
            msg := tgbotapi.NewMessage(chatID, text)
            msg.ParseMode = tgbotapi.ModeMarkdownV2
            sentMsg, _ := bot.Send(msg)
            messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)

        case "👨‍💻 Admin":
            text := "Perlukan bantuan lanjut? Hubungi admin..."
            msg := tgbotapi.NewMessage(chatID, text)
            msg.ParseMode = tgbotapi.ModeMarkdownV2
            sentMsg, _ := bot.Send(msg)
            messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)

        case "/claim":
            sendDetailedGuide(bot, chatID, guides["worldcoin_registration_guide"], messageIDsToDelete)

        case "/wallet", "/wallet hata":
            sendDetailedGuide(bot, chatID, guides["hata_setup_guide"], messageIDsToDelete)

        case "/cashout":
            sendDetailedGuide(bot, chatID, guides["cashout_guide"], messageIDsToDelete)

        case "/reset mesej", "🔄 Reset (Padam Mesej)":
            for _, id := range messageIDsToDelete[chatID] {
                bot.Request(tgbotapi.NewDeleteMessage(chatID, id))
            }
            messageIDsToDelete[chatID] = nil

            msg := tgbotapi.NewMessage(chatID, "🔄 Sesi telah direset semula. Tekan butang di bawah untuk mula sekali lagi.")
            msg.ParseMode = tgbotapi.ModeMarkdownV2
            msg.ReplyMarkup = initialKeyboard
            sentMsg, _ := bot.Send(msg)
            messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)

        default:
            msg := tgbotapi.NewMessage(chatID, "❌ Arahan tidak dikenali. Sila guna butang atau arahan yang sah.")
            sentMsg, _ := bot.Send(msg)
            messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
        }
    }
}
