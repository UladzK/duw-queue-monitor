package queuemonitor

// Notifier defines the interface for sending notifications about queue status updates.
type Notifier interface {
	// SendMessage sends a message to a specified chat ID
	SendMessage(chatID, text string) error
}
