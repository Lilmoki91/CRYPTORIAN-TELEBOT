package main

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ====================
//  LIGHT TEST (selalu jalan)
// ====================

// Dummy struct sama macam Guide di main.go
type Guide struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// Escape Markdown test
func TestEscapeMarkdownV2(t *testing.T) {
	input := "_hello*world?[]"
	expected := "\\_hello\\*world\\?\\[\\]"
	got := escapeMarkdownV2(input)
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

// Test parse JSON
func TestLoadGuides(t *testing.T) {
	data := `[{"title":"Intro","body":"Welcome"}]`
	var guides []Guide
	if err := json.Unmarshal([]byte(data), &guides); err != nil {
		t.Fatalf("JSON parse gagal: %v", err)
	}
	if guides[0].Title != "Intro" {
		t.Errorf("expected 'Intro', got %s", guides[0].Title)
	}
}

// Test file markdown.json wujud
func TestMarkdownFileExists(t *testing.T) {
	if _, err := os.Stat("markdown.json"); os.IsNotExist(err) {
		t.Errorf("File markdown.json tidak wujud!")
	}
}

// ====================
//  REAL BOT TEST (jalan hanya jika ada TOKEN & CHAT_ID)
// ====================

func TestBotInit(t *testing.T) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		t.Skip("TELEGRAM_BOT_TOKEN tidak diset, skip real bot test")
	}
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		t.Fatalf("Gagal buat bot: %v", err)
	}
	t.Logf("Bot username: %s", bot.Self.UserName)
}

func TestSendMessage(t *testing.T) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		t.Skip("TELEGRAM_BOT_TOKEN tidak diset, skip real bot test")
	}
	chatIDStr := os.Getenv("TEST_CHAT_ID")
	if chatIDStr == "" {
		t.Skip("TEST_CHAT_ID tidak diset, skip real bot test")
	}

	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		t.Fatalf("TEST_CHAT_ID bukan nombor: %v", err)
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		t.Fatalf("Gagal buat bot: %v", err)
	}

	msg := tgbotapi.NewMessage(chatID, "✅ Bot test berjaya!")
	_, err = bot.Send(msg)
	if err != nil {
		t.Fatalf("Gagal hantar mesej: %v", err)
	}
}
