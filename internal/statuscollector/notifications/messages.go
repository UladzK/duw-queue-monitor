package notifications

import "fmt"

const (
	msgQueueAvailable   = "Queue %s is available! Actual ticket: %s. Tickets left: %d."
	msgQueueUnavailable = "Queue %s is unavailable."
)

func buildQueueAvailableMsg(queueName string, queueEnabled bool, actualTicket string, numberOfTicketsLeft int) string {
	if !queueEnabled {
		return fmt.Sprintf(msgQueueUnavailable, queueName)
	}

	return fmt.Sprintf(msgQueueAvailable, queueName, actualTicket, numberOfTicketsLeft)
}
