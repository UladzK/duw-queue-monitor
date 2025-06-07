package notifications

type PushOverConfig struct {
	ApiUrl string `env:"NOTIFICATION_PUSHOVER_API_URL" envDefault:"https://api.pushover.net/1/messages.json"`
	Token  string `env:"NOTIFICATION_PUSHOVER_TOKEN,required"`
	User   string `env:"NOTIFICATION_PUSHOVER_USER,required"`
}
