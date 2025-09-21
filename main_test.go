package main

import (
	"encoding/json"
	"os"
	"testing"
)

// Test escapeMarkdownV2 correctness
func TestEscapeMarkdownV2(t *testing.T) {
	raw := "_test[link](url)!"
	expected := "\\_test\\[link\\]\\(url\\)\\!"
	got := escapeMarkdownV2(raw)
	if got != expected {
		t.Errorf("escapeMarkdownV2 failed: got %s, expected %s", got, expected)
	}
}

// Test JSON parsing into Guide struct
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
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	guide, ok := guides["worldcoin_registration_guide"]
	if !ok {
		t.Fatalf("Guide not found in parsed data")
	}

	if guide.Title != "WorldCoin Registration" {
		t.Errorf("Wrong title: %s", guide.Title)
	}
	if len(guide.Steps) != 1 {
		t.Errorf("Expected 1 step, got %d", len(guide.Steps))
	}
	if guide.Important.Title != "Note" {
		t.Errorf("Important title mismatch")
	}
}

// Test if markdown.json exists in repo
func TestMarkdownFileExists(t *testing.T) {
	if _, err := os.Stat("markdown.json"); os.IsNotExist(err) {
		t.Errorf("markdown.json does not exist")
	}
}
