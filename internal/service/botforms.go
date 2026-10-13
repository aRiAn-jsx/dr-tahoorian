package service

// Forwarding of contact-form submissions to the SSI holding Telegram bot.
// This file exists only for the Telegram bot integration: the bot stores each
// submission and notifies its admins.
//
// Environment variables (nothing is sent unless both are set):
//
//	BOT_FORMS_URL     base URL of the bot's HTTP server, e.g. https://bot.example.com
//	BOT_FORMS_SECRET  shared secret (SITE_FORMS_SECRET in the bot's .env)

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"dr-tahoorian/internal/domain"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const botFormsSite = "tahoorian"

var (
	botFormsRetryDelays = []time.Duration{0, 2 * time.Second, 8 * time.Second}
	botFormsClient      = &http.Client{Timeout: 5 * time.Second}
)

type botFormPayload struct {
	Form      string `json:"form"`
	FormTitle string `json:"form_title"`
	SiteRef   string `json:"site_ref"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	Subject   string `json:"subject"`
	Message   string `json:"message"`
}

// forwardToBot runs in its own goroutine so the visitor never waits for the bot.
func (s *Service) forwardToBot(c domain.Contact) {
	payload := botFormPayload{
		Form:      "contact",
		FormTitle: "فرم تماس",
		Name:      c.Name,
		Phone:     c.Phone,
		Email:     c.Email,
		Subject:   c.Subject,
		Message:   c.Message,
	}
	if c.ID > 0 {
		payload.SiteRef = strconv.FormatInt(c.ID, 10)
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return
	}

	for _, delay := range botFormsRetryDelays {
		time.Sleep(delay)
		if err = s.postToBot(body); err == nil {
			return
		}
		log.Printf("bot-forms: contact %d not delivered: %v", c.ID, err)
	}
}

func (s *Service) postToBot(body []byte) error {
	// Signature: HMAC-SHA256 over "timestamp.body" — the secret never travels on the wire.
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	mac := hmac.New(sha256.New, []byte(s.botFormsSecret))
	mac.Write([]byte(timestamp + "."))
	mac.Write(body)

	endpoint := strings.TrimRight(s.botFormsURL, "/") + "/api/forms/" + botFormsSite
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forms-Timestamp", timestamp)
	req.Header.Set("X-Forms-Signature", "sha256="+hex.EncodeToString(mac.Sum(nil)))

	resp, err := botFormsClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// A 4xx will not be fixed by retrying (wrong secret or bad payload).
	if resp.StatusCode >= 500 {
		return fmt.Errorf("bot responded %d", resp.StatusCode)
	}
	if resp.StatusCode >= 400 {
		log.Printf("bot-forms: rejected by bot: %d", resp.StatusCode)
	}
	return nil
}
