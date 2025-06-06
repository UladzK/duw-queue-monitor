package notifications

// Notifier defines the interface for sending notifications about queue status updates.
type Notifier interface {

	// SendGeneralQueueStatusUpdatePush sends a general notification about availability of the queue.
	// It includes the queue name, whether the queue is enabled, the actual ticket number and the number of tickets left.
	SendGeneralQueueStatusUpdatePush(queueName string, queueEnabled bool, actualTicket string, numberOfTicketsLeft int) error
}
