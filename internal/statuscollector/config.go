package statuscollector

type Config struct {
	StatusCheckInternalSeconds int `env:"STATUS_CHECK_INTERVAL_SECONDS" envDefault:"10"`
	StatusCollector            StatusCollectorConfig
	NotificationPushOver       PushOverNotificationConfig
}

type StatusCollectorConfig struct {
	StatusApiUrl string `env:"STATUS_API_URL" envDefault:"https://rezerwacje.duw.pl/status_kolejek/query.php?status="`
}

type PushOverNotificationConfig struct {
	ApiUrl string `env:"NOTIFICATION_PUSHOVER_API_URL" envDefault:"https://api.pushover.net/1/messages.json"`
	Token  string `env:"NOTIFICATION_PUSHOVER_TOKEN,required"` // aay6otxvgv5zwkwck6r6r6bch4qucs
	User   string `env:"NOTIFICATION_PUSHOVER_USER,required"`  // eun179bk9o34gn7tg3qk8s4jt8d4i5
}
