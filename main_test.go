package main

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

//
// ✅ Uji fungsi escapeMarkdownV2
//
func TestEscapeMarkdownV2(t *testing.T) {
	raw := "_test[link](url)!"
	expected := "\\_test\\[link\\]\\(url\\)\\!"
	got := escapeMarkdownV2(raw)
	if got != expected {
		t.Errorf("escapeMarkdownV2 salah: dapat %s, sepatutnya %s", got, expected)
	}
}

//
// ✅ Uji JSON parsing → struct Guide
//
func TestLoadGuides(t *testing.T) {
	data := `
	{
		"worldcoin_registration_guide": {
			"title": "WorldCoin Registration",
			"steps": [
				{"title": "Step 1", "desc": "Open app", "images": ["http://img1"]}
			],
			"important": {"title": "Note", "notes": ["Keep safe"]}
		}
	}`

	var guides map[string]Guide
	if err := json.Unmarshal([]byte(data), &guides); err != nil {
		t.Fatalf("Gagal unmarshal JSON: %v", err)
	}

	guide, ok := guides["worldcoin_registration_guide"]
	if !ok {
		t.Fatalf("Guide tidak dijumpai dalam data")
	}

	if guide.Title != "WorldCoin Registration" {
		t.Errorf("Title salah: %s", guide.Title)
	}
	if len(guide.Steps) != 1 {
		t.Errorf("Sepatutnya ada 1 step, dapat %d", len(guide.Steps))
	}
	if guide.Important.Title != "Note" {
		t.Errorf("Important.Title salah: %s", guide.Important.Title)
	}
}

//
// ✅ Uji kewujudan TELEGRAM_BOT_TOKEN
//
func TestBotTokenExists(t *testing.T) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		t.Fatal("TELEGRAM_BOT_TOKEN tidak diset! Jalankan `export TELEGRAM_BOT_TOKEN=xxx` sebelum test.")
	}
}

//
// ✅ Uji kewujudan file markdown.json
//
func TestMarkdownFileExists(t *testing.T) {
	if _, err := os.Stat("markdown.json"); os.IsNotExist(err) {
		t.Errorf("Fail markdown.json tidak wujud di root project")
	}
}

//
// ✅ Uji bot init → pastikan main() tidak panic
//
func TestBotInit(t *testing.T) {
	done := make(chan struct{})
	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("main() panic: %v", r)
			}
			close(done)
		}()
		main()
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		// Jika masih running selepas 2s, dianggap normal (bot loop)
	}
}
