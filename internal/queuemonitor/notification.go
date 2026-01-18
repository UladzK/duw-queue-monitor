package queuemonitor

import (
	"context"
	"fmt"
)

// Message constants for queue status notifications
const (
	msgQueueAvailableGeneral = "🔔 Kolejka <b>%s</b> jest teraz dostępna!\n🎟️ Ostatni przywołany bilet: <b>%s</b>\n🧾 Pozostało biletów: <b>%d</b>"
	msgQueueAvailableShort   = "🔔 Kolejka <b>%s</b> jest teraz dostępna!\n🧾 Pozostało biletów: <b>%d</b>"
	msgQueueUnavailable      = "💤 Kolejka <b>%s</b> jest obecnie niedostępna."
	msgQueueInactive         = "🌙 Kolejka <b>%s</b> jest nieaktywna — prawdopodobnie koniec godzin pracy DUW."
)

// Notifier defines the interface for sending notifications about queue status updates.
type Notifier interface {
	// SendMessage sends a message to a specified chat ID
	SendMessage(ctx context.Context, chatID, text string) error
}

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
