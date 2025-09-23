// ===============
// --- MAIN.GO ---
// ===============

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

// --- STRUCTS TO MATCH MARKDOWN.JSON ---
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

// --- HELPER FUNCTIONS ---

// Function to escape text for MarkdownV2
func escapeMarkdownV2(text string) string {
    re := regexp.MustCompile(`([_{}\[\]()~>#\+\-=|.!])`)
    return re.ReplaceAllString(text, `\$1`)
}

// Function to send the detailed, step-by-step guide from markdown.json
func sendDetailedGuide(bot *tgbotapi.BotAPI, chatID int64, guide Guide) {
    // Send main title
    titleMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("📖 *%s*", escapeMarkdownV2(guide.Title)))
    titleMsg.ParseMode = tgbotapi.ModeMarkdownV2
    bot.Send(titleMsg)

    // Send each step
    for _, step := range guide.Steps {
        var caption strings.Builder
        caption.WriteString(fmt.Sprintf("*%s*\n\n", escapeMarkdownV2(step.Title)))
        caption.WriteString(escapeMarkdownV2(step.Desc))

        if len(step.Images) == 0 {
            msg := tgbotapi.NewMessage(chatID, caption.String())
            msg.ParseMode = tgbotapi.ModeMarkdownV2
            bot.Send(msg)
        } else if len(step.Images) == 1 {
            photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(step.Images[0]))
            photo.Caption = caption.String()
            photo.ParseMode = tgbotapi.ModeMarkdownV2
            bot.Send(photo)
        } else {
            // Send as an album if there are multiple images
            var mediaGroup []interface{}
            for i, imgURL := range step.Images {
                photo := tgbotapi.NewInputMediaPhoto(tgbotapi.FileURL(imgURL))
                if i == 0 { // Attach caption to the first photo only
                    photo.Caption = caption.String()
                    photo.ParseMode = tgbotapi.ModeMarkdownV2
                }
                mediaGroup = append(mediaGroup, photo)
            }
            album := tgbotapi.NewMediaGroup(chatID, mediaGroup)
            bot.Send(album)
        }
    }

    // Send important notes if available
    if len(guide.Important.Notes) > 0 {
        var notesBuilder strings.Builder
        notesBuilder.WriteString(fmt.Sprintf("*%s*\n\n", escapeMarkdownV2(guide.Important.Title)))
        for _, note := range guide.Important.Notes {
            notesBuilder.WriteString(fmt.Sprintf("• %s\n", escapeMarkdownV2(note)))
        }
        msg := tgbotapi.NewMessage(chatID, notesBuilder.String())
        msg.ParseMode = tgbotapi.ModeMarkdownV2
        bot.Send(msg)
    }
}

