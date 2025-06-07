package notifications

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"uladzk/duw_kolejka_checker/internal/logger"
)

// TODO: refactor to use table tests
func TestSendGeneralQueueStatusUpdatePush_WhenAvailableMessage_SendsNotificationToTelegramApiWithCorrectFormatAndTemplate(t *testing.T) {
	// Arrange
	mockPushOverApi := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, fmt.Sprintf("Expected HTTP POST but got %v", r.Method), http.StatusInternalServerError)
			return
		}

		if r.URL.Path != "/bot123456/sendMessage" {
			http.Error(w, fmt.Sprintf("Expected URL to be '/bot123456/sendMessage' but got '%s'", r.URL.Path), http.StatusInternalServerError)
			return
		}

		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, fmt.Sprintf("Expected Content-Type to be 'application/json' but got '%s'", r.Header.Get("Content-Type")), http.StatusInternalServerError)
			return
		}

		var message SendMessageChannelRequest
		if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
			http.Error(w, fmt.Sprintf("Failed to decode request body: %v", err), http.StatusInternalServerError)
			return
		}

		expectedChatID := "@test-channel"
		if message.ChatID != expectedChatID {
			http.Error(w, fmt.Sprintf("Expected chat_id to be '%s' but got '%s'", expectedChatID, message.ChatID), http.StatusInternalServerError)
			return
		}

		expectedText := "Queue test-queue is available! Actual ticket: K80. Tickets left: 10."
		if message.Text != expectedText {
			http.Error(w, fmt.Sprintf("Expected text to be \n'%s' but got \n'%s'", expectedText, message.Text), http.StatusInternalServerError)
			return
		}

		fmt.Fprintln(w, `{"status": 200}`)
	}))

	defer mockPushOverApi.Close()

	cfg := &TelegramConfig{
		BaseApiUrl:           mockPushOverApi.URL,
		BotToken:             "123456",
		BroadcastChannelName: "test-channel",
	}

	logger := logger.NewLogger(&logger.Config{
		Level: "error"})

	sut := NewTelegramNotifier(cfg, logger, &http.Client{})

	// Act
	err := sut.SendGeneralQueueStatusUpdatePush("test-queue", true, "K80", 10)

	// Assert
	if err != nil {
		t.Fatalf("Expected successful notification sending, but got error: \"%v\"", err)
	}
}

func TestSendGeneralQueueStatusUpdatePush_WhenUnavailableMessage_SendsNotificationToTelegramApiWithCorrectFormatAndTemplate(t *testing.T) {
	// Arrange
	mockPushOverApi := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, fmt.Sprintf("Expected HTTP POST but got %v", r.Method), http.StatusInternalServerError)
			return
		}

		if r.URL.Path != "/bot123456/sendMessage" {
			http.Error(w, fmt.Sprintf("Expected URL to be '/bot123456/sendMessage' but got '%s'", r.URL.Path), http.StatusInternalServerError)
			return
		}

		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, fmt.Sprintf("Expected Content-Type to be 'application/json' but got '%s'", r.Header.Get("Content-Type")), http.StatusInternalServerError)
			return
		}

		var message SendMessageChannelRequest
		if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
			http.Error(w, fmt.Sprintf("Failed to decode request body: %v", err), http.StatusInternalServerError)
			return
		}

		expectedChatID := "@test-channel"
		if message.ChatID != expectedChatID {
			http.Error(w, fmt.Sprintf("Expected chat_id to be '%s' but got '%s'", expectedChatID, message.ChatID), http.StatusInternalServerError)
			return
		}

		expectedText := "Queue test-queue is unavailable."
		if message.Text != expectedText {
			http.Error(w, fmt.Sprintf("Expected text to be \n'%s' but got \n'%s'", expectedText, message.Text), http.StatusInternalServerError)
			return
		}

		fmt.Fprintln(w, `{"status": 200}`)
	}))

	defer mockPushOverApi.Close()

	cfg := &TelegramConfig{
		BaseApiUrl:           mockPushOverApi.URL,
		BotToken:             "123456",
		BroadcastChannelName: "test-channel",
	}

	logger := logger.NewLogger(&logger.Config{
		Level: "error"})

	sut := NewTelegramNotifier(cfg, logger, &http.Client{})

	// Act
	err := sut.SendGeneralQueueStatusUpdatePush("test-queue", false, "K80", 10)

	// Assert
	if err != nil {
		t.Fatalf("Expected successful notification sending, but got error: \"%v\"", err)
	}
}
