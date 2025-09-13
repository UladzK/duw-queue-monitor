package notifications

type TelegramConfig struct {
	BaseApiUrl            string `env:"NOTIFICATION_TELEGRAM_API_BASE_URL" envDefault:"https://api.telegram.org"`
	BotToken              string `env:"NOTIFICATION_TELEGRAM_BOT_TOKEN,required"`
	MaxRetryAttempts      uint   `env:"NOTIFICATION_TELEGRAM_MAX_RETRY_ATTEMPTS" envDefault:"5"`
	RetryDelayMs          uint   `env:"NOTIFICATION_TELEGRAM_RETRY_DELAY_MS" envDefault:"500"`
	RequestTimeoutSeconds uint   `env:"NOTIFICATION_TELEGRAM_REQUEST_TIMEOUT_SECONDS" envDefault:"5"`
}
