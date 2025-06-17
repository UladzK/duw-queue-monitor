package statuscollector

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"uladzk/duw_kolejka_checker/internal/logger"

	"github.com/google/go-cmp/cmp"
)

type mockNotifier struct {
	shouldFail     bool // used to simulate failure in sending notification
	called         bool
	lastSentStatus *Queue
}

func (f *mockNotifier) SendGeneralQueueStatusUpdateNotification(queueName string, active bool, enabled bool, actualTicket string, numberOfTicketsLeft int) error {
	f.called = true

	if f.shouldFail {
		return fmt.Errorf("failed to send notification")
	}

	f.lastSentStatus = &Queue{
		Name:        queueName,
		Active:      active,
		Enabled:     enabled,
		TicketValue: actualTicket,
		TicketsLeft: numberOfTicketsLeft,
	}
	return nil
}

func TestCheckAndProcessStatus_WhenStateIsNotInitialized_CorrectlyHandlesStateTransition(t *testing.T) {
	// Arrange
	const queueName = "Odbior karty"
	testConditions := []struct {
		name                     string
		newState                 Queue
		notificationShouldBeSent bool
		expectedNotification     *Queue
	}{
		{
			"Condition 1: \"queue is not active, state was not initialized, queue becomes active.\" Expected: \"notification should not be sent.\"",
			Queue{Name: queueName, Active: false, Enabled: false, TicketValue: "", TicketsLeft: 0},
			false,
			nil,
		},
		{
			"Condition 2: \"queue is active, state was not initialized.\" Expected: \"notification should be sent.\"",
			Queue{Name: queueName, Active: true, Enabled: true, TicketValue: "K123", TicketsLeft: 10},
			true,
			&Queue{Name: queueName, Active: true, Enabled: true, TicketValue: "K123", TicketsLeft: 10},
		},
	}

	for _, tc := range testConditions {
		t.Run(tc.name, func(t *testing.T) {

			mockDuwApi := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, `{
					"result": {
						"Wrocław": [{
							"id": 24,
							"name": "%v",
							"ticket_value": "%v",
							"tickets_left": %v,
							"active": %v,
							"enabled": %v
						}]
					}
				}`, queueName,
					tc.newState.TicketValue,
					tc.newState.TicketsLeft,
					tc.newState.Active,
					tc.newState.Enabled)
			}))
			defer mockDuwApi.Close()

			cfg := &Config{
				QueueMonitor: QueueMonitorConfig{
					StatusApiUrl: mockDuwApi.URL,
				},
			}

			collector := NewStatusCollector(&cfg.QueueMonitor, &http.Client{})
			logger := logger.NewLogger(&logger.Config{
				Level: "error"})

			notifier := &mockNotifier{}

			sut := NewQueueMonitor(cfg, logger, collector, notifier)
			expectedFinalState := &MonitorState{
				QueueActive:         tc.newState.Active,
				QueueEnabled:        tc.newState.Enabled,
				LastTicketProcessed: tc.newState.TicketValue,
				TicketsLeft:         tc.newState.TicketsLeft,
			}

			// Act
			err := sut.CheckAndProcessStatus()

			// Assert
			if err != nil {
				t.Fatalf("Expected successful execution, but execution returned error: %v", err)
			}

			if notifier.called != tc.notificationShouldBeSent {
				t.Errorf("Expected notification sending: %v, but it was: %v", tc.notificationShouldBeSent, notifier.called)
			}

			if diff := cmp.Diff(notifier.lastSentStatus, tc.expectedNotification); diff != "" {
				t.Errorf("Notification mismatch (-want +got):\n%s", diff)
			}

			if stateDiff := cmp.Diff(sut.GetState(), expectedFinalState); stateDiff != "" {
				t.Errorf("State mismatch between currently set state of monitor and latest state (-want +got):\n%s", stateDiff)
			}
		})
	}
}

