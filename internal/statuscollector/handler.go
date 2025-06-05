package statuscollector

import (
	"context"
	"fmt"
	"uladzk/duw_kolejka_checker/internal/statuscollector/notifications"

	"time"
)

type Handler struct {
	cfg       *Config
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

func NewHandler(cfg *Config) *Handler {
	return &Handler{
		cfg:       cfg,
		collector: NewStatusCollector(&cfg.StatusCollector),
		notifier:  notifications.NewPushOverNotifier(&cfg.NotificationPushOver),
		state: State{
			isStateInitialized: false,
		},
	}
}

func (h *Handler) Run(ctx context.Context, done chan<- bool) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Received shutdown signal, exiting...")
			done <- true
			return
		default:
			if err := h.checkAndProcessStatus(); err != nil {
				fmt.Printf("err during collecting status and pushing notifications: %v\n", err) // TODO: use logging
			}

			fmt.Printf("[%v] Checking again in %v seconds...\n", time.Now(), h.cfg.StatusCheckInternalSeconds) // TODO: use logging
			time.Sleep(time.Duration(h.cfg.StatusCheckInternalSeconds) * time.Second)                          // TODO: ticket is the better option?
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
