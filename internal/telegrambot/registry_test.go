package telegrambot

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"uladzk/duw_kolejka_checker/internal/logger"
	"uladzk/duw_kolejka_checker/internal/notifications"

	"github.com/go-telegram/bot"
)

func createTestHandlerRegistry() *HandlerRegistry {
	logger := logger.NewLogger(&logger.Config{Level: "error"})

	// Create a mock HTTP server for TelegramNotifier
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"ok":true}`)
	}))

	cfg := &notifications.TelegramConfig{
		BaseApiUrl: server.URL,
		BotToken:   "test-token",
	}

	telegramNotifier := notifications.NewTelegramNotifier(cfg, logger, &http.Client{})

	return NewHandlerRegistry(logger, telegramNotifier, "admin123")
}

func TestHandlerRegistry_RegisterAllHandlers_FullFunctionality(t *testing.T) {
	// Arrange
	registry := createTestHandlerRegistry()

	// Create a mock HTTP server to simulate bot API
	var capturedRegistrations []string
	mockBotServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedRegistrations = append(capturedRegistrations, r.URL.Path)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"ok":true,"result":{"message_id":1}}`)
	}))
	defer mockBotServer.Close()

	// Create a real bot instance pointed at our test server
	testBot, err := bot.New("test-token", bot.WithServerURL(mockBotServer.URL))
	if err != nil {
		t.Fatalf("Failed to create test bot: %v", err)
	}

	// Act
	registry.RegisterAllHandlers(testBot)

	// Assert - Test basic handler registration
	// We can't directly test the registration, but we can verify that the registry
	// contains the handlers we expect and that calling RegisterAllHandlers doesn't panic

	// Test that we can get the default handler
	defaultHandler := registry.GetDefaultHandler()
	if defaultHandler == nil {
		t.Error("Expected default handler to be available")
	}

	// Test that reply registry has been populated
	if registry.replyRegistry == nil {
		t.Error("Expected reply registry to be initialized")
	}

	// Test that feedback reply handler was registered
	feedbackReplyText := "Aby wysłać opinię, proszę odpowiedz na tę wiadomość swoją opinią:"
	feedbackReplyHandler := registry.replyRegistry.FindHandler(feedbackReplyText)
	if feedbackReplyHandler == nil {
		t.Error("Expected feedback reply handler to be registered in reply registry")
	}

	// Assert - Test reply registry functionality
	handler := registry.replyRegistry.FindHandler(feedbackReplyText)
	if handler == nil {
		t.Error("Expected to find feedback reply handler")
	} else {
		patterns := handler.GetReplyPatterns()
		if len(patterns) == 0 {
			t.Error("Expected handler to have reply patterns")
		}

		if !slices.Contains(patterns, feedbackReplyText) {
			t.Errorf("Expected handler to contain pattern '%s'", feedbackReplyText)
		}
	}
}
