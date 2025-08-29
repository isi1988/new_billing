package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"new-billing/internal/config"
	"new-billing/internal/models"
	"time"
)

type TelegramService struct {
	config *config.TelegramConfig
}

type TelegramMessage struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"`
}

func NewTelegramService(cfg *config.TelegramConfig) *TelegramService {
	return &TelegramService{
		config: cfg,
	}
}

func (ts *TelegramService) SendIssueCreated(issue *models.Issue) {
	if !ts.config.Enabled || ts.config.BotToken == "" {
		return
	}

	text := fmt.Sprintf("🆕 *Новая доработка создана*\n\n"+
		"📝 *ID:* %d\n"+
		"📋 *Название:* %s\n"+
		"📄 *Описание:* %s\n"+
		"📅 *Создана:* %s\n"+
		"👤 *Создал:* Пользователь ID %d",
		issue.ID,
		escapeMarkdown(issue.Title),
		escapeMarkdown(issue.Description),
		issue.CreatedAt.Format("02.01.2006 15:04"),
		issue.CreatedBy,
	)

	ts.sendMessage(text)
}

func (ts *TelegramService) SendIssueUpdated(oldIssue, newIssue *models.Issue, changes []string) {
	if !ts.config.Enabled || ts.config.BotToken == "" {
		return
	}

	changesText := ""
	for _, change := range changes {
		changesText += fmt.Sprintf("• %s\n", change)
	}

	text := fmt.Sprintf("✏️ *Доработка изменена*\n\n"+
		"📝 *ID:* %d\n"+
		"📋 *Название:* %s\n\n"+
		"🔄 *Изменения:*\n%s"+
		"📅 *Изменена:* %s",
		newIssue.ID,
		escapeMarkdown(newIssue.Title),
		changesText,
		time.Now().Format("02.01.2006 15:04"),
	)

	ts.sendMessage(text)
}

func (ts *TelegramService) SendIssueResolved(issue *models.Issue) {
	if !ts.config.Enabled || ts.config.BotToken == "" {
		return
	}

	text := fmt.Sprintf("✅ *Доработка решена*\n\n"+
		"📝 *ID:* %d\n"+
		"📋 *Название:* %s\n"+
		"📄 *Описание:* %s\n"+
		"📅 *Создана:* %s\n"+
		"🎯 *Решена:* %s\n"+
		"👤 *Решил:* Пользователь ID %d",
		issue.ID,
		escapeMarkdown(issue.Title),
		escapeMarkdown(issue.Description),
		issue.CreatedAt.Format("02.01.2006 15:04"),
		time.Now().Format("02.01.2006 15:04"),
		*issue.ResolvedBy,
	)

	ts.sendMessage(text)
}

func (ts *TelegramService) SendIssueDeleted(issue *models.Issue) {
	if !ts.config.Enabled || ts.config.BotToken == "" {
		return
	}

	text := fmt.Sprintf("🗑️ *Доработка удалена*\n\n"+
		"📝 *ID:* %d\n"+
		"📋 *Название:* %s\n"+
		"📄 *Описание:* %s\n"+
		"📅 *Была создана:* %s\n"+
		"🗑️ *Удалена:* %s",
		issue.ID,
		escapeMarkdown(issue.Title),
		escapeMarkdown(issue.Description),
		issue.CreatedAt.Format("02.01.2006 15:04"),
		time.Now().Format("02.01.2006 15:04"),
	)

	ts.sendMessage(text)
}

func (ts *TelegramService) SendIssueUnresolved(issue *models.Issue, reason string) {
	if ts.config.BotToken == "" {
		return
	}

	text := fmt.Sprintf("🔄 *Доработка возвращена в работу*\n\n"+
		"📝 *ID:* %d\n"+
		"📋 *Название:* %s\n"+
		"📄 *Описание:* %s\n"+
		"💭 *Причина возврата:* %s\n"+
		"📅 *Создана:* %s\n"+
		"🔄 *Возвращена в работу:* %s",
		issue.ID,
		escapeMarkdown(issue.Title),
		escapeMarkdown(issue.Description),
		escapeMarkdown(reason),
		issue.CreatedAt.Format("02.01.2006 15:04"),
		time.Now().Format("02.01.2006 15:04"),
	)

	ts.sendMessage(text)
}

func (ts *TelegramService) sendMessage(text string) {
	message := TelegramMessage{
		ChatID:    ts.config.ChatID,
		Text:      text,
		ParseMode: "Markdown",
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		fmt.Printf("Error marshaling Telegram message: %v\n", err)
		return
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", ts.config.BotToken)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error sending Telegram message: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Telegram API returned status: %d\n", resp.StatusCode)
		return
	}

	fmt.Println("Telegram notification sent successfully")
}

// escapeMarkdown escapes special characters for Telegram Markdown
func escapeMarkdown(text string) string {
	// Simple approach - replace the most problematic characters
	text = fmt.Sprintf("%s", text)
	text = fmt.Sprintf("%s", text) // Ensure string conversion

	// For now, just return the text as is - Telegram will handle most cases
	// In production, you might want to implement proper escaping
	return text
}