func TestCheckAndProcessStatus_WhenStateIsInitialized_CorrectlyHandlesStrateTransition(t *testing.T) {
	// Arrange
	const queueName = "Odbior karty"
	testConditions := []struct {
		name                     string
		isStateInitialized       bool
		initialState             MonitorState
		newState                 Queue
		notificationShouldBeSent bool
		expectedNotification     *Queue
	}{
		{
			"Condition 1: \"queue was active, state was initialized, no changes.\" Expected: \"notification shoud NOT be sent.\"",
			true,
			MonitorState{QueueActive: true, QueueEnabled: true, TicketsLeft: 10, LastTicketProcessed: "K123"},
			Queue{Name: queueName, Active: true, Enabled: true, TicketValue: "K123", TicketsLeft: 10},
			false,
			nil,
		},
		{
			"Condition 2: \"queue was active, state was initialized, queue becomes not active.\" Expected: \"notification should be sent.\"",
			true,
			MonitorState{QueueActive: true, QueueEnabled: true, TicketsLeft: 10, LastTicketProcessed: "K123"},
			Queue{Name: queueName, Active: false, Enabled: true, TicketValue: "K123", TicketsLeft: 10},
			true,
			&Queue{Name: queueName, Active: false, Enabled: true, TicketValue: "K123", TicketsLeft: 10},
		},
		{
			"Condition 3: \"queue was active, state was initialized, queue remains active, status becomes not enabled.\" Expected: \"notification should be sent.\"",
			true,
			MonitorState{QueueActive: true, QueueEnabled: true, TicketsLeft: 10, LastTicketProcessed: "K123"},
			Queue{Name: queueName, Active: true, Enabled: false, TicketValue: "K123", TicketsLeft: 0},
			true,
			&Queue{Name: queueName, Active: true, Enabled: false, TicketValue: "K123", TicketsLeft: 0},
		},
		{
			"Condition 4: \"queue was active, state was initialized, queue remains active and enabled, ticket left changed.\" Expected: \"notification should be sent.\"",
			true,
			MonitorState{QueueActive: true, QueueEnabled: true, TicketsLeft: 10, LastTicketProcessed: "K123"},
			Queue{Name: queueName, Active: true, Enabled: true, TicketValue: "K123", TicketsLeft: 5},
			true,
			&Queue{Name: queueName, Active: true, Enabled: true, TicketValue: "K123", TicketsLeft: 5},
		},
		{
			"Condition 5: \"queue was active, state was initialized, queue remains active and enabled, only ticket value changed.\" Expected: \"notification should NOT be sent.\"",
			true,
			MonitorState{QueueActive: true, QueueEnabled: true, TicketsLeft: 10, LastTicketProcessed: "K123"},
			Queue{Name: queueName, Active: true, Enabled: true, TicketValue: "K456", TicketsLeft: 10},
			false,
			nil,
		},
		{
			"Condition 6: \"queue was active, state was initialized, queue remains active and enabled, ticket value is empty and not changed.\" Expected: \"notification should NOT be sent.\"",
			true,
			MonitorState{QueueActive: true, QueueEnabled: true, TicketsLeft: 100, LastTicketProcessed: ""},
			Queue{Name: queueName, Active: true, Enabled: true, TicketsLeft: 100, TicketValue: ""},
			false,
			nil,
		},
	}

	for _, tc := range testConditions {
		t.Run(tc.name, func(t *testing.T) {

			mockDuwApi := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, `{
					"result": {
						"Wrocław": [{
							"id": 24,
							"name": "%v",
							"ticket_value": "%v",
							"tickets_left": %v,
							"active": %v,
							"enabled": %v
						}]
					}
				}`, queueName,
					tc.newState.TicketValue,
					tc.newState.TicketsLeft,
					tc.newState.Active,
					tc.newState.Enabled)
			}))
			defer mockDuwApi.Close()

			cfg := &Config{
				QueueMonitor: QueueMonitorConfig{
					StatusApiUrl: mockDuwApi.URL,
				},
			}

			collector := NewStatusCollector(&cfg.QueueMonitor, &http.Client{})
			logger := logger.NewLogger(&logger.Config{Level: "error"})

			notifier := &mockNotifier{}

			sut := NewQueueMonitor(cfg, logger, collector, notifier)
			sut.Init(&tc.initialState)
			expectedFinalState := &MonitorState{
				QueueActive:         tc.newState.Active,
				QueueEnabled:        tc.newState.Enabled,
				TicketsLeft:         tc.newState.TicketsLeft,
				LastTicketProcessed: tc.newState.TicketValue,
			}

			// Act
			err := sut.CheckAndProcessStatus()

			// Assert
			if err != nil {
				t.Fatalf("Expected successful execution, but execution returned error: %v", err)
			}

			if notifier.called != tc.notificationShouldBeSent {
				t.Errorf("Expected notification sending: %v, but it was: %v", tc.notificationShouldBeSent, notifier.called)
			}

			if diff := cmp.Diff(notifier.lastSentStatus, tc.expectedNotification); diff != "" {
				t.Errorf("Notification mismatch (-want +got):\n%s", diff)
			}

			if diffState := cmp.Diff(sut.GetState(), expectedFinalState); diffState != "" {
				t.Errorf("State mismatch between currently set state of monitor and latest state (-want +got):\n%s", diffState)
			}
		})
	}
}

