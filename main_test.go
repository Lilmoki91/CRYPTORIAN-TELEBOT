package main

import (
	"os"
	"testing"
)

func TestEscapeMarkdown(t *testing.T) {
	input := "_Test *Markdown* [Escape](url) ~`>#+-=|{}.!"
	expected := "\\_Test \\*Markdown\\* \\[Escape\\]\\(url\\) \\~\\`\\>\\#\\+\\-\\=\\|\\{\\}\\.\\!"
	result := EscapeMarkdown(input)
	if result != expected {
		t.Errorf("EscapeMarkdown failed.\nExpected: %s\nGot: %s", expected, result)
	}
}

func TestLoadMarkdownData(t *testing.T) {
	// Prepare a temporary JSON file for testing
	jsonContent := `
{
  "worldcoin_registration": {
    "title": "Test Title",
    "steps": [],
    "points": [],
    "important": {
      "title": "Imp",
      "notes": []
    }
  },
  "hata_wallet_setup": {
    "title": "",
    "steps": [],
    "points": [],
    "important": {
      "title": "",
      "notes": []
    }
  },
  "withdraw_to_hata": {
    "title": "",
    "steps": [],
    "points": [],
    "important": {
      "title": "",
      "notes": []
    }
  },
  "cashout_to_bank": {
    "title": "",
    "steps": [],
    "points": [],
    "important": {
      "title": "",
      "notes": []
    }
  },
  "security_notes": {
    "title": "",
    "steps": [],
    "points": [],
    "important": {
      "title": "",
      "notes": []
    }
  }
}
`
	tmpFile, err := os.CreateTemp("", "test_markdown.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(jsonContent)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	data, err := LoadMarkdownData(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadMarkdownData failed: %v", err)
	}
	if data.WorldcoinRegistration.Title != "Test Title" {
		t.Errorf("Expected title 'Test Title', got '%s'", data.WorldcoinRegistration.Title)
	}
}

func TestFormatSection(t *testing.T) {
	section := Section{
		Title: "Section Title",
		Steps: []Step{
			{Step: 1, Title: "First", Desc: "Do this", Images: nil, Links: nil},
		},
		Points: []string{"Point 1", "Point 2"},
		Important: struct {
			Title string   `json:"title"`
			Notes []string `json:"notes"`
		}{
			Title: "Important",
			Notes: []string{"Note 1"},
		},
	}
	result := FormatSection(section)
	if len(result) == 0 {
		t.Error("FormatSection returned empty string")
	}
	if want := "*Section Title*"; !contains(result, want) {
		t.Errorf("FormatSection missing section title: want '%s'", want)
	}
	if want := "Step 1"; !contains(result, want) {
		t.Errorf("FormatSection missing step: want '%s'", want)
	}
	if want := "Point 1"; !contains(result, want) {
		t.Errorf("FormatSection missing point: want '%s'", want)
	}
	if want := "Important"; !contains(result, want) {
		t.Errorf("FormatSection missing important title: want '%s'", want)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr))))
}
