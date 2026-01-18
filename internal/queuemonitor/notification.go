package queuemonitor

import (
	"context"
	"fmt"
)

// Notifier defines the interface for sending notifications about queue status updates.
type Notifier interface {
	// SendMessage sends a message to a specified chat ID
	SendMessage(ctx context.Context, chatID, text string) error
}

// sendNotification sends a notification about the queue status during state transitions.
func sendNotification(ctx context.Context, notifier Notifier, channelName string, queue *Queue, isInactive bool) error {
	chatID := fmt.Sprintf("@%s", channelName)
	var message string
	if isInactive {
		message = buildQueueInactiveMsg(queue.Name)
	} else {
		message = buildQueueAvailableMsg(queue.Name, queue.Enabled, queue.TicketValue, queue.TicketsLeft)
	}
	if err := notifier.SendMessage(ctx, chatID, message); err != nil {
		return fmt.Errorf("error sending queue notification: %w", err)
	}
	return nil
}
