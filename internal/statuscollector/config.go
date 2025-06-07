package statuscollector

import "uladzk/duw_kolejka_checker/internal/statuscollector/notifications"

type Config struct {
	StatusCheckInternalSeconds int  `env:"STATUS_CHECK_INTERVAL_SECONDS" envDefault:"10"`
	UseTelegramNotifications   bool `env:"USE_TELEGRAM_NOTIFICATIONS" envDefault:"false"`
	StatusCollector            StatusCollectorConfig
	NotificationPushOver       notifications.PushOverConfig
	NotificationTelegram       notifications.TelegramConfig
}

type StatusCollectorConfig struct {
	StatusApiUrl string `env:"STATUS_API_URL" envDefault:"https://rezerwacje.duw.pl/status_kolejek/query.php?status="`
}
