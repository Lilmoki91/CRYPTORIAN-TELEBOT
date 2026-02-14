package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// Konfigurasi Utama
const ADMIN_ID int64 = 007 // 7348614053  ID Asal Mr JOHAN

var (
	githubToken = os.Getenv("GITHUB_TOKEN")
	githubRepo  = "Lilmoki91/CRYPTORIAN-TELEBOT"
	// Gunakan link raw yang stabil
	termsURL = "https://raw.githubusercontent.com/Lilmoki91/CRYPTORIAN-TELEBOT/main/terms.json"
)

// Struktur Data disederhanakan supaya lebih fleksibel
type TermsData struct {
	ProjectName        string `json:"project_name"`
	TermsAndConditions struct {
		Title    string `json:"title"`
		Sections []struct {
			ID      int      `json:"id"`
			Heading string   `json:"heading"`
			Content []string `json:"content"`
		} `json:"sections"`
	} `json:"terms_and_conditions"`
}

func IsAdmin(userID int64) bool {
	return userID == ADMIN_ID
}

func IsBanned(userID int64) bool {
	if IsAdmin(userID) {
		return false
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/contents/blacklist/%d.json", githubRepo, userID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+githubToken)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func HasAgreed(userID int64) bool {
	if IsAdmin(userID) {
		return true
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/contents/agreements/%d.json", githubRepo, userID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+githubToken)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// BuildTermsUI - VERSI BAIKI
func BuildTermsUI() (string, error) {
	resp, err := http.Get(termsURL)
	if err != nil {
		return getFallbackTerms(), nil // Guna Plan B jika internet ralat
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var data TermsData
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Printf("Ralat JSON: %v", err)
		return getFallbackTerms(), nil // Guna Plan B jika JSON salah format
	}

	var sb strings.Builder
	// Header
	sb.WriteString(fmt.Sprintf("🛡️ *%s*\n\n", escapeMarkdownV2(data.TermsAndConditions.Title)))

	// Kandungan
	for _, sec := range data.TermsAndConditions.Sections {
		sb.WriteString(fmt.Sprintf("%d\\. *%s*\n", sec.ID, escapeMarkdownV2(sec.Heading)))
		for _, line := range sec.Content {
			sb.WriteString(fmt.Sprintf("• %s\n", escapeMarkdownV2(line)))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("───────────────\n")
	sb.WriteString("_Sila tekan butang di bawah untuk bersetuju dan memulakan sesi\\._")

	return sb.String(), nil
}

// Plan B jika sistem gagal baca JSON
func getFallbackTerms() string {
	return "📜 *TERMA & SYARAT CRYPTORIAN*\n\n" +
		"1\\. *Kedaulatan Sistem*\n" +
		"Anda bersetuju untuk mematuhi segala peraturan yang ditetapkan\\.\n\n" +
		"2\\. *Privasi*\n" +
		"Data anda hanya digunakan untuk tujuan audit bot\\.\n\n" +
		"───────────────\n" +
		"_Sistem gagal memuat fail JSON, menggunakan teks kecemasan\\._"
}

// Fungsi Escape MarkdownV2 yang wajib ada
func escapeMarkdownV2(text string) string {
	replacer := strings.NewReplacer(
		"_", "\\_", "*", "\\*", "[", "\\[", "]", "\\]", "(", "\\(",
		")", "\\)", "~", "\\~", "`", "\\`", ">", "\\>", "#", "\\#",
		"+", "\\+", "-", "\\-", "=", "\\=", "|", "\\|", "{", "\\{",
		"}", "\\}", ".", "\\.", "!", "\\!",
	)
	return replacer.Replace(text)
}

// Fungsi simpan ke GitHub dan Ban tetap sama...
func SaveAgreementToGithub(userID int64, username string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/contents/agreements/%d.json", githubRepo, userID)
	logData := map[string]interface{}{
		"user_id":    userID,
		"username":   username,
		"agreed_at":  time.Now().Format(time.RFC3339),
		"status":     "AGREED",
	}
	jsonBytes, _ := json.MarshalIndent(logData, "", "  ")
	payload := map[string]interface{}{
		"message": fmt.Sprintf("Audit Log: User %d agreed", userID),
		"content": base64.StdEncoding.EncodeToString(jsonBytes),
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+githubToken)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	return nil
}
