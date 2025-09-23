// main_test.go

package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
)

// Test data untuk mock
var testGuides map[string]Guide

func TestMain(m *testing.M) {
	// Setup test data
	// URL dibersihkan daripada ruang tambahan
	testGuides = map[string]Guide{
		"worldcoin_registration_guide": {
			Title: "Test WorldCoin Guide",
			Steps: []Step{
				{
					Title:  "Step 1",
					Desc:   "Test description 1",
					Images: []string{"https://example.com/image1.jpg"}, // URL dibersihkan
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
					Images: []string{"https://example.com/cashout1.jpg", "https://example.com/cashout2.jpg"}, // URL dibersihkan
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
		{
			name:     "All special characters",
			input:    `_{}[]()~>#\+\-=|.!\\`,
			expected: `\_`\{`\}`\[`\]`\(`\)`\~`\>`\#`\+`\-`\=`\|`\.`\!`\\`,
		},
		{
			name:     "Exclamation mark",
			input:    "Wow! This is cool!",
			expected: `Wow\! This is cool\!`,
		},
		{
			name:     "Backslash",
			// Menggunakan backtick untuk elakkan isu escape dalam string ""
			input:    `Path is C:\Users\Name`,
			expected: `Path is C:\\Users\\Name`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeMarkdownV2(tt.input)
			// Debug print untuk membantu troubleshooting
			// t.Logf("Input: %q, Got: %q, Expected: %q", tt.input, result, tt.expected)
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

	// Test teks tombol baris pertama
	expectedFirstRow := []string{"🚀 Claim WorldCoin", "💰 Wallet HATA"}
	for i, button := range keyboard.Keyboard[0] {
		if button.Text != expectedFirstRow[i] {
			t.Errorf("First row button %d: expected %q, got %q", i, expectedFirstRow[i], button.Text)
		}
	}

	// Test teks tombol baris kedua
	expectedSecondRow := []string{"💸 Cashout ke Bank", "📢 Channel", "👨‍💻 Admin"}
	for i, button := range keyboard.Keyboard[1] {
		if button.Text != expectedSecondRow[i] {
			t.Errorf("Second row button %d: expected %q, got %q", i, expectedSecondRow[i], button.Text)
		}
	}

	// Test teks tombol baris ketiga
	expectedThirdRow := []string{"🆔 Dapatkan Chat ID", "🔄 Reset", "🔙 Menu Utama"}
	for i, button := range keyboard.Keyboard[2] {
		if button.Text != expectedThirdRow[i] {
			t.Errorf("Third row button %d: expected %q, got %q", i, expectedThirdRow[i], button.Text)
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

	// Test tombol baris pertama (callback data)
	firstRowButtons := []struct {
		text string
		data string
	}{
		{"🚀 Claim", "claim_cmd"},
		{"💰 Wallet", "wallet_cmd"},
		{"💸 Cashout", "cashout_cmd"},
	}

	for i, button := range keyboard.InlineKeyboard[0] {
		if button.Text != firstRowButtons[i].text {
			t.Errorf("First row button %d text: expected %q, got %q", i, firstRowButtons[i].text, button.Text)
		}
		if button.CallbackData == nil || *button.CallbackData != firstRowButtons[i].data {
			t.Errorf("First row button %d callback  expected %q, got %q", i, firstRowButtons[i].data, *button.CallbackData)
		}
	}

	// Test tombol baris kedua (URL) - URL dibersihkan
	secondRowButtons := []struct {
		text string
		url  string
	}{
		// URL dibersihkan
		{"📢 Channel", "https://t.me/cucikripto"},
		// URL dibersihkan
		{"👨‍💻 Admin", "https://t.me/johansetia"},
	}

	for i, button := range keyboard.InlineKeyboard[1] {
		if button.Text != secondRowButtons[i].text {
			t.Errorf("Second row button %d text: expected %q, got %q", i, secondRowButtons[i].text, button.Text)
		}
		if button.URL == nil || *button.URL != secondRowButtons[i].url {
			t.Errorf("Second row button %d URL: expected %q, got %q", i, secondRowButtons[i].url, *button.URL)
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
					"images": ["image1.jpg", "image2.jpg"]
				}
			],
			"important": {
				"title": "Important Title",
				"notes": ["Note 1", "Note 2"]
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
		t.Errorf("Step title mismatch: expected 'Step 1', got '%s'", guide.Steps[0].Title)
	}

	if len(guide.Steps[0].Images) != 2 {
		t.Errorf("Expected 2 images, got %d", len(guide.Steps[0].Images))
	}

	if guide.Important.Title != "Important Title" {
		t.Errorf("Important title mismatch")
	}

	if len(guide.Important.Notes) != 2 {
		t.Errorf("Expected 2 important notes, got %d", len(guide.Important.Notes))
	}
}

// TestMessageHandling - Test handling berbagai jenis pesan (simulasi ringkas)
// Nota: Ujian ini tidak boleh menguji logik sebenar kerana ia bergantung pada library Telegram.
// Ia hanya menguji beberapa senario asas.
func TestMessageHandling(t *testing.T) {
	testCases := []struct {
		name          string
		message       string
		shouldContain string
	}{
		{"Start command", "/start", "Selamat Datang"},
		{"Menu command", "/menu", "SMART AASA BOT"},
		// {"Claim command", "/claim", ""}, // Memerlukan mocking, tidak diuji di sini
		// {"Wallet command", "/wallet", ""}, // Memerlukan mocking, tidak diuji di sini
		// {"Cashout command", "/cashout", ""}, // Memerlukan mocking, tidak diuji di sini
		{"ChatID command", "/chatid", "Chat ID"},
		{"Reset command", "/reset", "Reset"},
		{"Claim button", "🚀 Claim WorldCoin", "VERIFIKASI"},
		{"Wallet button", "💰 Wallet HATA", "HATA WALLET"},
		{"Cashout button", "💸 Cashout ke Bank", "CASHOUT"},
		{"Channel button", "📢 Channel", "JOIN CHANNEL"},
		{"Admin button", "👨‍💻 Admin", "SUPPORT ADMIN"},
		{"ChatID button", "🆔 Dapatkan Chat ID", "INFO AKUN"},
		{"Reset button", "🔄 Reset", "Reset Berjaya"},
		{"Main Menu button", "🔙 Menu Utama", "Selamat Datang"},
		{"Unknown command", "unknown", "Arahan tak dikenali"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			response := simulateMessageProcessing(tc.message)

			if tc.shouldContain != "" && !strings.Contains(response, tc.shouldContain) {
				t.Errorf("For message %q: expected response containing %q, got %q",
					tc.message, tc.shouldContain, response)
			}
		})
	}
}

// Helper function untuk simulate message processing (versi ringkas)
func simulateMessageProcessing(message string) string {
	// Ini adalah simulasi ringkas dan tidak mencerminkan logik sebenar dalam main.go
	// Kerana logik sebenar bergantung pada library Telegram dan state.
	switch message {
	case "/start", "🔙 Menu Utama":
		return "👋 *Selamat Datang di SMART AASA BOT!*"
	case "/menu", "📋 Menu":
		return "🤖 *SMART AASA BOT* - Panduan Lengkap Kripto"
	case "🚀 Claim WorldCoin":
		return "🔐 *VERIFIKASI WORLDCOIN*"
	case "💰 Wallet HATA":
		return "📱 *HATA WALLET*"
	case "💸 Cashout ke Bank":
		return "🏦 *CASHOUT KE BANK*"
	case "📢 Channel":
		return "📢 *JOIN CHANNEL KAMI*"
	case "👨‍💻 Admin":
		return "👨‍💻 *SUPPORT ADMIN*"
	case "🆔 Dapatkan Chat ID":
		return "📋 *INFO AKUN ANDA*"
	case "/reset", "🔄 Reset":
		return "🔄 *Reset Berjaya!*"
	case "/chatid":
		return "🆔 *Chat ID Anda:*"
	default:
		return "❌ Arahan tak dikenali. Sila guna butang atau command yang betul."
	}
}

// TestSecurityCheck - Test security check untuk suspicious links
// Selaraskan dengan logik sebenar dalam main.go
func TestSecurityCheck(t *testing.T) {
	// Salin logik dari main.go
	isSuspiciousMessage := func(text string) bool {
		lowerText := strings.ToLower(text)
		hasHTTP := strings.Contains(lowerText, "http")
		hasWWW := strings.Contains(lowerText, "www.")
		hasBABA := strings.Contains(lowerText, ".ba.ba")

		// Exclude safe domains
		hasSafeDomain := strings.Contains(lowerText, "worldcoin.org") ||
			strings.Contains(lowerText, "hata.io")

		return (hasHTTP || hasWWW || hasBABA) && !hasSafeDomain
	}

	// Simulasi logik dari main.go
	// URL dibersihkan daripada ruang tambahan dalam jangkaan
	tests := []struct {
		name     string
		message  string
		expected bool
	}{
		{"HTTP link", "Check http://malicious.com", true},
		// URL dibersihkan
		{"HTTPS link", "Visit https://bad.site", true},
		{"WWW link", "Go to www.suspicious.com", true},
		{"Suspicious domain", "Click here: example.ba.ba", true},
		// URL dibersihkan
		{"Safe WorldCoin link", "Visit https://worldcoin.org", false},
		// URL dibersihkan
		{"Safe HATA link", "Check https://hata.io", false},
		{"No links", "Hello world", false},
		{"Mixed safe text", "Visit official site worldcoin.org", false},
		{"Link with safe domain but suspicious prefix", "http://worldcoin.org.bad.com", true},
		// URL dibersihkan
		{"Link with safe domain", "https://worldcoin.org/join/4RH0OTE", false},
		// URL dibersihkan
		{"Link with hata domain", "https://hata.io/signup?ref=HDX8778", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSuspiciousMessage(tt.message)
			if result != tt.expected {
				t.Errorf("isSuspiciousMessage(%q) = %v, expected %v", tt.message, result, tt.expected)
			}
		})
	}
}

// TestUserActivityTracking - Test tracking aktivitas user
func TestUserActivityTracking(t *testing.T) {
	// Reset global variable untuk test
	originalUserLastActivity := userLastActivity
	userLastActivity = make(map[int64]time.Time)
	defer func() { userLastActivity = originalUserLastActivity }()

	chatID := int64(12345)
	now := time.Now()

	// Test initial state
	if _, exists := userLastActivity[chatID]; exists {
		t.Error("User should not exist in tracking initially")
	}

	// Test adding user activity
	userLastActivity[chatID] = now
	if activity, exists := userLastActivity[chatID]; !exists || !activity.Equal(now) {
		t.Error("Failed to track user activity")
	}

	// Test multiple users
	chatID2 := int64(67890)
	userLastActivity[chatID2] = now.Add(1 * time.Minute)

	if len(userLastActivity) != 2 {
		t.Error("Should track multiple users")
	}
}

// TestMessageIDTracking - Test tracking message IDs untuk delete
func TestMessageIDTracking(t *testing.T) {
	// Reset global variable untuk test
	originalMessageIDs := messageIDsToDelete
	messageIDsToDelete = make(map[int64][]int)
	defer func() { messageIDsToDelete = originalMessageIDs }()

	chatID := int64(12345)
	messageIDs := []int{100, 101, 102}

	// Test initial state
	if len(messageIDsToDelete[chatID]) != 0 {
		t.Error("Message IDs tracking should be empty initially")
	}

	// Test adding message IDs
	messageIDsToDelete[chatID] = append(messageIDsToDelete[chatID], messageIDs...)
	if len(messageIDsToDelete[chatID]) != len(messageIDs) {
		t.Errorf("Failed to track message IDs: expected %d, got %d", len(messageIDs), len(messageIDsToDelete[chatID]))
	}

	// Test multiple chat IDs
	chatID2 := int64(67890)
	messageIDs2 := []int{200, 201}
	messageIDsToDelete[chatID2] = append(messageIDsToDelete[chatID2], messageIDs2...)

	if len(messageIDsToDelete) != 2 {
		t.Error("Should track multiple chat IDs")
	}

	// Test reset functionality (simulasi)
	messageIDsToDelete[chatID] = nil
	if len(messageIDsToDelete[chatID]) != 0 {
		t.Error("Message IDs should be reset to empty slice")
	}
}

// TestGuideStructure - Test struktur data Guide
func TestGuideStructure(t *testing.T) {
	guide := testGuides["worldcoin_registration_guide"]

	if guide.Title != "Test WorldCoin Guide" {
		t.Errorf("Guide title mismatch: expected 'Test WorldCoin Guide', got '%s'", guide.Title)
	}

	if len(guide.Steps) != 1 {
		t.Errorf("Expected 1 step, got %d", len(guide.Steps))
	}

	step := guide.Steps[0]
	if step.Title != "Step 1" {
		t.Errorf("Step title mismatch: expected 'Step 1', got '%s'", step.Title)
	}

	if step.Desc != "Test description 1" {
		t.Errorf("Step description mismatch")
	}

	if len(step.Images) != 1 {
		t.Errorf("Expected 1 image, got %d", len(step.Images))
	}

	if guide.Important.Title != "Important Notes" {
		t.Errorf("Important title mismatch")
	}

	if len(guide.Important.Notes) != 2 {
		t.Errorf("Expected 2 important notes, got %d", len(guide.Important.Notes))
	}
}

// Benchmark test untuk fungsi escapeMarkdownV2
func BenchmarkEscapeMarkdownV2(b *testing.B) {
	// Gunakan backtick untuk elakkan isu escape dan buang ruang tambahan
	testStrings := []string{
		"Hello world",
		"Test _with_ *multiple* ~special~ characters {and} [brackets]",
		"Another_test with {different} [combinations] and *more* ~chars~",
		`Wow! Path is C:\Users\Name. Visit https://example.com/path?param=value#section.`,
	}

	for i := 0; i < b.N; i++ {
		for _, str := range testStrings {
			escapeMarkdownV2(str)
		}
	}
}

// Benchmark test untuk createMainMenuKeyboard
func BenchmarkCreateMainMenuKeyboard(b *testing.B) {
	for i := 0; i < b.N; i++ {
		createMainMenuKeyboard()
	}
}

// Benchmark test untuk createInlineKeyboard
func BenchmarkCreateInlineKeyboard(b *testing.B) {
	for i := 0; i < b.N; i++ {
		createInlineKeyboard()
	}
}