func main() {
    // --- INITIALIZATION ---
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

    // --- KEYBOARDS DEFINITION ---
    
    // Inline keyboard for the initial /start message
    initialKeyboard := tgbotapi.NewInlineKeyboardMarkup(
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData("▶️ Papar Menu Utama", "show_main_menu"),
        ),
    )

    // Reply Keyboard for quick links (Option 1)
    mainMenuKeyboard := tgbotapi.NewReplyKeyboard(
        tgbotapi.NewKeyboardButtonRow(
            tgbotapi.NewKeyboardButton("🪙 Claim WorldCoin"),
            tgbotapi.NewKeyboardButton("🛅 Wallet HATA"),
        ),
        tgbotapi.NewKeyboardButtonRow(
            tgbotapi.NewKeyboardButton("🏧 Cashout ke Bank"),
            tgbotapi.NewKeyboardButton("📢 Channel"),
            tgbotapi.NewKeyboardButton("👨‍💻 Admin"),
        ),
        tgbotapi.NewKeyboardButtonRow(
            tgbotapi.NewKeyboardButton("🔙 Kembali Menu Utama"), // User requested button
            tgbotapi.NewKeyboardButton("🔄 Reset (Padam Mesej)"), // User requested button
        ),
    )

    // --- BOT UPDATE LOOP ---
    u := tgbotapi.NewUpdate(0)
    u.Timeout = 60
    updates := bot.GetUpdatesChan(u)

    // In-memory store for message IDs to delete
    var messageIDsToDelete = make(map[int64][]int)

    for update := range updates {
        // Handle inline button callbacks
        if update.CallbackQuery != nil {
            bot.Request(tgbotapi.NewCallback(update.CallbackQuery.ID, "")) // Acknowledge callback

            if update.CallbackQuery.Data == "show_main_menu" {
                chatID := update.CallbackQuery.Message.Chat.ID
                
                text := "Sila tekan *_Butang Aksi Pantas_* di bawah untuk pautan link pantas, atau taip teks `/claim` `/wallet` `/cashout` untuk langkah panduan bergambar\\."
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
            msg := tgbotapi.NewMessage(chatID, "👋 Selamat Datang ke 🤖*_`CRYTORIAN TELEBOT`_*\\! Tekan butang ▶️ di bawah untuk memaparkan menu\\.")
            msg.ParseMode = tgbotapi.ModeMarkdownV2
            msg.ReplyMarkup = initialKeyboard
            sentMsg, _ := bot.Send(msg)
            messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)

        // --- OPTION 1: REPLY KEYBOARD FOR QUICK LINKS ---
        case "💸 Claim WorldCoin":
            text := "Untuk mendaftar *WorldCoin*, sila guna pautan di bawah:\n\n🔗 [Daftar WorldCoin Di Sini](https://worldcoin.org/join/4RH0OTE)\n\nUntuk panduan penuh bergambar, taip: `/claim`"
            msg := tgbotapi.NewMessage(chatID, text)
            msg.ParseMode = tgbotapi.ModeMarkdownV2
            sentMsg, _ := bot.Send(msg)
            messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
            
        case "🛅 Wallet HATA":
            text := "Untuk mendaftar *HATA Wallet*, sila guna pautan di bawah:\n\n🔗 [Daftar HATA Di Sini](https://hata.io/signup?ref=HDX8778)\n\nUntuk panduan penuh bergambar, taip: `/wallet`"
            msg := tgbotapi.NewMessage(chatID, text)
            msg.ParseMode = tgbotapi.ModeMarkdownV2
            sentMsg, _ := bot.Send(msg)
            messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)

        case "🏧 Cashout ke Bank":
            text := "Panduan untuk menjual WorldCoin dan mengeluarkannya ke akaun bank anda boleh didapati melalui arahan di bawah\\.\n\ntaip: `/cashout`"
            msg := tgbotapi.NewMessage(chatID, text)
            msg.ParseMode = tgbotapi.ModeMarkdownV2
            sentMsg, _ := bot.Send(msg)
            messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)

        case "📢 Channel":
            text := "Sertai saluran Telegram kami untuk info dan maklumat lanjutan\\.\n\n🔗 [Join Channel](https://t.me/cucikripto)"
            msg := tgbotapi.NewMessage(chatID, text)
            msg.ParseMode = tgbotapi.ModeMarkdownV2
            sentMsg, _ := bot.Send(msg)
            messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)
        
        case "👨‍💻 Admin":
            text := "🆘 *_Perlukan bantuan lanjut_*\\? 📞Hubungi admin secara terus 🕣 Waktu urusan *setiap hari*\\.\n\n🔗 [Hubungi Admin](https://t.me/johansetia)"
            msg := tgbotapi.NewMessage(chatID, text)
            msg.ParseMode = tgbotapi.ModeMarkdownV2
            sentMsg, _ := bot.Send(msg)
            messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)

        // --- OPTION 2: COMMANDS FOR DETAILED GUIDES ---
        case "/claim":
            sendDetailedGuide(bot, chatID, guides["worldcoin_registration_guide"])
            
        case "/wallet", "/wallet hata": // Added alias
            sendDetailedGuide(bot, chatID, guides["hata_setup_guide"])
            
        case "/cashout":
            sendDetailedGuide(bot, chatID, guides["cashout_guide"])

        case "/reset mesej", "🔄 Reset (Padam Mesej)":
            for _, id := range messageIDsToDelete[chatID] {
                bot.Request(tgbotapi.NewDeleteMessage(chatID, id))
            }
            messageIDsToDelete[chatID] = nil // Clear the slice

            // Send a fresh start message
            msg := tgbotapi.NewMessage(chatID, "🔄 Sesi telah direset semula\\. Tekan butang di bawah untuk mula sekali lagi\\.")
            msg.ParseMode = tgbotapi.ModeMarkdownV2
            msg.ReplyMarkup = initialKeyboard
            bot.Send(msg)

        default:
            msg := tgbotapi.NewMessage(chatID, "❌ Arahan tidak dikenali\\ ⚠️ Sila guna butang atau taip teks arahan yang sah,   /claim /wallet /cashout\\.")
            sentMsg, _ := bot.Send(msg)
            messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], sentMsg.MessageID)

        }
    }
}
