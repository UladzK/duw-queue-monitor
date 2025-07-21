package queuemonitor

import "time"

// SystemDateTimeProvider provides the current system time.
// Can be moved to a shared package if needed elsewhere.
type SystemDateTimeProvider struct{}

func NewSystemDateTimeProvider() *SystemDateTimeProvider {
	return &SystemDateTimeProvider{}
}

func (r *SystemDateTimeProvider) Now() time.Time {
	return time.Now()
}
