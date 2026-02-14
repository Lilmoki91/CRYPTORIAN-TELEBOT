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

// Konfigurasi - GITHUB_TOKEN mesti diset di Koyeb Environment Variables
var (
    githubToken = os.Getenv("GITHUB_TOKEN")
    githubRepo  = "Lilmoki91/CRYPTORIAN-TELEBOT"
    // Gunakan URL raw yang lebih ringkas
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

// HasAgreed menyemak jika fail ID user wujud di GitHub
func HasAgreed(userID int64) bool {
    if githubToken == "" {
        return false
    }
    
    url := fmt.Sprintf("https://api.github.com/repos/%s/contents/agreements/%d.json", githubRepo, userID)
    req, _ := http.NewRequest("GET", url, nil)
    req.Header.Set("Authorization", "Bearer "+githubToken) // Gunakan Bearer untuk token moden
    req.Header.Set("Accept", "application/vnd.github+json")

    client := &http.Client{Timeout: 5 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return false
    }
    defer resp.Body.Close()

    // Jika status 200 OK, bermakna fail wujud (user sudah setuju)
    return resp.StatusCode == http.StatusOK
}

// BuildTermsUI mengambil JSON dan menukarnya menjadi teks cantik untuk Telegram, escape markdown V2
func BuildTermsUI() (string, error) {
    resp, err := http.Get(termsURL)
    if err != nil {
        return "", fmt.Errorf("gagal akses URL Terma: %v", err)
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)
    var data TermsData
    json.Unmarshal(body, &data)

    var sb strings.Builder
    
    // Tajuk (Bold)
    sb.WriteString(fmt.Sprintf("*%s*\n\n", escapeMarkdownV2(data.TermsAndConditions.Title)))
    sb.WriteString(escapeMarkdownV2("Sila baca dan patuhi:") + "\n\n")

    for _, sec := range data.TermsAndConditions.Sections {
        // Nombor Seksyen (Contoh: 1\.)
        sb.WriteString(fmt.Sprintf("%d\\. *%s*\n", sec.ID, escapeMarkdownV2(sec.Heading)))
        for _, line := range sec.Content {
            sb.WriteString(fmt.Sprintf("• %s\n", escapeMarkdownV2(line)))
        }
        sb.WriteString("\n")
    }

    sb.WriteString("\\-\\-\\-\n") // Separator
    sb.WriteString("_" + escapeMarkdownV2("Untuk teruskan sesi operasi bot sila pilih:") + "_")

    return sb.String(), nil
}


// SaveAgreementToGithub menyimpan fail JSON baru ke repo sebagai bukti audit
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
        "message": fmt.Sprintf("Audit Log: User %d has agreed to terms", userID),
        "content": base64.StdEncoding.EncodeToString(jsonBytes),
    }
    
    body, _ := json.Marshal(payload)
    req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(body))
    req.Header.Set("Authorization", "Bearer "+githubToken)
    req.Header.Set("Accept", "application/vnd.github+json")
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
        return fmt.Errorf("GitHub API Error: %s", resp.Status)
    }

    return nil
}
