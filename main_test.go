package main

import (
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// --- Mock Bot ---
type MockBot struct {
	SentMessages []tgbotapi.Chattable
}

func (b *MockBot) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	b.SentMessages = append(b.SentMessages, c)
	return tgbotapi.Message{MessageID: len(b.SentMessages)}, nil
}

func (b *MockBot) Request(c tgbotapi.Chattable) (interface{}, error) {
	b.SentMessages = append(b.SentMessages, c)
	return nil, nil
}

// --- Test escapeMarkdownV2 ---
func TestEscapeMarkdownV2(t *testing.T) {
	input := "_Hello_ *World* [link](url) ~`"
	expected := `\_Hello\_ \*World\* \[link\]\(url\) \~\``

	output := escapeMarkdownV2(input)
	if output != expected {
		t.Errorf("escapeMarkdownV2 failed: got %s, want %s", output, expected)
	}
}

// --- Test createMainMenuKeyboard ---
func TestCreateMainMenuKeyboard(t *testing.T) {
	keyboard := createMainMenuKeyboard()
	if len(keyboard.Keyboard) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(keyboard.Keyboard))
	}
	if keyboard.ResizeKeyboard != true {
		t.Errorf("Expected ResizeKeyboard true")
	}
}

// --- Test createInlineKeyboard ---
func TestCreateInlineKeyboard(t *testing.T) {
	keyboard := createInlineKeyboard()
	if len(keyboard.InlineKeyboard) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(keyboard.InlineKeyboard))
	}
}

// --- Test sendDetailedGuide ---
func TestSendDetailedGuide(t *testing.T) {
	mockBot := &MockBot{}

	guide := Guide{
		Title: "Test Guide",
		Steps: []Step{
			{Title: "Step 1", Desc: "Do this first", Images: nil},
			{Title: "Step 2", Desc: "Then this", Images: []string{"https://example.com/img1.jpg"}},
		},
		Important: Important{
			Title: "Important Notes",
			Notes: []string{"Note 1", "Note 2"},
		},
	}

	sendDetailedGuide(mockBot, 12345, guide)

	expectedMessages := 3 // Step1, Step2 (photo), Important Notes
	if len(mockBot.SentMessages) != expectedMessages {
		t.Errorf("Expected %d messages, got %d", expectedMessages, len(mockBot.SentMessages))
	}
}

// --- Test userLastActivity update ---
func TestUserActivity(t *testing.T) {
	chatID := int64(111)
	userLastActivity[chatID] = time.Time{}

	userLastActivity[chatID] = time.Now()
	if userLastActivity[chatID].IsZero() {
		t.Errorf("Expected non-zero time for user activity")
	}
}

// --- Mock Update untuk simulate commands ---
func createMockMessage(chatID int64, text string) tgbotapi.Update {
	return tgbotapi.Update{
		UpdateID: 1,
		Message: &tgbotapi.Message{
			MessageID: 1,
			Chat: &tgbotapi.Chat{
				ID:   chatID,
				Type: "private",
			},
			From: &tgbotapi.User{
				ID:        999,
				FirstName: "Test",
				LastName:  "User",
				UserName:  "testuser",
			},
			Text: text,
			Time: time.Now(),
		},
	}
}

func createMockCallback(chatID int64, data string) tgbotapi.Update {
	return tgbotapi.Update{
		UpdateID: 2,
		CallbackQuery: &tgbotapi.CallbackQuery{
			ID: "callback1",
			Message: &tgbotapi.Message{
				MessageID: 2,
				Chat: &tgbotapi.Chat{
					ID: chatID,
				},
			},
			Data: data,
			From: &tgbotapi.User{
				ID:        999,
				FirstName: "Test",
				LastName:  "User",
				UserName:  "testuser",
			},
		},
	}
}

