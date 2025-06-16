package notifications

import "fmt"

const (
	msgQueueAvailableGeneral = "Queue %s is available! Actual ticket: %s. Tickets left: %d."
	msgQueueAvailableShort   = "Queue %s is available! Tickets left: %d."
	msgQueueUnavailable      = "Queue %s is unavailable."
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
