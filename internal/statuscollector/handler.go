package statuscollector

import (
	"context"
	"fmt"
	"uladzk/duw_kolejka_checker/internal/logger"
	"uladzk/duw_kolejka_checker/internal/statuscollector/notifications"

	"time"
)

type Handler struct {
	cfg       *Config
	log       *logger.Logger
	collector *StatusCollector
	notifier  *notifications.PushOverNotifier
	state     State
}

type State struct {
	isStateInitialized  bool
	queueActive         bool
	queueEnabled        bool
	lastTicketProcessed string
	ticketsLeft         int
}

func NewHandler(cfg *Config, log *logger.Logger) *Handler {
	return &Handler{
		cfg:       cfg,
		log:       log,
		collector: NewStatusCollector(&cfg.StatusCollector),
		notifier:  notifications.NewPushOverNotifier(&cfg.NotificationPushOver, log),
		state: State{
			isStateInitialized: false,
		},
	}
}

func (h *Handler) Run(ctx context.Context, done chan<- bool) {
	for {
		select {
		case <-ctx.Done():
			h.log.Info("Received shutdown signal, exiting...")
			done <- true
			return
		default:
			if err := h.checkAndProcessStatus(); err != nil {
				h.log.Error("Error during collecting status and pushing notifications", err)
			}

			h.log.Debug(fmt.Sprintf("Status collection is completed. Checking again in %v seconds", h.cfg.StatusCheckInternalSeconds))
			time.Sleep(time.Duration(h.cfg.StatusCheckInternalSeconds) * time.Second) // TODO: will be sleeping even after SIGTERM. ticket is the better option?
		}
	}
}

func (h *Handler) checkAndProcessStatus() error {

	newState, err := h.collector.GetQueueStatus()
	if err != nil {
		return fmt.Errorf("error getting queue status: %w", err)
	}

	if !h.state.isStateInitialized || h.statusChanged(newState) {
		if err := h.pushQueueEnabledNotification(newState); err != nil {
			return err
		}

		h.updateState(newState)

		return nil
	}

	if !newState.Enabled {
		return nil
	}

	if h.state.ticketsLeft != newState.TicketsLeft {
		if err := h.pushQueueEnabledNotification(newState); err != nil {
			return err
		}

		h.updateState(newState)
	}

	return nil
}

func (h *Handler) statusChanged(newQueueStatus *Queue) bool {
	return h.state.queueActive != newQueueStatus.Active || h.state.queueEnabled != newQueueStatus.Enabled
}

func (h *Handler) pushQueueEnabledNotification(newQueueStatus *Queue) error {
	if err := h.notifier.SendGeneralQueueStatusUpdatePush(newQueueStatus.Name, newQueueStatus.Enabled, newQueueStatus.TicketValue, newQueueStatus.TicketsLeft); err != nil {
		return fmt.Errorf("error sending queue enabled notifiication: %w", err)
	}

	return nil
}

func (h *Handler) updateState(newQueueStatus *Queue) {
	h.state.isStateInitialized = true
	h.state.lastTicketProcessed = newQueueStatus.TicketValue
	h.state.queueEnabled = newQueueStatus.Enabled
	h.state.queueActive = newQueueStatus.Active
}