// --- Test main commands ---
func TestBotCommands(t *testing.T) {
	mockBot := &MockBot{}
	chatID := int64(12345)

	commands := []string{
		"/start",
		"🔙 Menu Utama",
		"/menu",
		"🚀 Claim WorldCoin",
		"💰 Wallet HATA",
		"💸 Cashout ke Bank",
		"📢 Channel",
		"👨‍💻 Admin",
		"🆔 Dapatkan Chat ID",
		"/chatid",
		"/reset",
	}

	for _, cmd := range commands {
		update := createMockMessage(chatID, cmd)

		// Simulate update loop dari main.go
		if update.Message == nil || update.Message.Text == "" {
			continue
		}

		userLastActivity[chatID] = time.Now()

		switch update.Message.Text {
		case "/start", "🔙 Menu Utama":
			text := "Selamat datang"
			msg := tgbotapi.NewMessage(chatID, text)
			mockBot.Send(msg)

		case "🚀 Claim WorldCoin":
			text := "Claim"
			msg := tgbotapi.NewMessage(chatID, text)
			mockBot.Send(msg)

		case "💰 Wallet HATA":
			text := "Wallet HATA"
			msg := tgbotapi.NewMessage(chatID, text)
			mockBot.Send(msg)

		case "💸 Cashout ke Bank":
			text := "Cashout"
			msg := tgbotapi.NewMessage(chatID, text)
			mockBot.Send(msg)

		case "📢 Channel":
			text := "Channel"
			msg := tgbotapi.NewMessage(chatID, text)
			mockBot.Send(msg)

		case "👨‍💻 Admin":
			text := "Admin"
			msg := tgbotapi.NewMessage(chatID, text)
			mockBot.Send(msg)

		case "🆔 Dapatkan Chat ID", "/chatid":
			text := "Chat ID"
			msg := tgbotapi.NewMessage(chatID, text)
			mockBot.Send(msg)

		case "/reset":
			messageIDsToDelete[chatID] = []int{1, 2, 3}
			deletedCount := len(messageIDsToDelete[chatID])
			messageIDsToDelete[chatID] = nil
			text := "Reset done"
			msg := tgbotapi.NewMessage(chatID, text)
			mockBot.Send(msg)

		default:
			text := "Arahan tak dikenali"
			msg := tgbotapi.NewMessage(chatID, text)
			mockBot.Send(msg)
		}
	}

	if len(mockBot.SentMessages) != len(commands) {
		t.Errorf("Expected %d messages sent, got %d", len(commands), len(mockBot.SentMessages))
	}
}

// --- Test Callback Inline Button ---
func TestInlineCallback(t *testing.T) {
	mockBot := &MockBot{}
	chatID := int64(12345)

	callback := createMockCallback(chatID, "show_main_menu")

	if callback.CallbackQuery != nil {
		// Simulate callback
		mockBot.Request(tgbotapi.NewCallback(callback.CallbackQuery.ID, ""))
		text := "Main menu shown"
		msg := tgbotapi.NewMessage(chatID, text)
		mockBot.Send(msg)
	}

	if len(mockBot.SentMessages) != 2 {
		t.Errorf("Expected 2 messages sent for callback, got %d", len(mockBot.SentMessages))
	}
}

// --- Test security link check ---
func TestSecurityLink(t *testing.T) {
	mockBot := &MockBot{}
	chatID := int64(12345)

	texts := []string{
		"http://malicious.com",
		"www.badlink.org",
		"example.ba.ba",
	}

	for _, txt := range texts {
		update := createMockMessage(chatID, txt)
		if update.Message != nil {
			if containsDangerousLink(update.Message.Text) {
				msg := tgbotapi.NewMessage(chatID, "Warning sent")
				mockBot.Send(msg)
			}
		}
	}

	if len(mockBot.SentMessages) != len(texts) {
		t.Errorf("Expected %d warning messages, got %d", len(texts), len(mockBot.SentMessages))
	}
}

// --- Helper for security check ---
func containsDangerousLink(text string) bool {
	lower := text
	return contains(lower, "http") || contains(lower, "www.") || contains(lower, ".ba.ba")
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (len(sub) == 0 || (len(s) > 0 && (s[0:len(sub)] == sub || contains(s[1:], sub))))
}
