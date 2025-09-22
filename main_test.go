package main

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TestEscapeMarkdownV2 - Test fungsi escapeMarkdownV2
func TestEscapeMarkdownV2(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello_world", `hello\_world`},
		{"test[text]", `test\[text\]`},
		{"a*b+c-d", `a\*b\+c\-d`},
		{"normal text", "normal text"},
		{"", ""},
	}

	for _, test := range tests {
		result := escapeMarkdownV2(test.input)
		if result != test.expected {
			t.Errorf("escapeMarkdownV2(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

// TestCreateMainMenuKeyboard - Test fungsi createMainMenuKeyboard
func TestCreateMainMenuKeyboard(t *testing.T) {
	keyboard := createMainMenuKeyboard()

	// Test jumlah baris
	if len(keyboard.Keyboard) != 3 {
		t.Errorf("Expected 3 rows in keyboard, got %d", len(keyboard.Keyboard))
	}

	// Test jumlah tombol per baris
	expectedButtonsPerRow := []int{2, 3, 3}
	for i, row := range keyboard.Keyboard {
		if len(row) != expectedButtonsPerRow[i] {
			t.Errorf("Row %d: expected %d buttons, got %d", i, expectedButtonsPerRow[i], len(row))
		}
	}

	// Test properti keyboard
	if !keyboard.ResizeKeyboard {
		t.Error("Expected ResizeKeyboard to be true")
	}
	if keyboard.OneTimeKeyboard {
		t.Error("Expected OneTimeKeyboard to be false")
	}
}

// TestCreateInlineKeyboard - Test fungsi createInlineKeyboard
func TestCreateInlineKeyboard(t *testing.T) {
	keyboard := createInlineKeyboard()

	// Test jumlah baris
	if len(keyboard.InlineKeyboard) != 2 {
		t.Errorf("Expected 2 rows in inline keyboard, got %d", len(keyboard.InlineKeyboard))
	}

	// Test jumlah tombol per baris
	if len(keyboard.InlineKeyboard[0]) != 3 {
		t.Errorf("First row: expected 3 buttons, got %d", len(keyboard.InlineKeyboard[0]))
	}
	if len(keyboard.InlineKeyboard[1]) != 2 {
		t.Errorf("Second row: expected 2 buttons, got %d", len(keyboard.InlineKeyboard[1]))
	}

	// Test teks tombol
	expectedFirstRowTexts := []string{"🚀 Claim", "💰 Wallet", "💸 Cashout"}
	for i, button := range keyboard.InlineKeyboard[0] {
		if button.Text != expectedFirstRowTexts[i] {
			t.Errorf("Button %d text: expected %s, got %s", i, expectedFirstRowTexts[i], button.Text)
		}
	}
}

// TestLoadJSONData - Test loading data dari markdown.json
func TestLoadJSONData(t *testing.T) {
	// Buat file JSON test sementara
	testJSON := `{
		"worldcoin_registration_guide": {
			"title": "Test Guide",
			"steps": [
				{
					"title": "Step 1",
					"desc": "Description 1",
					"images": ["image1.jpg"]
				}
			],
			"important": {
				"title": "Important Notes",
				"notes": ["Note 1", "Note 2"]
			}
		}
	}`

	// Tulis ke file sementara
	tmpfile, err := os.CreateTemp("", "test_markdown.json")
	if err != nil {
		t.Fatal("Error creating temp file:", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(testJSON)); err != nil {
		t.Fatal("Error writing to temp file:", err)
	}
	tmpfile.Close()

	// Load data dari file
	jsonData, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatal("Error reading test JSON file:", err)
	}

	var guides map[string]Guide
	if err := json.Unmarshal(jsonData, &guides); err != nil {
		t.Fatal("JSON parsing error:", err)
	}

	// Test data yang diload
	if len(guides) != 1 {
		t.Errorf("Expected 1 guide, got %d", len(guides))
	}

	guide := guides["worldcoin_registration_guide"]
	if guide.Title != "Test Guide" {
		t.Errorf("Expected title 'Test Guide', got '%s'", guide.Title)
	}

	if len(guide.Steps) != 1 {
		t.Errorf("Expected 1 step, got %d", len(guide.Steps))
	}

	if guide.Steps[0].Title != "Step 1" {
		t.Errorf("Expected step title 'Step 1', got '%s'", guide.Steps[0].Title)
	}

	if len(guide.Important.Notes) != 2 {
		t.Errorf("Expected 2 important notes, got %d", len(guide.Important.Notes))
	}
}

// TestUserLastActivity - Test manajemen aktivitas user
func TestUserLastActivity(t *testing.T) {
	// Reset global variable untuk test
	userLastActivity = make(map[int64]time.Time)

	// Test menambah aktivitas baru
	testChatID := int64(12345)
	userLastActivity[testChatID] = time.Now()

	// Test aktivitas ada
	if _, exists := userLastActivity[testChatID]; !exists {
		t.Error("User activity was not recorded")
	}

	// Test aktivitas terbaru
	oldTime := time.Now().Add(-10 * time.Minute)
	userLastActivity[testChatID] = oldTime

	if userLastActivity[testChatID] != oldTime {
		t.Error("User activity time was not set correctly")
	}
}

// TestMessageIDsToDelete - Test manajemen message IDs
func TestMessageIDsToDelete(t *testing.T) {
	// Reset global variable untuk test
	messageIDsToDelete = make(map[int64][]int)

	// Test menambah message IDs
	testChatID := int64(12345)
	messageIDsToDelete[testChatID] = []int{1, 2, 3}

	// Test message IDs ada
	if ids, exists := messageIDsToDelete[testChatID]; !exists || len(ids) != 3 {
		t.Error("Message IDs were not recorded correctly")
	}

	// Test menambah message ID baru
	messageIDsToDelete[testChatID] = append(messageIDsToDelete[testChatID], 4)
	if len(messageIDsToDelete[testChatID]) != 4 {
		t.Error("Message ID was not appended correctly")
	}
}

// MockBot untuk testing
type MockBot struct {
	Messages []tgbotapi.MessageConfig
	Photos   []tgbotapi.PhotoConfig
}

func (m *MockBot) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	switch msg := c.(type) {
	case tgbotapi.MessageConfig:
		m.Messages = append(m.Messages, msg)
	case tgbotapi.PhotoConfig:
		m.Photos = append(m.Photos, msg)
	}
	return tgbotapi.Message{}, nil
}

// TestSendDetailedGuide - Test fungsi sendDetailedGuide dengan mock bot
func TestSendDetailedGuide(t *testing.T) {
	mockBot := &MockBot{}
	testChatID := int64(12345)

	testGuide := Guide{
		Title: "Test Guide",
		Steps: []Step{
			{
				Title:  "Step 1",
				Desc:   "Description 1",
				Images: []string{"image1.jpg"},
			},
			{
				Title:  "Step 2",
				Desc:   "Description 2",
				Images: []string{},
			},
		},
		Important: Important{
			Title: "Important Notes",
			Notes: []string{"Note 1", "Note 2"},
		},
	}

	// Reset global variable untuk test
	messageIDsToDelete = make(map[int64][]int)

	sendDetailedGuide(mockBot, testChatID, testGuide)

	// Test jumlah message yang dikirim
	if len(mockBot.Messages) != 3 { // Title + Step 2 + Important
		t.Errorf("Expected 3 messages, got %d", len(mockBot.Messages))
	}

	// Test jumlah photo yang dikirim
	if len(mockBot.Photos) != 1 { // Step 1 dengan gambar
		t.Errorf("Expected 1 photo, got %d", len(mockBot.Photos))
	}

	// Test message IDs disimpan
	if len(messageIDsToDelete[testChatID]) != 4 { // 3 messages + 1 photo
		t.Errorf("Expected 4 message IDs stored, got %d", len(messageIDsToDelete[testChatID]))
	}
}

// TestMain - Test entry point (hanya test bahwa tidak panic)
func TestMain(m *testing.M) {
	// Simulate environment variable for token
	os.Setenv("TELEGRAM_BOT_TOKEN", "test_token")
	
	// Run tests
	code := m.Run()
	
	// Clean up
	os.Unsetenv("TELEGRAM_BOT_TOKEN")
	os.Exit(code)
}
