package queuemonitor

import "context"

// ActiveEnabledState represents the state when queue is active (DUW working hours) and there are tickets available.
type ActiveEnabledState struct {
	notifier    Notifier
	channelName string
	ticketsLeft int
}

func (s *ActiveEnabledState) Name() string     { return "ActiveEnabled" }
func (s *ActiveEnabledState) TicketsLeft() int { return s.ticketsLeft }

func (s *ActiveEnabledState) Handle(ctx context.Context, queue *Queue) (QueueState, error) {
	if !queue.Active {
		if err := sendNotification(ctx, s.notifier, s.channelName, queue, true); err != nil {
			return s, err
		}
		return &InactiveState{notifier: s.notifier, channelName: s.channelName}, nil
	}

	if !queue.Enabled {
		if err := sendNotification(ctx, s.notifier, s.channelName, queue, false); err != nil {
			return s, err
		}
		return &ActiveDisabledState{notifier: s.notifier, channelName: s.channelName}, nil
	}

	// Still enabled - check if tickets changed
	if queue.TicketsLeft != s.ticketsLeft {
		if err := sendNotification(ctx, s.notifier, s.channelName, queue, false); err != nil {
			return s, err
		}
		return &ActiveEnabledState{notifier: s.notifier, channelName: s.channelName, ticketsLeft: queue.TicketsLeft}, nil
	}

	return s, nil
}
