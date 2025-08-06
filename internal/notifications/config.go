package notifications

type PushOverConfig struct {
	ApiUrl string `env:"NOTIFICATION_PUSHOVER_API_URL" envDefault:"https://api.pushover.net/1/messages.json"`
	Token  string `env:"NOTIFICATION_PUSHOVER_TOKEN"`
	User   string `env:"NOTIFICATION_PUSHOVER_USER"`
}

type TelegramConfig struct {
	BaseApiUrl string `env:"NOTIFICATION_TELEGRAM_API_BASE_URL" envDefault:"https://api.telegram.org"`
	BotToken   string `env:"NOTIFICATION_TELEGRAM_BOT_TOKEN,required"`
}
