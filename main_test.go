package main

import (
	"os"
	"testing"
)

// ===============================
// Test asas Markdown
// ===============================
func TestEscapeMarkdownV2(t *testing.T) {
	in := "_Hello* [World]"
	out := escapeMarkdownV2(in)

	if out == in {
		t.Errorf("escapeMarkdownV2 gagal, output sama: %s", out)
	}
}

// ===============================
// Test load fail panduan
// ===============================
func TestLoadGuides(t *testing.T) {
	_, err := loadGuides()
	if err != nil {
		t.Errorf("loadGuides gagal: %v", err)
	}
}

// ===============================
// Test fail markdown wujud
// ===============================
func TestMarkdownFileExists(t *testing.T) {
	files := []string{"README.md"}
	for _, f := range files {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			t.Errorf("Fail markdown tidak wujud: %s", f)
		}
	}
}

// ===============================
// Test token Telegram
// ===============================
func TestBotTokenExists(t *testing.T) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		t.Skip("TELEGRAM_BOT_TOKEN tidak diset, skip test ini")
	}
}

// ===============================
// Test inisialisasi bot
// ===============================
func TestBotInit(t *testing.T) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		t.Skip("TELEGRAM_BOT_TOKEN tidak diset, skip TestBotInit")
	}

	// Cuba buat bot instance
	_, err := initBot()
	if err != nil {
		t.Fatalf("Gagal init bot: %v", err)
	}
}
