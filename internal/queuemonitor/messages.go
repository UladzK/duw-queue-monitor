package queuemonitor

import (
	"fmt"
)

// Message constants for queue status notifications
const (
	msgQueueAvailableGeneral = "🔔 Kolejka <b>%s</b> jest teraz dostępna!\n🎟️ Ostatni przywołany bilet: <b>%s</b>\n🧾 Pozostało biletów: <b>%d</b>"
	msgQueueAvailableShort   = "🔔 Kolejka <b>%s</b> jest teraz dostępna!\n🧾 Pozostało biletów: <b>%d</b>"
	msgQueueUnavailable      = "💤 Kolejka <b>%s</b> jest obecnie niedostępna."
	msgQueueInactive         = "💤 Kolejka <b>%s</b> jest teraz nieaktywna."
	parseMode                = "HTML"
)

// buildQueueAvailableMsg creates a formatted message based on queue status
func buildQueueAvailableMsg(queueName string, queueEnabled bool, actualTicket string, numberOfTicketsLeft int) string {
	if !queueEnabled {
		return fmt.Sprintf(msgQueueUnavailable, queueName)
	}

	if actualTicket == "" {
		return fmt.Sprintf(msgQueueAvailableShort, queueName, numberOfTicketsLeft)
	}
	return fmt.Sprintf(msgQueueAvailableGeneral, queueName, actualTicket, numberOfTicketsLeft)
}

func buildQueueInactiveMsg(queueName string) string {
	return fmt.Sprintf(msgQueueInactive, queueName)
}
