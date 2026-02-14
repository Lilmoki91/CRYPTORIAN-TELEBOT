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
	} `json:"terms_and_conditions"`
}

func IsAdmin(userID int64) bool {
	return userID == ADMIN_ID
}

func IsBanned(userID int64) bool {
	if IsAdmin(userID) { return false }
	url := fmt.Sprintf("https://api.github.com/repos/%s/contents/blacklist/%d.json", githubRepo, userID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+githubToken)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil { return false }
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func HasAgreed(userID int64) bool {
	if IsAdmin(userID) { return true }
	url := fmt.Sprintf("https://api.github.com/repos/%s/contents/agreements/%d.json", githubRepo, userID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+githubToken)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil { return false }
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// BuildTermsUI - Hanya membina string, tiada pembersihan markdown di sini
func BuildTermsUI() (string, error) {
	resp, err := http.Get(termsURL)
	if err != nil {
		return "📜 TERMA & SYARAT\n\nGagal memuatkan terma.", nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var data TermsData
	json.Unmarshal(body, &data)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s\n\n", data.TermsAndConditions.Title))

	for _, sec := range data.TermsAndConditions.Sections {
		sb.WriteString(fmt.Sprintf("%d. %s\n", sec.ID, sec.Heading))
		for _, line := range sec.Content {
			sb.WriteString(fmt.Sprintf("• %s\n", line))
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

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
