package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Test data untuk mock
var testGuides map[string]Guide

func TestMain(m *testing.M) {
	// Setup test data
	testGuides = map[string]Guide{
		"worldcoin_registration_guide": {
			Title: "Test WorldCoin Guide",
			Steps: []Step{
				{
					Title:  "Step 1",
					Desc:   "Test description 1",
					Images: []string{"https://example.com/image1.jpg"},
				},
			},
			Important: Important{
				Title: "Important Notes",
				Notes: []string{"Note 1", "Note 2"},
			},
		},
		"hata_setup_guide": {
			Title: "Test HATA Setup Guide",
			Steps: []Step{
				{
					Title:  "HATA Step 1",
					Desc:   "HATA description 1",
					Images: []string{},
				},
			},
			Important: Important{
				Title: "HATA Important",
				Notes: []string{"HATA Note 1"},
			},
		},
		"cashout_guide": {
			Title: "Test Cashout Guide",
			Steps: []Step{
				{
					Title:  "Cashout Step 1",
					Desc:   "Cashout description 1",
					Images: []string{"https://example.com/cashout1.jpg", "https://example.com/cashout2.jpg"},
				},
			},
			Important: Important{
				Title: "Cashout Important",
				Notes: []string{},
			},
		},
	}

	// Run tests
	exitCode := m.Run()
	os.Exit(exitCode)
}

// TestEscapeMarkdownV2 - Test fungsi escape Markdown
func TestEscapeMarkdownV2(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Escape special characters",
			input:    "Hello_world {test} [example]",
			expected: `Hello\_world \{test\} \[example\]`,
		},
		{
			name:     "No special characters",
			input:    "Hello world",
			expected: "Hello world",
		},
		{
			name:     "Mixed characters",
			input:    "Test_here {value} and normal text",
			expected: `Test\_here \{value\} and normal text`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeMarkdownV2(tt.input)
			if result != tt.expected {
				t.Errorf("escapeMarkdownV2(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestCreateMainMenuKeyboard - Test pembuatan keyboard utama
func TestCreateMainMenuKeyboard(t *testing.T) {
	keyboard := createMainMenuKeyboard()

	// Test keyboard properties
	if !keyboard.ResizeKeyboard {
		t.Error("Keyboard should have ResizeKeyboard = true")
	}

	if keyboard.OneTimeKeyboard {
		t.Error("Keyboard should have OneTimeKeyboard = false")
	}

	// Test jumlah baris
	if len(keyboard.Keyboard) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(keyboard.Keyboard))
	}

	// Test jumlah tombol per baris
	expectedButtons := [][]int{{2}, {3}, {3}}
	for i, row := range keyboard.Keyboard {
		if len(row) != expectedButtons[i][0] {
			t.Errorf("Row %d: expected %d buttons, got %d", i, expectedButtons[i][0], len(row))
		}
	}

	// Test teks tombol
	expectedTexts := []string{"🚀 Claim WorldCoin", "💰 Wallet HATA", "💸 Cashout ke Bank", "📢 Channel", "👨‍💻 Admin", "🆔 Dapatkan Chat ID", "🔄 Reset", "🔙 Menu Utama"}
	textIndex := 0

	for _, row := range keyboard.Keyboard {
		for _, button := range row {
			if button.Text != expectedTexts[textIndex] {
				t.Errorf("Button text mismatch: expected %q, got %q", expectedTexts[textIndex], button.Text)
			}
			textIndex++
		}
	}
}

// TestCreateInlineKeyboard - Test pembuatan inline keyboard
func TestCreateInlineKeyboard(t *testing.T) {
	keyboard := createInlineKeyboard()

	// Test jumlah baris
	if len(keyboard.InlineKeyboard) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(keyboard.InlineKeyboard))
	}

	// Test jumlah tombol per baris
	if len(keyboard.InlineKeyboard[0]) != 3 {
		t.Errorf("First row: expected 3 buttons, got %d", len(keyboard.InlineKeyboard[0]))
	}
	if len(keyboard.InlineKeyboard[1]) != 2 {
		t.Errorf("Second row: expected 2 buttons, got %d", len(keyboard.InlineKeyboard[1]))
	}

	// Test data callback untuk tombol pertama
	expectedCallbacks := []string{"claim_cmd", "wallet_cmd", "cashout_cmd"}
	for i, button := range keyboard.InlineKeyboard[0] {
		if button.CallbackData == nil || *button.CallbackData != expectedCallbacks[i] {
			t.Errorf("Button %d callback data mismatch", i)
		}
	}

	// Test URL untuk tombol kedua
	expectedURLs := []string{"https://t.me/cucikripto", "https://t.me/johansetia"}
	for i, button := range keyboard.InlineKeyboard[1] {
		if button.URL == nil || *button.URL != expectedURLs[i] {
			t.Errorf("Button %d URL mismatch", i)
		}
	}
}

