
        sb.WriteString("\n")
    }

    if len(section.Important.Notes) > 0 {
        sb.WriteString(fmt.Sprintf("⚡ *%s:*\n", EscapeMarkdown(section.Important.Title)))
        for _, n := range section.Important.Notes {
            sb.WriteString(fmt.Sprintf("• %s\n", EscapeMarkdown(n)))
        }
        sb.WriteString("\n")
    }

    return sb.String()
}

// ================================
// 🔐⛓️ Block 4 - Main Function
// ================================

func main() {
    botToken := os.Getenv("BOT_TOKEN")
    if botToken == "" {
        log.Fatal("BOT_TOKEN not set")
    }

    bot, err := tgbotapi.NewBotAPI(botToken)
    if err != nil {
        log.Fatal(err)
    }

    bot.Debug = true
    log.Printf("Bot %s started!", bot.Self.UserName)

    // Load JSON
    data, err := LoadMarkdownData("config/markdown.json")
    if err != nil {
        log.Fatalf("Error load JSON: %v", err) // Guna log.Fatalf untuk error
    }

    u := tgbotapi.NewUpdate(0)
    u.Timeout = 60
    updates := bot.GetUpdatesChan(u)

    for update := range updates {
        if update.Message != nil && update.Message.IsCommand() {
            switch update.Message.Command() {
            case "start":
                handleStart(bot, update.Message.Chat.ID)
            case "claim":
                sendSection(bot, update.Message.Chat.ID, data.WorldcoinRegistration)
            case "hata_wallet":
                sendSection(bot, update.Message.Chat.ID, data.HataWalletSetup)
            case "cashout":
                sendSection(bot, update.Message.Chat.ID, data.CashoutToBank)
            case "security":
                sendSection(bot, update.Message.Chat.ID, data.SecurityNotes)
            }
        }

        if update.CallbackQuery != nil {
            handleCallback(bot, update.CallbackQuery)
        }
    }
}

// ================================
// 🔐⛓️ Block 5 - Handler
// ================================
// Hantar mesej pembukaan dengan 5 butang

func handleStart(bot *tgbotapi.BotAPI, chatID int64) {
    welcomeText := "👋 *Selamat datang ke CRYPTORIAN BOT*\n\nPilih aksi pantas di bawah 🚀"
    msg := tgbotapi.NewMessage(chatID, welcomeText)
    msg.ParseMode = "MarkdownV2"

    keyboard := tgbotapi.NewInlineKeyboardMarkup(
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonURL("🪙 Claim", "https://worldcoin.org/join/4RH0OTE"),
        ),
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonURL("💳 Hata Wallet", "https://hata.io"),
        ),
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonURL("📢 Telegram Channel", "https://t.me/yourchannel"),
        ),
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonURL("👨‍💻 Admin", "https://t.me/yourusername"),
        ),
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData("♻️ Reset", "reset"),
        ),
    )

    msg.ReplyMarkup = keyboard
    bot.Send(msg)
}

// Hantar section ikut command
func sendSection(bot *tgbotapi.BotAPI, chatID int64, section Section) {
    text := FormatSection(section)
    msg := tgbotapi.NewMessage(chatID, text)
    msg.ParseMode = "MarkdownV2"
    bot.Send(msg)

    // Hantar gambar (jika ada)
    for _, step := range section.Steps {
        for _, img := range step.Images {
            photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(img))
            bot.Send(photo)
        }
    }
}

// Handle Callback Inline Button
func handleCallback(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
    chatID := callback.Message.Chat.ID
    data := callback.Data

    if data == "reset" {
        delMsg := tgbotapi.NewDeleteMessage(chatID, callback.Message.MessageID)
        bot.Send(delMsg)

        confirm := tgbotapi.NewMessage(chatID, "✅ Semua mesej dipadam. Tekan /start untuk mula semula.")
        bot.Send(confirm)
    }
}           
