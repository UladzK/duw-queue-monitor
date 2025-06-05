package statuscollector

import (
	"fmt"
	"uladzk/duw_kolejka_checker/internal/statuscollector/notifications"

	"time"
)

type Handler struct {
	cfg                  *Config
	collectorService     *StatusCollectorService
	notificationsService *notifications.PushOverService
	isStateInitialized   bool
	queueActive          bool
	queueEnabled         bool
	lastTicketProcessed  string
	ticketsLeft          int
}

func NewHandler(cfg *Config) *Handler {
	return &Handler{
		cfg:                  cfg,
		collectorService:     NewStatusCollectorService(&cfg.StatusCollector),
		notificationsService: notifications.NewPushOverService(&cfg.NotificationPushOver),
		isStateInitialized:   false,
	}
}

func (h *Handler) Run() {
	for {
		if err := h.checkAndProcessStatus(); err != nil {
			fmt.Printf("err during collecting status and pushing notifications: %v\n", err) // TODO: use logging
		}

		fmt.Printf("[%v] Checking again in %v seconds...\n", time.Now(), h.cfg.StatusCheckInternalSeconds) // TODO: use logging
		time.Sleep(time.Duration(h.cfg.StatusCheckInternalSeconds) * time.Second)                          // TODO: ticket is the better option?
	}
}

func (h *Handler) checkAndProcessStatus() error {

	newQueueStatus, err := h.collectorService.getQueueStatus()
	if err != nil {
		return err
	}

	if !h.isStateInitialized || h.statusChanged(newQueueStatus) {
		if err := h.pushQueueEnabledNotification(newQueueStatus); err != nil {
			return err
		}

		h.updateState(newQueueStatus)

		return nil
	}

	if !newQueueStatus.Enabled {
		return nil
	}

	if h.ticketsLeft != newQueueStatus.TicketsLeft {
		if err := h.pushQueueEnabledNotification(newQueueStatus); err != nil {
			return err
		}

		h.updateState(newQueueStatus)
	}

	return nil
}

func (h *Handler) statusChanged(newQueueStatus *Queue) bool {
	return h.queueActive != newQueueStatus.Active || h.queueEnabled != newQueueStatus.Enabled
}

func (h *Handler) pushQueueEnabledNotification(newQueueStatus *Queue) error {
	if err := h.notificationsService.SendGeneralQueueStatusUpdatePush(newQueueStatus.Name, newQueueStatus.Enabled, newQueueStatus.TicketValue, newQueueStatus.TicketsLeft); err != nil {
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