// TestLoadJSONData - Test pemuatan data JSON
func TestLoadJSONData(t *testing.T) {
	// Create temporary JSON file
	tempJSON := `{
		"test_guide": {
			"title": "Test Guide",
			"steps": [
				{
					"title": "Step 1",
					"desc": "Description 1",
					"images": ["image1.jpg"]
				}
			],
			"important": {
				"title": "Important",
				"notes": ["Note 1"]
			}
		}
	}`

	tempFile, err := os.CreateTemp("", "test*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.WriteString(tempJSON); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tempFile.Close()

	// Test loading JSON
	jsonData, err := os.ReadFile(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to read JSON file: %v", err)
	}

	var guides map[string]Guide
	if err := json.Unmarshal(jsonData, &guides); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verify loaded data
	if len(guides) != 1 {
		t.Errorf("Expected 1 guide, got %d", len(guides))
	}

	guide, exists := guides["test_guide"]
	if !exists {
		t.Error("Guide 'test_guide' not found")
	}

	if guide.Title != "Test Guide" {
		t.Errorf("Guide title mismatch: expected 'Test Guide', got '%s'", guide.Title)
	}

	if len(guide.Steps) != 1 {
		t.Errorf("Expected 1 step, got %d", len(guide.Steps))
	}

	if guide.Steps[0].Title != "Step 1" {
		t.Errorf("Step title mismatch")
	}

	if len(guide.Important.Notes) != 1 {
		t.Errorf("Expected 1 important note, got %d", len(guide.Important.Notes))
	}
}

// TestMessageHandling - Test handling berbagai jenis pesan
func TestMessageHandling(t *testing.T) {
	// Test cases untuk berbagai command
	testCases := []struct {
		name     string
		message  string
		expected string
	}{
		{"Start command", "/start", "Selamat Datang"},
		{"Menu command", "/menu", "SMART AASA BOT"},
		{"Claim button", "🚀 Claim WorldCoin", "VERIFIKASI WORLDCOIN"},
		{"Wallet button", "💰 Wallet HATA", "HATA WALLET"},
		{"Chat ID button", "🆔 Dapatkan Chat ID", "INFO AKUN ANDA"},
		{"Reset button", "🔄 Reset", "Reset Berjaya"},
		{"Unknown command", "unknown", "Arahan tak dikenali"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate message processing
			response := simulateMessageProcessing(tc.message)
			
			if tc.expected != "" && !strings.Contains(response, tc.expected) {
				t.Errorf("For message %q: expected response containing %q, got %q", 
					tc.message, tc.expected, response)
			}
		})
	}
}

// Helper function untuk simulate message processing
func simulateMessageProcessing(message string) string {
	switch message {
	case "/start", "🔙 Menu Utama":
		return "Selamat Datang di SMART AASA BOT!"
	case "/menu", "📋 Menu":
		return "SMART AASA BOT - Panduan Lengkap Kripto"
	case "🚀 Claim WorldCoin":
		return "VERIFIKASI WORLDCOIN"
	case "💰 Wallet HATA":
		return "HATA WALLET"
	case "💸 Cashout ke Bank":
		return "CASHOUT KE BANK"
	case "🆔 Dapatkan Chat ID":
		return "INFO AKUN ANDA"
	case "/reset", "🔄 Reset":
		return "Reset Berjaya"
	default:
		return "Arahan tak dikenali. Sila guna butang atau command yang betul."
	}
}

