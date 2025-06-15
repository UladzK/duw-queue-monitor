package notifications

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"uladzk/duw_kolejka_checker/internal/logger"
)

func TestSendGeneralQueueStatusUpdateNotification_WhenRequestSuccessful_SendsNotificationToTelegramApiWithCorrectFormatAndTemplate(t *testing.T) {
	// Arrange
	testBotToken := "123456789:ABCdefGHIjklMNOpqrSTUvwxYZ"
	testChannel := "test-channel"
	testConditions := []struct {
		name                string
		queueEnabled        bool
		queueName           string
		actualTicket        string
		numberOfTicketsLeft int
		expectedMessage     string
	}{
		{"Test with available queue", true, "test-queue", "K80", 10, "Queue test-queue is available! Actual ticket: K80. Tickets left: 10."},
		{"Test with unavailable queue", false, "test-queue", "K80", 10, "Queue test-queue is unavailable."},
	}

	for _, tc := range testConditions {
		t.Run(tc.name, func(t *testing.T) {

			mockPushOverApi := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					http.Error(w, fmt.Sprintf("Expected HTTP POST but got %v", r.Method), http.StatusInternalServerError)
					return
				}

				if r.Header.Get("Content-Type") != "application/json" {
					http.Error(w, fmt.Sprintf("Expected Content-Type to be 'application/json' but got '%s'", r.Header.Get("Content-Type")), http.StatusInternalServerError)
					return
				}

				if r.URL.Path != fmt.Sprintf("/bot%v/sendMessage", testBotToken) {
					http.Error(w, fmt.Sprintf("Expected URL to be '/bot%v/sendMessage' but got '%s'", testBotToken, r.URL.Path), http.StatusInternalServerError)
					return
				}

				var message SendMessageChannelRequest
				if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
					http.Error(w, fmt.Sprintf("Failed to decode request body: %v", err), http.StatusInternalServerError)
					return
				}

				expectedChatID := fmt.Sprintf("@%v", testChannel)
				if message.ChatID != expectedChatID {
					http.Error(w, fmt.Sprintf("Expected chat_id to be '%s' but got '%s'", expectedChatID, message.ChatID), http.StatusInternalServerError)
					return
				}

				if message.Text != tc.expectedMessage {
					http.Error(w, fmt.Sprintf("Expected text to be \n'%s' but got \n'%s'", tc.expectedMessage, message.Text), http.StatusInternalServerError)
					return
				}

				fmt.Fprintln(w, `{"status": 200}`)
			}))

			defer mockPushOverApi.Close()

			cfg := &TelegramConfig{
				BaseApiUrl:           mockPushOverApi.URL,
				BotToken:             testBotToken,
				BroadcastChannelName: testChannel,
			}

			logger := logger.NewLogger(&logger.Config{
				Level: "error"})

			sut := NewTelegramNotifier(cfg, logger, &http.Client{})

			// Act
			err := sut.SendGeneralQueueStatusUpdateNotification(tc.queueName, true, tc.queueEnabled, tc.actualTicket, tc.numberOfTicketsLeft)

			// Assert
			if err != nil {
				t.Fatalf("Expected successful notification sending, but got error: \"%v\"", err)
			}
		})
	}
}
