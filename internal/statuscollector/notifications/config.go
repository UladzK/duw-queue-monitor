package notifications

type PushOverConfig struct {
	ApiUrl string `env:"NOTIFICATION_PUSHOVER_API_URL" envDefault:"https://api.pushover.net/1/messages.json"`
	Token  string `env:"NOTIFICATION_PUSHOVER_TOKEN,required"` // aay6otxvgv5zwkwck6r6r6bch4qucs
	User   string `env:"NOTIFICATION_PUSHOVER_USER,required"`  // eun179bk9o34gn7tg3qk8s4jt8d4i5
}
