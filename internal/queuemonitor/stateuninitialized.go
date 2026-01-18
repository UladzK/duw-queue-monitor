package queuemonitor

import "context"

// UninitializedState represents the initial state before first check.
type UninitializedState struct {
	notifier    Notifier
	channelName string
}

func (s *UninitializedState) Name() string     { return "Uninitialized" }
func (s *UninitializedState) TicketsLeft() int { return 0 }

func (s *UninitializedState) Handle(ctx context.Context, queue *Queue) (QueueState, error) {
	if !queue.Active {
		return &InactiveState{notifier: s.notifier, channelName: s.channelName}, nil
	}

	// Queue has become active - always notify
	if err := sendNotification(ctx, s.notifier, s.channelName, queue, false); err != nil {
		return s, err
	}

	if queue.Enabled {
		return &ActiveEnabledState{notifier: s.notifier, channelName: s.channelName, ticketsLeft: queue.TicketsLeft}, nil
	}
	return &ActiveDisabledState{notifier: s.notifier, channelName: s.channelName}, nil
}
