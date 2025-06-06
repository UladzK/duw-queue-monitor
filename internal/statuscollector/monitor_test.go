package statuscollector

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"uladzk/duw_kolejka_checker/internal/logger"
)

type mockSuccessfulNotifier struct {
	called     bool
	lastStatus *Queue
}

func (f *mockSuccessfulNotifier) SendGeneralQueueStatusUpdatePush(queueName string, enabled bool, actualTicket string, numberOfTicketsLeft int) error {
	f.called = true
	f.lastStatus = &Queue{
		Name:        queueName,
		Enabled:     enabled,
		TicketValue: actualTicket,
		TicketsLeft: numberOfTicketsLeft,
	}
	return nil
}

func TestCheckAndProcessStatus_WhenStateIsNotInitialized_AlwaysTriggersNotification(t *testing.T) {
	// Arrange: mock DUW API response
	expectedNotification := &Queue{
		Name:        "Odbior karty",
		Enabled:     false,
		TicketValue: "K123",
		TicketsLeft: 10,
	}

	mockDuwApi := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{
			"result": {
				"Wrocław": [{
					"id": 24,
					"name": "Odbior karty",
					"ticket_value": "K123",
					"tickets_left": 10,
					"active": true,
					"enabled": false
				}]
			}
		}`)
	}))
	defer mockDuwApi.Close()

	cfg := &Config{
		StatusCollector: StatusCollectorConfig{
			StatusApiUrl: mockDuwApi.URL,
		},
	}

	collector := NewStatusCollector(&cfg.StatusCollector, &http.Client{})
	logger := logger.NewLogger(&logger.Config{
		Level: "error"})

	notifier := &mockSuccessfulNotifier{}

	monitor := NewQueueMonitor(cfg, logger, collector, notifier)

	// Act
	err := monitor.CheckAndProcessStatus()

	// Assert
	if err != nil {
		t.Fatalf("Expected successful execution, but execution returned error: %v", err)
	}
	if !notifier.called {
		t.Error("Expected notification to be sent, but it wasn't")
	}

	if notifier.lastStatus.Name != expectedNotification.Name {
		t.Errorf("Expected queue name %s, got %s", expectedNotification.Name, notifier.lastStatus.Name)
	}

	if notifier.lastStatus.TicketValue != expectedNotification.TicketValue {
		t.Errorf("Expected ticket %s, got %s", expectedNotification.TicketValue, notifier.lastStatus.TicketValue)
	}

	if notifier.lastStatus.TicketsLeft != expectedNotification.TicketsLeft {
		t.Errorf("Expected tickets left %d, got %d", expectedNotification.TicketsLeft, notifier.lastStatus.TicketsLeft)
	}

	if notifier.lastStatus.Enabled != expectedNotification.Enabled {
		t.Errorf("Expected queue enabled %t, got %t", expectedNotification.Enabled, notifier.lastStatus.Enabled)
	}
}

// TODO: make table tests for different states
func TestCheckAndProcessStatus_WhenQueueEnabledStateChanges_TriggersNotification(t *testing.T) {
	// Arrange: mock DUW API response
	expectedNotification := &Queue{
		Name:        "Odbior karty",
		Enabled:     true,
		TicketValue: "K123",
		TicketsLeft: 10,
	}

	mockDuwApi := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{
			"result": {
				"Wrocław": [{
					"id": 24,
					"name": "Odbior karty",
					"ticket_value": "K123",
					"tickets_left": 10,
					"active": true,
					"enabled": true
				}]
			}
		}`)
	}))
	defer mockDuwApi.Close()

	cfg := &Config{
		StatusCollector: StatusCollectorConfig{
			StatusApiUrl: mockDuwApi.URL,
		},
	}

	collector := NewStatusCollector(&cfg.StatusCollector, &http.Client{})
	logger := logger.NewLogger(&logger.Config{
		Level: "error"})

	notifier := &mockSuccessfulNotifier{}

	monitor := NewQueueMonitor(cfg, logger, collector, notifier)
	monitor.state.isStateInitialized = true
	monitor.state.queueActive = false

	// Act
	err := monitor.CheckAndProcessStatus()

	// Assert
	if err != nil {
		t.Fatalf("Expected successful execution, but execution returned error: %v", err)
	}
	if !notifier.called {
		t.Error("Expected notification to be sent, but it wasn't")
	}

	if notifier.lastStatus.Name != expectedNotification.Name {
		t.Errorf("Expected queue name %s, got %s", expectedNotification.Name, notifier.lastStatus.Name)
	}

	if notifier.lastStatus.TicketValue != expectedNotification.TicketValue {
		t.Errorf("Expected ticket %s, got %s", expectedNotification.TicketValue, notifier.lastStatus.TicketValue)
	}

	if notifier.lastStatus.TicketsLeft != expectedNotification.TicketsLeft {
		t.Errorf("Expected tickets left %d, got %d", expectedNotification.TicketsLeft, notifier.lastStatus.TicketsLeft)
	}

	if notifier.lastStatus.Enabled != expectedNotification.Enabled {
		t.Errorf("Expected queue enabled %t, got %t", expectedNotification.Enabled, notifier.lastStatus.Enabled)
	}
}

