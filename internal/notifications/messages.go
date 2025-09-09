package notifications

import "fmt"

// TODO: move messages to monitor and telegrambot service which actually define them
const (
	msgQueueAvailableGeneral = "🔔 Kolejka <b>%s</b> jest teraz dostępna!\n🎟️ Ostatni przywołany bilet: <b>%s</b>\n🧾 Pozostało biletów: <b>%d</b>"
	msgQueueAvailableShort   = "🔔 Kolejka <b>%s</b> jest teraz dostępna!\n🧾 Pozostało biletów: <b>%d</b>"
	msgQueueUnavailable      = "💤 Kolejka <b>%s</b> jest obecnie niedostępna."
	parseMode                = "HTML"
)

func buildQueueAvailableMsg(queueName string, queueEnabled bool, actualTicket string, numberOfTicketsLeft int) string {
	if !queueEnabled {
		return fmt.Sprintf(msgQueueUnavailable, queueName)
	}

	if actualTicket == "" {
		return fmt.Sprintf(msgQueueAvailableShort, queueName, numberOfTicketsLeft)
	}
	return fmt.Sprintf(msgQueueAvailableGeneral, queueName, actualTicket, numberOfTicketsLeft)
}
