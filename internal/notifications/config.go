package notifications

type TelegramConfig struct {
	BaseApiUrl string `env:"NOTIFICATION_TELEGRAM_API_BASE_URL" envDefault:"https://api.telegram.org"`
	BotToken   string `env:"NOTIFICATION_TELEGRAM_BOT_TOKEN,required"`
}
