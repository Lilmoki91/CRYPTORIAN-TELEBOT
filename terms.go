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
const ADMIN_ID int64 = 007 // 7348614053 

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

// IsBanned menyemak jika ID user wujud dalam folder blacklist/ di GitHub
func IsBanned(userID int64) bool {
	if IsAdmin(userID) {
		return false // Admin tidak boleh di-ban
	}

	if githubToken == "" {
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

// HasAgreed menyemak jika fail ID user wujud di GitHub (Admin automatik lepas)
func HasAgreed(userID int64) bool {
	if IsAdmin(userID) {
		return true // Mr JOHAN tak perlu klik setuju
	}

	if githubToken == "" {
		return false
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/contents/agreements/%d.json", githubRepo, userID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+githubToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// BuildTermsUI mengambil JSON dan menukarnya menjadi teks Markdown Standard (Tanpa V2 Escape)
func BuildTermsUI() (string, error) {
	resp, err := http.Get(termsURL)
	if err != nil {
		return "", fmt.Errorf("gagal akses URL Terma: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var data TermsData
	if err := json.Unmarshal(body, &data); err != nil {
		return "", fmt.Errorf("gagal parse JSON: %v", err)
	}

	var sb strings.Builder

	// Header
	// Menggunakan format **Teks** untuk bold dalam Markdown Standard
	sb.WriteString(fmt.Sprintf("**%s**\n\n", data.TermsAndConditions.Title))
	sb.WriteString("Sila baca dan patuhi terma dan syarat berikut:\n\n")

	// Sections
	for _, sec := range data.TermsAndConditions.Sections {
		// Menggunakan titik biasa ".", bukan "\."
		sb.WriteString(fmt.Sprintf("%d. **%s**\n", sec.ID, sec.Heading))
		for _, line := range sec.Content {
			sb.WriteString(fmt.Sprintf("• %s\n", line))
		}
		sb.WriteString("\n")
	}

	// Footer
	// Menggunakan garisan visual biasa
	sb.WriteString("─────────────────\n\n")
	// Menggunakan _Teks_ untuk italic dalam Markdown Standard
	sb.WriteString("_Untuk teruskan sesi operasi bot sila pilih:_")

	return sb.String(), nil
}

// SaveAgreementToGithub menyimpan fail JSON baru ke repo (Audit Log)
func SaveAgreementToGithub(userID int64, username string) error {
	if githubToken == "" {
		return fmt.Errorf("GITHUB_TOKEN tidak ditetapkan")
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/contents/agreements/%d.json", githubRepo, userID)

	logData := map[string]interface{}{
		"user_id":   userID,
		"username":  username,
		"agreed_at": time.Now().Format(time.RFC3339),
		"status":    "AGREED",
	}

	jsonBytes, _ := json.MarshalIndent(logData, "", "  ")

	// Check if file already exists to get SHA (untuk update)
	var sha string
	existingURL := fmt.Sprintf("https://api.github.com/repos/%s/contents/agreements/%d.json", githubRepo, userID)
	req, _ := http.NewRequest("GET", existingURL, nil)
	req.Header.Set("Authorization", "Bearer "+githubToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err == nil && resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		var existingData map[string]interface{}
		body, _ := io.ReadAll(resp.Body)
		json.Unmarshal(body, &existingData)
		if s, ok := existingData["sha"].(string); ok {
			sha = s
		}
	}

	payload := map[string]interface{}{
		"message": fmt.Sprintf("Audit Log: User %d has agreed to terms", userID),
		"content": base64.StdEncoding.EncodeToString(jsonBytes),
	}

	if sha != "" {
		payload["sha"] = sha
	}

	body, _ := json.Marshal(payload)
	req, _ = http.NewRequest("PUT", url, bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+githubToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Github API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// BanUser digunakan oleh Admin untuk sekat user secara automatik ke GitHub
func BanUser(userID int64, reason string) error {
	if githubToken == "" {
		return fmt.Errorf("GITHUB_TOKEN tidak ditetapkan")
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/contents/blacklist/%d.json", githubRepo, userID)

	logData := map[string]interface{}{
		"user_id":   userID,
		"reason":    reason,
		"banned_at": time.Now().Format(time.RFC3339),
		"by_admin":  "Mr JOHAN",
	}

	jsonBytes, _ := json.MarshalIndent(logData, "", "  ")

	// Check if file already exists to get SHA
	var sha string
	existingURL := fmt.Sprintf("https://api.github.com/repos/%s/contents/blacklist/%d.json", githubRepo, userID)
	req, _ := http.NewRequest("GET", existingURL, nil)
	req.Header.Set("Authorization", "Bearer "+githubToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err == nil && resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		var existingData map[string]interface{}
		body, _ := io.ReadAll(resp.Body)
		json.Unmarshal(body, &existingData)
		if s, ok := existingData["sha"].(string); ok {
			sha = s
		}
	}

	payload := map[string]interface{}{
		"message": fmt.Sprintf("Admin Action: Banning user %d", userID),
		"content": base64.StdEncoding.EncodeToString(jsonBytes),
	}

	if sha != "" {
		payload["sha"] = sha
	}

	body, _ := json.Marshal(payload)
	req, _ = http.NewRequest("PUT", url, bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+githubToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Github API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
