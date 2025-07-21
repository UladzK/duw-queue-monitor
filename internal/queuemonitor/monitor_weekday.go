package queuemonitor

import (
	"time"
	"uladzk/duw_kolejka_checker/internal/logger"
)

type WeekdayQueueMonitor struct {
	decoratedMonitor QueueMonitorInterface
	timeProvider     TimeProvider
	log              *logger.Logger
}

type TimeProvider interface {
	Now() time.Time
}

func NewWeekdayQueueMonitor(decoratedMonitor QueueMonitorInterface, timeProvider TimeProvider, log *logger.Logger) *WeekdayQueueMonitor {
	return &WeekdayQueueMonitor{
		decoratedMonitor: decoratedMonitor,
		log:              log,
		timeProvider:     timeProvider,
	}
}

func (w *WeekdayQueueMonitor) Init(initState *MonitorState) {
	w.decoratedMonitor.Init(initState)
}

func (w *WeekdayQueueMonitor) GetState() *MonitorState {
	return w.decoratedMonitor.GetState()
}

func (w *WeekdayQueueMonitor) CheckAndProcessStatus() error {
	if w.isDuwOffTime() {
		w.log.Debug("Queue monitoring is disabled on weekends and off hours (06:00 - 18:00 UTC), skipping status check")
		return nil
	}

	return w.decoratedMonitor.CheckAndProcessStatus()
}

func (w *WeekdayQueueMonitor) isDuwOffTime() bool {
	now := w.timeProvider.Now().UTC()

	if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
		return true
	}

	if now.Hour() < 6 || now.Hour() >= 17 {
		return true
	}

	return false
}
