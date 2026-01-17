package queuemonitor

import "context"

// Notifier defines the interface for sending notifications about queue status updates.
type Notifier interface {
	// SendMessage sends a message to a specified chat ID
	SendMessage(ctx context.Context, chatID, text string) error
}
