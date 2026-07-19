package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
)

// ZohoMailPayload defines the structure Zoho Mail REST API expects
type ZohoMailPayload struct {
	FromAddress string `json:"fromAddress"`
	ToAddress   string `json:"toAddress"`
	Subject     string `json:"subject"`
	Content     string `json:"content"`
	Mode        string `json:"mode"`       // "mail" to send immediately, "draft" to save
	MailFormat  string `json:"mailFormat"` // "plaintext" or "html"
}

// func DraftEmail() bool {
// 	slog.Info("begin drafting")

// 	// 1. Connection configuration
// 	imapServer := "imap.zoho.com:993"
// 	username := os.Getenv("SENDER")
// 	password := os.Getenv("ZOHO_PWD") // Use an App-Specific Password if 2FA is active

// 	slog.Info("Connecting to Zoho IMAP...")

// 	c, err := client.DialTLS(imapServer, nil)
// 	if err != nil {
// 		slog.Error("Connection failed:", slog.String("error msg", err.Error()))
// 		return false
// 	}
// 	defer c.Logout()

// 	// 2. Authenticate
// 	if err := c.Login(username, password); err != nil {
// 		slog.Error("Login failed:", slog.String("error msg", err.Error()))
// 		return false
// 	}
// 	slog.Info("Logged in successfully.")

// 	// 3. Construct the raw email payload
// 	from := username
// 	to := "alex@steadypartner.online"
// 	subject := "Draft email created via IMAP in Go"
// 	body := "Test draft."

// 	// Format matching standard internet message syntax (RFC 5322)
// 	var msgBytes bytes.Buffer
// 	msgBytes.WriteString(fmt.Sprintf("From: %s\r\n", from))
// 	msgBytes.WriteString(fmt.Sprintf("To: %s\r\n", to))
// 	msgBytes.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
// 	msgBytes.WriteString("MIME-Version: 1.0\r\n")
// 	msgBytes.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
// 	msgBytes.WriteString("\r\n") // Blank line separating headers from body
// 	msgBytes.WriteString(body)

// 	// 4. Append message to the target folder "Drafts"
// 	targetFolder := "Drafts"

// 	// Add the `\Draft` flag so the UI natively shows it as an editable draft
// 	flags := []string{imap.DraftFlag}

// 	err = c.Append(targetFolder, flags, time.Now(), &msgBytes)
// 	if err != nil {
// 		slog.Error("Failed to append draft:", slog.String("error msg", err.Error()))
// 		return false
// 	}

// 	slog.Info("Success! The email has been saved", slog.String("target folder", targetFolder))
// 	return true
// }

func DraftEmail(accessToken, accountID string) bool {
	apiURL := fmt.Sprintf("https://mail.zoho.com/api/accounts/%s/messages", accountID)

	slog.Info(apiURL)

	payload := ZohoMailPayload{
		FromAddress: os.Getenv("SENDER"),
		ToAddress:   "alex@steadypartner.online",
		Subject:     "test Draft",
		Content:     "This draft was constructed entirely using the automated Go pipeline.",
		Mode:        "draft", // Ensures it saves instead of sends
		MailFormat:  "plaintext",
	}

	jsonBytes, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBytes))
	req.Header.Set("Authorization", "Zoho-oauthtoken "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("API call failed:", slog.String("error msg", err.Error()))
		return false
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	slog.Info("HTTP Status", slog.String("Response", string(respBody)))
	return true
}