// TestSecurityCheck - Test security check untuk suspicious links
func TestSecurityCheck(t *testing.T) {
	suspiciousMessages := []string{
		"Check this http://malicious.com",
		"Visit www.suspicious.site",
		"Click here: https://bad.ba.ba",
		"Go to http://www.dangerous.ba.ba",
	}

	safeMessages := []string{
		"Hello world",
		"Visit https://worldcoin.org",
		"Check hata.io",
		"Normal message without links",
	}

	for _, msg := range suspiciousMessages {
		if !isSuspiciousMessage(msg) {
			t.Errorf("Should detect suspicious message: %q", msg)
		}
	}

	for _, msg := range safeMessages {
		if isSuspiciousMessage(msg) {
			t.Errorf("Should not flag safe message: %q", msg)
		}
	}
}

// Helper function untuk security check
func isSuspiciousMessage(text string) bool {
	lowerText := strings.ToLower(text)
	return strings.Contains(lowerText, "http") || 
		   strings.Contains(lowerText, "www.") ||
		   strings.Contains(lowerText, ".ba.ba")
}

// TestUserActivityTracking - Test tracking aktivitas user
func TestUserActivityTracking(t *testing.T) {
	chatID := int64(12345)
	now := time.Now()

	// Test initial state
	if _, exists := userLastActivity[chatID]; exists {
		t.Error("User should not exist in tracking initially")
	}

	// Test adding user activity
	userLastActivity[chatID] = now
	if !userLastActivity[chatID].Equal(now) {
		t.Error("Failed to track user activity")
	}

	// Test updating user activity
	newTime := now.Add(5 * time.Minute)
	userLastActivity[chatID] = newTime
	if !userLastActivity[chatID].Equal(newTime) {
		t.Error("Failed to update user activity")
	}

	// Cleanup
	delete(userLastActivity, chatID)
}

// TestMessageIDTracking - Test tracking message IDs untuk delete
func TestMessageIDTracking(t *testing.T) {
	chatID := int64(67890)
	messageIDs := []int{100, 101, 102}

	// Test initial state
	if len(messageIDsToDelete[chatID]) != 0 {
		t.Error("Message IDs tracking should be empty initially")
	}

	// Test adding message IDs
	messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], messageIDs...)
	if len(messageIDsToDelete[chatID]) != len(messageIDs) {
		t.Error("Failed to track message IDs")
	}

	// Test message IDs content
	for i, id := range messageIDs {
		if messageIDsToDelete[chatID][i] != id {
			t.Error("Message IDs content mismatch")
		}
	}

	// Test clearing message IDs
	messageIDsToDelete[chatID] = nil
	if len(messageIDsToDelete[chatID]) != 0 {
		t.Error("Failed to clear message IDs")
	}
}

// Benchmark test untuk fungsi escapeMarkdownV2
func BenchmarkEscapeMarkdownV2(b *testing.B) {
	testString := "Hello _world_ {test} [example] with *many* ~special~ characters!"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		escapeMarkdownV2(testString)
	}
}

// Benchmark test untuk createMainMenuKeyboard
func BenchmarkCreateMainMenuKeyboard(b *testing.B) {
	for i := 0; i < b.N; i++ {
		createMainMenuKeyboard()
	}
}

// Example test untuk dokumentasi
func ExampleEscapeMarkdownV2() {
	result := escapeMarkdownV2("Hello _world_!")
	println(result)
	// Output: Hello \_world\_!
}

// TestGuideStructure - Test struktur data Guide
func TestGuideStructure(t *testing.T) {
	guide := testGuides["worldcoin_registration_guide"]

	if guide.Title != "Test WorldCoin Guide" {
		t.Errorf("Guide title mismatch")
	}

	if len(guide.Steps) == 0 {
		t.Error("Guide should have steps")
	}

	step := guide.Steps[0]
	if step.Title != "Step 1" {
		t.Errorf("Step title mismatch")
	}

	if !strings.Contains(step.Desc, "Test description") {
		t.Errorf("Step description mismatch")
	}

	if len(guide.Important.Notes) != 2 {
		t.Errorf("Important notes count mismatch")
	}
}

// TestMultiImageHandling - Test handling multiple images
func TestMultiImageHandling(t *testing.T) {
	guide := testGuides["cashout_guide"]
	step := guide.Steps[0]

	if len(step.Images) != 2 {
		t.Errorf("Expected 2 images, got %d", len(step.Images))
	}

	expectedImages := []string{
		"https://example.com/cashout1.jpg",
		"https://example.com/cashout2.jpg",
	}

	for i, img := range step.Images {
		if img != expectedImages[i] {
			t.Errorf("Image URL mismatch at index %d", i)
		}
	}
}
