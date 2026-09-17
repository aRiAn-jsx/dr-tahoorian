package notification

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/tahoorian/tahoorian/internal/domain"
)

// SendContact forwards a submitted contact form to a Telegram chat through a bot.
//
// Configuration (environment variables):
//   TELEGRAM_BOT_TOKEN  — bot token from @BotFather
//   TELEGRAM_CHAT_ID    — the chat/user/group id to receive the message
//
// When either value is missing the call is a no-op (returns nil), so enabling
// notifications is entirely optional and safe to leave unconfigured.
//
// Intended to be fired in the background (via `go`) so a slow Telegram call
// never blocks or fails the web form.
func SendContact(contact *domain.Contact) error {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")
	if token == "" || chatID == "" {
		slog.Info("telegram notification disabled (set TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID)")
		return nil
	}

	vals := url.Values{}
	vals.Add("chat_id", chatID)
	vals.Add("text", buildMessage(contact))
	vals.Add("disable_web_page_preview", "true")

	apiURL := "https://api.telegram.org/bot" + token + "/sendMessage"
	resp, err := http.PostForm(apiURL, vals)
	if err != nil {
		slog.Warn("telegram send failed", "error", err)
		return fmt.Errorf("telegram send failed: %w", err)
	}
	// Drain and close the body so the HTTP/1.x connection can be reused.
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Warn("telegram send returned bad status", "status", resp.StatusCode)
		return fmt.Errorf("telegram send bad status: %d", resp.StatusCode)
	}

	slog.Info("contact submitted to telegram", "chat_id", chatID)
	return nil
}

// buildMessage formats the contact nicely in Persian for the Telegram chat.
func buildMessage(c *domain.Contact) string {
	var lines []string
	lines = append(lines, "📩 پیام جدید از فرم تماس وب‌سایت")
	lines = append(lines, "")
	lines = append(lines, "👤 نام: " + c.Name)
	if c.Phone != "" {
		lines = append(lines, "📞 تلفن: " + c.Phone)
	}
	if c.Email != "" {
		lines = append(lines, "📧 ایمیل: " + c.Email)
	}
	if c.Subject != "" {
		lines = append(lines, "🗂 موضوع: " + c.Subject)
	}
	lines = append(lines, "")
	lines = append(lines, "✍️ پیام:")
	lines = append(lines, c.Message)
	lines = append(lines, "")
	lines = append(lines, "🕒 " + c.CreatedAt.Format("2006/01/02 15:04"))

	return strings.Join(lines, "\n")
}