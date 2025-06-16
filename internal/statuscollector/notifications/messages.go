package notifications

import "fmt"

const (
	msgQueueAvailableGeneral = "🔔 Kolejka **%s** jest teraz dostępna!\n🎟️ Ostatni przywołany bilet: **%s**\n🧾 Pozostało biletów: **%d**"
	msgQueueAvailableShort   = "🔔 Kolejka **%s** jest teraz dostępna!\n🧾 Pozostało biletów: **%d**"
	msgQueueUnavailable      = "💤 Kolejka **%s** jest obecnie niedostępna."
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
