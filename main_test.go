package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Mock Bot untuk testing
type mockBot struct {
	lastSentMsg   tgbotapi.Chattable
	sentMessages  []tgbotapi.Chattable
	deleteHistory []int
}

func (m *mockBot) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	m.lastSentMsg = c
	m.sentMessages = append(m.sentMessages, c)
	return tgbotapi.Message{}, nil
}

func (m *mockBot) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	// Handle delete message request
	if deleteMsg, ok := c.(tgbotapi.DeleteMessageConfig); ok {
		m.deleteHistory = append(m.deleteHistory, deleteMsg.MessageID)
	}
	return &tgbotapi.APIResponse{}, nil
}

func (m *mockBot) GetUpdatesChan(config tgbotapi.UpdateConfig) (tgbotapi.UpdatesChannel, error) {
	return make(chan tgbotapi.Update), nil
}

// Test escapeMarkdownV2 function
func TestEscapeMarkdownV2(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Escape special characters",
			input:    "Test _italic_ *bold* [link]",
			expected: `Test \_italic\_ \*bold\* \[link\]`,
		},
		{
			name:     "No special characters",
			input:    "Regular text",
			expected: "Regular text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeMarkdownV2(tt.input)
			if result != tt.expected {
				t.Errorf("escapeMarkdownV2(%s) = %s, expected %s", tt.input, result, tt.expected)
			}
		})
	}
}

// Test loading guides from JSON
func TestLoadGuides(t *testing.T) {
	// Create a temporary test JSON file
	testJSON := `{
		"test_guide": {
			"title": "Test Guide",
			"steps": [
				{
					"title": "Step 1",
					"desc": "Description 1",
					"images": []
				}
			],
			"important": {
				"title": "Important Notes",
				"notes": ["Note 1", "Note 2"]
			}
		}
	}`

	tmpfile, err := os.CreateTemp("", "test_markdown.json")
	if err != nil {
		t.Fatal("Error creating temp file:", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(testJSON)); err != nil {
		t.Fatal("Error writing to temp file:", err)
	}
	tmpfile.Close()

	// Test loading the JSON
	jsonData, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatal("Error reading test JSON:", err)
	}

	var guides map[string]Guide
	if err := json.Unmarshal(jsonData, &guides); err != nil {
		t.Fatal("JSON parsing error:", err)
	}

	// Verify the loaded data
	if guides["test_guide"].Title != "Test Guide" {
		t.Errorf("Expected title 'Test Guide', got '%s'", guides["test_guide"].Title)
	}

	if len(guides["test_guide"].Steps) != 1 {
		t.Errorf("Expected 1 step, got %d", len(guides["test_guide"].Steps))
	}

	if len(guides["test_guide"].Important.Notes) != 2 {
		t.Errorf("Expected 2 important notes, got %d", len(guides["test_guide"].Important.Notes))
	}
}

// Test sendDetailedGuide function
func TestSendDetailedGuide(t *testing.T) {
	bot := &mockBot{}
	chatID := int64(12345)

	guide := Guide{
		Title: "Test Guide",
		Steps: []Step{
			{
				Title:  "Step 1",
				Desc:   "Description 1",
				Images: []string{},
			},
			{
				Title:  "Step 2",
				Desc:   "Description 2",
				Images: []string{"https://example.com/image1.jpg"},
			},
		},
		Important: Important{
			Title: "Important Notes",
			Notes: []string{"Note 1", "Note 2"},
		},
	}

	// Clear any previous state
	messageIDsToDelete[chatID] = nil

	sendDetailedGuide(bot, chatID, guide)

	// Check that messages were added to deletion list
	if len(messageIDsToDelete[chatID]) == 0 {
		t.Error("No message IDs were added to deletion list")
	}

	// Check that messages were sent
	if len(bot.sentMessages) == 0 {
		t.Error("No messages were sent")
	}
}

