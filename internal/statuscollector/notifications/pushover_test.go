package notifications

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"uladzk/duw_kolejka_checker/internal/logger"
)

// todo: rewrite using table tests
func TestSendGeneralQueueStatusUpdatePush_WhenAvailableMessage_SendsNotificationToPushOverApiWithCorrectFormatAndTemplate(t *testing.T) {
	// Arrange
	mockPushOverApi := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, fmt.Sprintf("Expected HTTP POST but got %v", r.Method), http.StatusInternalServerError)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, fmt.Sprintf("Failed to parse form data: %v", err), http.StatusInternalServerError)
			return
		}

		token := r.FormValue("token")
		if token != "test-token" {
			http.Error(w, fmt.Sprintf("Expected token to be 'test-token' but got '%s'", token), http.StatusInternalServerError)
			return
		}

		user := r.FormValue("user")
		if user != "test-user" {
			http.Error(w, fmt.Sprintf("Expected user to be 'test-user' but got '%s'", user), http.StatusInternalServerError)
			return
		}

		message := r.FormValue("message")
		expectedMessage := "🔔 Kolejka **test-queue** jest teraz dostępna!\n🎟️ Ostatni przywołany bilet: **K80**\n🧾 Pozostało biletów: **10**"

		if message != expectedMessage {
			http.Error(w, fmt.Sprintf("Expected message to be \n'%s' but got \n'%s'", expectedMessage, message), http.StatusInternalServerError)
			return
		}

		fmt.Fprintln(w, `{"status": 200}`)
	}))

	defer mockPushOverApi.Close()

	cfg := &PushOverConfig{
		Token:  "test-token",
		User:   "test-user",
		ApiUrl: mockPushOverApi.URL,
	}

	logger := logger.NewLogger(&logger.Config{
		Level: "error"})

	sut := NewPushOverNotifier(cfg, logger, &http.Client{})

	// Act
	err := sut.SendGeneralQueueStatusUpdateNotification("test-queue", true, true, "K80", 10)

	// Assert
	if err != nil {
		t.Fatalf("Expected successful notification sending, but got error: \"%v\"", err)
	}
}

func TestSendGeneralQueueStatusUpdatePush_WhenUnavailableMessage_SendsNotificationToPushOverApiWithCorrectFormatAndTemplate(t *testing.T) {
	// Arrange
	mockPushOverApi := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, fmt.Sprintf("Expected HTTP POST but got %v", r.Method), http.StatusInternalServerError)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, fmt.Sprintf("Failed to parse form data: %v", err), http.StatusInternalServerError)
			return
		}

		token := r.FormValue("token")
		if token != "test-token" {
			http.Error(w, fmt.Sprintf("Expected token to be 'test-token' but got '%s'", token), http.StatusInternalServerError)
			return
		}

		user := r.FormValue("user")
		if user != "test-user" {
			http.Error(w, fmt.Sprintf("Expected user to be 'test-user' but got '%s'", user), http.StatusInternalServerError)
			return
		}

		message := r.FormValue("message")
		expectedMessage := "💤 Kolejka **test-queue** jest obecnie niedostępna."

		if message != expectedMessage {
			http.Error(w, fmt.Sprintf("Expected message to be \n'%s' but got \n'%s'", expectedMessage, message), http.StatusInternalServerError)
			return
		}

		fmt.Fprintln(w, `{"status": 200}`)
	}))

	defer mockPushOverApi.Close()

	cfg := &PushOverConfig{
		Token:  "test-token",
		User:   "test-user",
		ApiUrl: mockPushOverApi.URL,
	}

	logger := logger.NewLogger(&logger.Config{
		Level: "error"})

	sut := NewPushOverNotifier(cfg, logger, &http.Client{})

	// Act
	err := sut.SendGeneralQueueStatusUpdateNotification("test-queue", true, false, "K80", 10)

	// Assert
	if err != nil {
		t.Fatalf("Expected successful notification sending, but got error: \"%v\"", err)
	}
}

func TestSendGeneralQueueStatusUpdatePush_WhenSendNotificationFailed_ReturnsError(t *testing.T) {
	// Arrange
	mockPushOverApi := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "mocked internal error", http.StatusInternalServerError)
	}))

	defer mockPushOverApi.Close()

	cfg := &PushOverConfig{
		Token:  "test-token",
		User:   "test-user",
		ApiUrl: mockPushOverApi.URL,
	}

	logger := logger.NewLogger(&logger.Config{
		Level: "error"})

	sut := NewPushOverNotifier(cfg, logger, &http.Client{})

	// Act
	err := sut.SendGeneralQueueStatusUpdateNotification("test-queue", true, true, "K80", 10)

	// Assert
	if err == nil {
		t.Fatal("Expected error to be returned but go no one.")
	}
}
