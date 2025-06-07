package statuscollector

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"uladzk/duw_kolejka_checker/internal/logger"
)

type mockNotifier struct {
	shouldFail bool // used to simulate failure in sending notification
	called     bool
	lastStatus *Queue
}

func (f *mockNotifier) SendGeneralQueueStatusUpdatePush(queueName string, enabled bool, actualTicket string, numberOfTicketsLeft int) error {
	f.called = true

	if f.shouldFail {
		return fmt.Errorf("failed to send notification")
	}

	f.lastStatus = &Queue{
		Name:        queueName,
		Enabled:     enabled,
		TicketValue: actualTicket,
		TicketsLeft: numberOfTicketsLeft,
	}
	return nil
}

func TestCheckAndProcessStatus_WhenStateIsNotInitialized_AlwaysTriggersNotification(t *testing.T) {
	// Arrange
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

	notifier := &mockNotifier{}

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
	// Arrange
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

	notifier := &mockNotifier{}

	sut := NewQueueMonitor(cfg, logger, collector, notifier)
	sut.state.isStateInitialized = true
	sut.state.queueActive = false

	// Act
	err := sut.CheckAndProcessStatus()

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

	notifier := &mockNotifier{}

	sut := NewQueueMonitor(cfg, logger, collector, notifier)
	sut.state.isStateInitialized = true
	sut.state.queueActive = true
	sut.state.queueEnabled = false

	// Act
	err := sut.CheckAndProcessStatus()

	// Assert
	if err != nil {
		t.Fatalf("Expected successful execution, but execution returned error: %v", err)
	}
	if notifier.called {
		t.Errorf("Expected no notification to be sent, but there was one %+v", notifier.lastStatus)
	}
}

func TestCheckAndProcessStatus_WhenQueueEnabledAndTicketsLeftChanged_TriggersNotification(t *testing.T) {
	// Arrange

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

	notifier := &mockNotifier{}

	sut := NewQueueMonitor(cfg, logger, collector, notifier)
	sut.state.isStateInitialized = true
	sut.state.queueActive = true
	sut.state.queueEnabled = true
	sut.state.ticketsLeft = 10

	// Act
	err := sut.CheckAndProcessStatus()

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
		StatusCollector: StatusCollectorConfig{
			StatusApiUrl: mockDuwApi.URL,
		},
	}

	collector := NewStatusCollector(&cfg.StatusCollector, &http.Client{})
	logger := logger.NewLogger(&logger.Config{
		Level: "error"})

	notifier := &mockNotifier{shouldFail: true}

	sut := NewQueueMonitor(cfg, logger, collector, notifier)
	sut.state.isStateInitialized = true
	sut.state.queueActive = true
	sut.state.queueEnabled = true
	sut.state.ticketsLeft = 10

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

func TestCheckAndProcessStatus_WhenCollectingQueueStatusFailed_DoesNotPushNotificationAndReturnsError(t *testing.T) {
	// Arrange
	mockDuwApi := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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

	notifier := &mockNotifier{shouldFail: true}

	sut := NewQueueMonitor(cfg, logger, collector, notifier)
	sut.state.isStateInitialized = true
	sut.state.queueActive = true
	sut.state.queueEnabled = true
	sut.state.ticketsLeft = 10

	// Act
	err := sut.CheckAndProcessStatus()

	// Assert
	if err == nil {
		t.Fatal("Expected error to be returned, but there is no one.", err)
	}
	if notifier.called {
		t.Errorf("Expected no notification to be sent, but there was one %+v", notifier.lastStatus)
	}
}