func TestCheckAndProcessStatus_WhenCollectingQueueStatusFailed_DoesNotPushNotificationAndReturnsError(t *testing.T) {
	// Arrange
	mockDuwApi := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))

	defer mockDuwApi.Close()

	cfg := &Config{
		QueueMonitor: QueueMonitorConfig{
			StatusApiUrl: mockDuwApi.URL,
		},
	}

	collector := NewStatusCollector(&cfg.QueueMonitor, &http.Client{})
	logger := logger.NewLogger(&logger.Config{
		Level: "error"})

	notifier := &mockNotifier{shouldFail: true}

	sut := NewQueueMonitor(cfg, logger, collector, notifier)
	sut.Init(&MonitorState{
		QueueActive:         true,
		QueueEnabled:        true,
		TicketsLeft:         10,
		LastTicketProcessed: "K123",
	})

	// Act
	err := sut.CheckAndProcessStatus()

	// Assert
	if err == nil {
		t.Fatal("Expected error to be returned, but there is no one.", err)
	}

	if notifier.called {
		t.Errorf("Expected no notification to be sent.\", but there was one %+v", notifier.lastSentStatus)
	}
}

func TestCheckAndProcessStatus_WhenPushNotificationFailed_ReturnsError(t *testing.T) {
	// Arrange
	mockDuwApi := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{
			"result": {
				"Wrocław": [{
					"id": 24,
					"name": "Odbior karty",
					"ticket_value": "K123",
					"tickets_left": 0,
					"active": true,
					"enabled": true
				}]
			}
		}`)
	}))
	defer mockDuwApi.Close()

	cfg := &Config{
		QueueMonitor: QueueMonitorConfig{
			StatusApiUrl: mockDuwApi.URL,
		},
	}

	collector := NewStatusCollector(&cfg.QueueMonitor, &http.Client{})
	logger := logger.NewLogger(&logger.Config{
		Level: "error"})

	notifier := &mockNotifier{shouldFail: true}

	sut := NewQueueMonitor(cfg, logger, collector, notifier)
	sut.isStateInitialized = true
	sut.state.QueueActive = true
	sut.state.QueueEnabled = true
	sut.state.TicketsLeft = 10

	// Act
	err := sut.CheckAndProcessStatus()

	// Assert
	if err == nil {
		t.Fatal("Expected error to be returned, but there is no one.", err)
	}

	if !notifier.called {
		t.Error("Expected notification to be sent, but it wasn't")
	}
}
