package queuemonitor

import "uladzk/duw_kolejka_checker/internal/notifications"

type Config struct {
	StatusCheckInternalSeconds int    `env:"STATUS_CHECK_INTERVAL_SECONDS" envDefault:"10"`
	BroadcastChannelName       string `env:"NOTIFICATION_TELEGRAM_BROADCAST_CHANNEL_NAME,required"`
	QueueMonitor               QueueMonitorConfig
	NotificationPushOver       notifications.PushOverConfig
	NotificationTelegram       notifications.TelegramConfig
}

type QueueMonitorConfig struct {
	StatusApiUrl    string `env:"STATUS_API_URL" envDefault:"https://rezerwacje.duw.pl/status_kolejek/query.php?status="`
	RedisConString  string `env:"STATE_REDIS_CONNECTION_STRING,required"`
	StateTtlSeconds int    `env:"STATE_TTL_SECONDS" envDefault:"60"`
}
