package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Konfigurasi Utama
// TIPS: Tukar ID ini ke 007 untuk Misi Penyamaran, tukar balik ke asal untuk jadi Raja.
const ADMIN_ID int64 = 7348614053 

var (
	githubToken = os.Getenv("GITHUB_TOKEN")
	githubRepo  = "Lilmoki91/CRYPTORIAN-TELEBOT"
	termsURL    = "https://raw.githubusercontent.com/Lilmoki91/CRYPTORIAN-TELEBOT/main/terms.json"
)

type TermsData struct {
	ProjectName        string `json:"project_name"`
	TermsAndConditions struct {
		Title    string `json:"title"`
		Sections []struct {
			ID      int      `json:"id"`
			Heading string   `json:"heading"`
			Content []string `json:"content"`
		} `json:"sections"`
		Footer    string `json:"footer"`
		Copyright string `json:"copyright"`
	} `json:"terms_and_conditions"`
}

// IsAdmin menyemak jika pengguna adalah Mr JOHAN
func IsAdmin(userID int64) bool {
	return userID == ADMIN_ID
}

// IsBanned menyemak jika ID user wujud dalam folder blacklist/
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

// HasAgreed menyemak jika fail ID user wujud di folder agreements/
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

// BuildTermsUI mengambil JSON dan menukarnya menjadi teks MarkdownV2
func BuildTermsUI() (string, error) {
	resp, err := http.Get(termsURL)
	if err != nil {
		return "📜 *TERMA & SYARAT*\n\nSila patuhi peraturan empire\\.", nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var data TermsData
	json.Unmarshal(body, &data)

	var sb strings.Builder
	
	// Gunakan escapeMarkdownV2 yang sedia ada di main.go
	sb.WriteString(fmt.Sprintf("🛡️ *%s*\n\n", escapeMarkdownV2(data.TermsAndConditions.Title)))

	for _, sec := range data.TermsAndConditions.Sections {
		sb.WriteString(fmt.Sprintf("%d\\. *%s*\n", sec.ID, escapeMarkdownV2(sec.Heading)))
		for _, line := range sec.Content {
			sb.WriteString(fmt.Sprintf("• %s\n", escapeMarkdownV2(line)))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("───────────────\n")
	sb.WriteString("_Sila tekan butang di bawah untuk bermula\\._")

	return sb.String(), nil
}

// SaveAgreementToGithub menyimpan fail JSON baru ke repo (Audit Log)
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

// BanUser digunakan oleh Admin untuk sekat user
func BanUser(userID int64, reason string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/contents/blacklist/%d.json", githubRepo, userID)
	
	logData := map[string]interface{}{
		"user_id":    userID,
		"reason":     reason,
		"banned_at":  time.Now().Format(time.RFC3339),
		"by_admin":   "Mr JOHAN",
	}
	
	jsonBytes, _ := json.MarshalIndent(logData, "", "  ")
	payload := map[string]interface{}{
		"message": fmt.Sprintf("Admin Action: Banning user %d", userID),
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