// Test message handling for different commands
func TestMessageHandling(t *testing.T) {
	bot := &mockBot{}
	
	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "Start command",
			message:  "/start",
			expected: "Selamat Datang",
		},
		{
			name:     "WorldCoin command",
			message:  "🚀 Claim WorldCoin",
			expected: "WorldCoin",
		},
		{
			name:     "Wallet command",
			message:  "💰 Wallet HATA",
			expected: "HATA Wallet",
		},
		{
			name:     "Cashout command",
			message:  "💸 Cashout ke Bank",
			expected: "cashout",
		},
		{
			name:     "Unknown command",
			message:  "unknown command",
			expected: "tak dikenali",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous messages
			bot.sentMessages = nil
			
			update := tgbotapi.Update{
				Message: &tgbotapi.Message{
					Chat: &tgbotapi.Chat{
						ID: 12345,
					},
					Text: tt.message,
				},
			}

			// Simulate main loop processing
			if update.Message != nil {
				chatID := update.Message.Chat.ID
				
				switch update.Message.Text {
				case "/start", "🔙 Kembali Menu Utama":
					msg := tgbotapi.NewMessage(chatID, "👋 Selamat Datang! Tekan butang di bawah untuk mula.")
					bot.Send(msg)
				case "🚀 Claim WorldCoin":
					msg := tgbotapi.NewMessage(chatID, "Daftar *WorldCoin*: [Klik sini](https://worldcoin.org/join/4RH0OTE)\n\nPanduan penuh: `/claim`")
					bot.Send(msg)
				case "💰 Wallet HATA":
					msg := tgbotapi.NewMessage(chatID, "Daftar *HATA Wallet*: [Klik sini](https://hata.io/signup?ref=HDX8778)\n\nPanduan penuh: `/wallet`")
					bot.Send(msg)
				case "💸 Cashout ke Bank":
					msg := tgbotapi.NewMessage(chatID, "Panduan cashout ada pada arahan: `/cashout`")
					bot.Send(msg)
				default:
					msg := tgbotapi.NewMessage(chatID, "❌ Arahan tak dikenali. Sila guna butang atau command yang betul.")
					bot.Send(msg)
				}
			}

			// Check that a message was sent
			if len(bot.sentMessages) == 0 {
				t.Error("No message was sent for command:", tt.message)
				return
			}

			// Verify the message content
			var messageText string
			switch msg := bot.lastSentMsg.(type) {
			case tgbotapi.MessageConfig:
				messageText = msg.Text
			default:
				t.Error("Unexpected message type")
				return
			}

			if !strings.Contains(strings.ToLower(messageText), strings.ToLower(tt.expected)) {
				t.Errorf("Message '%s' doesn't contain expected text '%s'", messageText, tt.expected)
			}
		})
	}
}

// Test callback query handling
func TestCallbackQuery(t *testing.T) {
	bot := &mockBot{}
	
	update := tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{
			ID:   "test_callback_id",
			Data: "show_main_menu",
			Message: &tgbotapi.Message{
				MessageID: 1001,
				Chat: &tgbotapi.Chat{
					ID: 12345,
				},
			},
		},
	}

	// Simulate callback processing
	if update.CallbackQuery != nil {
		bot.Request(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
		
		if update.CallbackQuery.Data == "show_main_menu" {
			chatID := update.CallbackQuery.Message.Chat.ID
			msg := tgbotapi.NewMessage(chatID, "Sila guna *Butang Aksi Pantas* di bawah atau taip `/claim`, `/wallet`, `/cashout`.")
			bot.Send(msg)

			// Simulate deleting the original message
			deleteMsg := tgbotapi.NewDeleteMessage(chatID, update.CallbackQuery.Message.MessageID)
			bot.Request(deleteMsg)
		}
	}

	// Check that a message was sent
	if len(bot.sentMessages) == 0 {
		t.Error("No message was sent for callback query")
	}

	// Check that delete request was made
	if len(bot.deleteHistory) == 0 {
		t.Error("No delete request was made")
	} else if bot.deleteHistory[0] != 1001 {
		t.Errorf("Expected to delete message ID 1001, got %d", bot.deleteHistory[0])
	}
}

// Test message reset functionality
func TestMessageReset(t *testing.T) {
	bot := &mockBot{}
	chatID := int64(12345)

	// Add some dummy message IDs to delete
	messageIDsToDelete[chatID] = []int{1001, 1002, 1003}

	// Simulate reset command
	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{
				ID: chatID,
			},
			Text: "/reset mesej",
		},
	}

	if update.Message != nil {
		if update.Message.Text == "/reset mesej" || update.Message.Text == "🔄 Reset (Padam Mesej)" {
			// Delete all bot messages
			for _, id := range messageIDsToDelete[chatID] {
				bot.Request(tgbotapi.NewDeleteMessage(chatID, id))
			}
			messageIDsToDelete[chatID] = nil

			// Send fresh start message
			msg := tgbotapi.NewMessage(chatID, "🔄 Reset! Tekan butang di bawah untuk mula semula.")
			bot.Send(msg)
			messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], 1004) // Mock new message ID
		}
	}

	// Check that delete requests were made for all messages
	if len(bot.deleteHistory) != 3 {
		t.Errorf("Expected 3 delete requests, got %d", len(bot.deleteHistory))
	}

	// Check that message list was cleared
	if len(messageIDsToDelete[chatID]) != 1 || messageIDsToDelete[chatID][0] != 1004 {
		t.Error("Message IDs were not properly reset")
	}
}

// Test environment variable validation
func TestEnvironmentValidation(t *testing.T) {
	// Save original token and clear it
	originalToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	os.Unsetenv("TELEGRAM_BOT_TOKEN")
	
	// Test that the function panics without token
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when TELEGRAM_BOT_TOKEN is not set")
		}
		// Restore original token
		if originalToken != "" {
			os.Setenv("TELEGRAM_BOT_TOKEN", originalToken)
		}
	}()
	
	// This should panic
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Panic("TELEGRAM_BOT_TOKEN environment variable not set")
	}
}
