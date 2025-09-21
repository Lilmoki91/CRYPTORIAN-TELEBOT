	
package main

import (
	"testing"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Mock Bot untuk testing
type mockBot struct {
	lastSentMsg tgbotapi.Chattable
}

func (m *mockBot) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	m.lastSentMsg = c
	return tgbotapi.Message{}, nil
}

func (m *mockBot) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	return &tgbotapi.APIResponse{}, nil
}

func (m *mockBot) GetUpdatesChan(config tgbotapi.UpdateConfig) (tgbotapi.UpdatesChannel, error) {
	return make(chan tgbotapi.Update), nil
}

// Test fungsi handleMessage dengan berbagai input
func TestHandleMessage(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "Test start command",
			message:  "/start",
			expected: "Welcome",
		},
		{
			name:     "Test help command",
			message:  "/help",
			expected: "help",
		},
		{
			name:     "Test hello message",
			message:  "hello",
			expected: "Hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handleMessage(tt.message)
			if !strings.Contains(strings.ToLower(result), strings.ToLower(tt.expected)) {
				t.Errorf("handleMessage(%s) = %s, expected to contain %s", tt.message, result, tt.expected)
			}
		})
	}
}

// Test pemrosesan update
func TestProcessUpdate(t *testing.T) {
	bot := &mockBot{}
	
	// Test update dengan message
	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{
				ID: 12345,
			},
			Text: "/start",
		},
	}

	processUpdate(bot, update)
	
	// Verifikasi bahwa bot mencoba mengirim message
	if bot.lastSentMsg == nil {
		t.Error("Bot should have tried to send a message")
	}
}

// Test fungsi bantuan (helper functions)
func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		item     string
		expected bool
	}{
		{
			name:     "Item exists in slice",
			slice:    []string{"hello", "world"},
			item:     "hello",
			expected: true,
		},
		{
			name:     "Item does not exist in slice",
			slice:    []string{"hello", "world"},
			item:     "test",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.slice, tt.item)
			if result != tt.expected {
				t.Errorf("contains(%v, %s) = %v, expected %v", tt.slice, tt.item, result, tt.expected)
			}
		})
	}
}

// Test initialization
func TestBotInitialization(t *testing.T) {
	// Test bahwa bot bisa dibuat dengan token mock
	// Ini akan gagal jika ada panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Bot initialization panicked: %v", r)
		}
	}()

	// Test dengan token kosong (should handle gracefully)
	_ = initBot("")
}

// Benchmark test untuk measure performance
func BenchmarkHandleMessage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		handleMessage("/start")
	}
}

// Contoh test dengan mock
func TestHandleMessageWithMock(t *testing.T) {
	mockBot := &mockBot{}
	
	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{
				ID: 12345,
			},
			Text: "test message",
		},
	}
	
	// Process update
	processUpdate(mockBot, update)
	
	// Anda bisa menambahkan assertions lebih spesifik di sini
	// berdasarkan implementasi aktual bot anda
}
