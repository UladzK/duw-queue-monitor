package statuscollector

import (
	"fmt"
	"os"
	"strconv"

	"time"
)

type Handler struct {
	isStateInitialized  bool
	queueActive         bool
	queueEnabled        bool
	lastTicketProcessed string
	ticketsLeft         int
}

func Run() {
	// sendPushesAlways, _ := strconv.ParseBool(os.Getenv("SEND_PUSHES_ALWAYS"))                 // TODO: follow up on err handling, TODO: move to Config
	statusCheckIntervalSeconds, _ := strconv.Atoi(os.Getenv("STATUS_CHECK_INTERVAL_SECONDS")) // TODO: follow up on err handling, TODO: move to Config

	handler := Handler{
		isStateInitialized: false,
	}

	for {
		if err := handler.checkAndProcessStatus(); err != nil {
			fmt.Printf("err during collecting status and pushing notifications: %v\n", err) // TODO: use logging
		}

		fmt.Printf("[%v] Checking again in %v seconds...\n", time.Now(), statusCheckIntervalSeconds) // TODO: use logging
		time.Sleep(time.Duration(statusCheckIntervalSeconds) * time.Second)                          // TODO: ticket is the better option?
	}
}

func (h *Handler) checkAndProcessStatus() error {

	newQueueStatus, err := getQueueStatus()
	if err != nil {
		return err
	}

	if !h.isStateInitialized || h.statusChanged(newQueueStatus) {
		if err := pushQueueEnabledNotification(newQueueStatus); err != nil {
			return err
		}

		h.updateState(newQueueStatus)

		return nil
	}

	if !newQueueStatus.Enabled {
		return nil
	}

	if h.ticketsLeft != newQueueStatus.TicketsLeft {
		if err := pushQueueEnabledNotification(newQueueStatus); err != nil {
			return err
		}

		h.updateState(newQueueStatus)
	}

	return nil
}

func (h *Handler) statusChanged(newQueueStatus *Queue) bool {
	return h.queueActive != newQueueStatus.Active || h.queueEnabled != newQueueStatus.Enabled
}

func pushQueueEnabledNotification(newQueueStatus *Queue) error {
	if err := sendGeneralQueueStatusUpdatePush(newQueueStatus.Name, newQueueStatus.Enabled, newQueueStatus.TicketValue, newQueueStatus.TicketsLeft); err != nil {
		return fmt.Errorf("error sending queue enabled notifiication: %w", err)
	}

	return nil
}

func (h *Handler) updateState(newQueueStatus *Queue) {
	h.isStateInitialized = true
	h.lastTicketProcessed = newQueueStatus.TicketValue
	h.queueEnabled = newQueueStatus.Enabled
	h.queueActive = newQueueStatus.Active
}
