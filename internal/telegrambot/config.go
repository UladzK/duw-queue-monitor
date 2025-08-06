package telegrambot

import "uladzk/duw_kolejka_checker/internal/notifications"

type Config struct {
	FeedbackChatID       string `env:"NOTIFICATION_TELEGRAM_FEEDBACK_CHAT_ID,required"`
	NotificationTelegram notifications.TelegramConfig
}