func TestCheckAndProcessStatus_WhenQueueDisabledAndStateNotChanged_DoesNotTriggerNotification(t *testing.T) {
	// Arrange: mock DUW API response
	mockDuwApi := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{
			"result": {
				"Wrocław": [{
					"id": 24,
					"name": "Odbior karty",
					"ticket_value": "K123",
					"tickets_left": 0,
					"active": true,
					"enabled": false
				}]
			}
		}`)
	}))
	defer mockDuwApi.Close()

	cfg := &Config{
		StatusCollector: StatusCollectorConfig{
			StatusApiUrl: mockDuwApi.URL,
		},
	}

	collector := NewStatusCollector(&cfg.StatusCollector, &http.Client{})
	logger := logger.NewLogger(&logger.Config{
		Level: "error"})

	notifier := &mockSuccessfulNotifier{}

	monitor := NewQueueMonitor(cfg, logger, collector, notifier)
	monitor.state.isStateInitialized = true
	monitor.state.queueActive = true
	monitor.state.queueEnabled = false

	// Act
	err := monitor.CheckAndProcessStatus()

	// Assert
	if err != nil {
		t.Fatalf("Expected successful execution, but execution returned error: %v", err)
	}
	if notifier.called {
		t.Errorf("Expected no notification to be sent, but there was one %+v", notifier.lastStatus)
	}
}

func TestCheckAndProcessStatus_WhenQueueEnabledAndJustTicketsLeftChanged_TriggersNotification(t *testing.T) {
	// Arrange: mock DUW API response

	expectedNotification := &Queue{
		Name:        "Odbior karty",
		Enabled:     true,
		TicketValue: "K123",
		TicketsLeft: 0,
	}

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
		StatusCollector: StatusCollectorConfig{
			StatusApiUrl: mockDuwApi.URL,
		},
	}

	collector := NewStatusCollector(&cfg.StatusCollector, &http.Client{})
	logger := logger.NewLogger(&logger.Config{
		Level: "error"})

	notifier := &mockSuccessfulNotifier{}

	monitor := NewQueueMonitor(cfg, logger, collector, notifier)
	monitor.state.isStateInitialized = true
	monitor.state.queueActive = true
	monitor.state.queueEnabled = true
	monitor.state.ticketsLeft = 10

	// Act
	err := monitor.CheckAndProcessStatus()

	// Assert
	if err != nil {
		t.Fatalf("Expected successful execution, but execution returned error: %v", err)
	}
	if !notifier.called {
		t.Error("Expected notification to be sent, but it wasn't")
	}

	if notifier.lastStatus.Name != expectedNotification.Name {
		t.Errorf("Expected queue name %s, got %s", expectedNotification.Name, notifier.lastStatus.Name)
	}

	if notifier.lastStatus.TicketValue != expectedNotification.TicketValue {
		t.Errorf("Expected ticket %s, got %s", expectedNotification.TicketValue, notifier.lastStatus.TicketValue)
	}

	if notifier.lastStatus.TicketsLeft != expectedNotification.TicketsLeft {
		t.Errorf("Expected tickets left %d, got %d", expectedNotification.TicketsLeft, notifier.lastStatus.TicketsLeft)
	}

	if notifier.lastStatus.Enabled != expectedNotification.Enabled {
		t.Errorf("Expected queue enabled %t, got %t", expectedNotification.Enabled, notifier.lastStatus.Enabled)
	}
}
