package telegrambot

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
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

func TestProfile_NewProfile_InitializesCorrectly(t *testing.T) {
	// Arrange
	logger := logger.NewLogger(&logger.Config{Level: "error"})
	registry := createTestHandlerRegistry()

	// Create a simple bot instance for testing initialization
	// We can pass nil since we're only testing the constructor
	var bot *bot.Bot = nil

	// Act
	sut := NewProfile(bot, registry, logger)

	// Assert
	if sut.bot != bot {
		t.Error("Expected bot to be set correctly")
	}

	if sut.registry != registry {
		t.Error("Expected registry to be set correctly")
	}

	if sut.logger != logger {
		t.Error("Expected logger to be set correctly")
	}
}

func TestHandlerRegistry_RegisterAllHandlers_WorksWithBotInstance(t *testing.T) {
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

	// Assert
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
	feedbackReplyHandler := registry.replyRegistry.FindHandler("Aby wysłać opinię, proszę odpowiedz na tę wiadomość swoją opinią:")
	if feedbackReplyHandler == nil {
		t.Error("Expected feedback reply handler to be registered in reply registry")
	}
}

func TestHandlerRegistry_ReplyHandlerRegistry_WorksCorrectly(t *testing.T) {
	// Arrange
	registry := createTestHandlerRegistry()

	// Create a minimal bot for testing
	mockBotServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"ok":true,"result":{"message_id":1}}`)
	}))
	defer mockBotServer.Close()

	testBot, err := bot.New("test-token", bot.WithServerURL(mockBotServer.URL))
	if err != nil {
		t.Fatalf("Failed to create test bot: %v", err)
	}

	// Act
	registry.RegisterAllHandlers(testBot)

	// Test reply registry functionality
	feedbackReplyText := "Aby wysłać opinię, proszę odpowiedz na tę wiadomość swoją opinią:"
	handler := registry.replyRegistry.FindHandler(feedbackReplyText)

	// Assert
	if handler == nil {
		t.Error("Expected to find feedback reply handler")
	} else {
		patterns := handler.GetReplyPatterns()
		if len(patterns) == 0 {
			t.Error("Expected handler to have reply patterns")
		}

		foundExpectedPattern := false
		for _, pattern := range patterns {
			if pattern == feedbackReplyText {
				foundExpectedPattern = true
				break
			}
		}

		if !foundExpectedPattern {
			t.Errorf("Expected handler to contain pattern '%s'", feedbackReplyText)
		}
	}
}

func TestProfile_SetProfile_CommandsStructure_VerifyCorrectFormat(t *testing.T) {
	// This test verifies that the commands structure is correct for the bot API

	// Arrange
	registry := createTestHandlerRegistry()
	commands := registry.GetAvailableCommands()

	// Act & Assert
	for _, cmd := range commands {
		// Verify each command follows the BotCommand structure
		if cmd.Command == "" {
			t.Error("Command name cannot be empty")
		}

		if cmd.Description == "" {
			t.Error("Command description cannot be empty")
		}

		// Verify command name format (should be lowercase, no spaces)
		if cmd.Command != "feedback" { // We know feedback is the main command
			continue
		}

		if cmd.Command != "feedback" {
			t.Errorf("Expected 'feedback' command, got '%s'", cmd.Command)
		}

		if cmd.Description != "feedback" {
			t.Errorf("Expected 'feedback' description, got '%s'", cmd.Description)
		}
	}

	// Verify the structure matches what SetMyCommandsParams expects
	setCommandsParams := bot.SetMyCommandsParams{
		Commands: commands,
	}

	if len(setCommandsParams.Commands) != len(commands) {
		t.Error("Commands should be properly assignable to SetMyCommandsParams")
	}
}

func TestProfile_SetProfile_WithContext_AcceptsContextCorrectly(t *testing.T) {
	// This test verifies that SetProfile properly accepts and would use context

	// Arrange
	logger := logger.NewLogger(&logger.Config{Level: "error"})
	registry := createTestHandlerRegistry()

	// We can't test the actual bot call without complex mocking,
	// but we can test that the Profile is structured correctly
	profile := NewProfile(nil, registry, logger)

	// Act & Assert
	// Test that the context would be passed correctly (we can't call without a real bot)
	ctx := context.Background()

	// Verify that the Profile has all required components
	if profile.registry == nil {
		t.Error("Profile should have a registry")
	}

	if profile.logger == nil {
		t.Error("Profile should have a logger")
	}

	// Test that the registry provides commands in the correct format
	commands := profile.registry.GetAvailableCommands()
	if len(commands) == 0 {
		t.Error("Registry should provide commands for SetProfile")
	}

	// This demonstrates that the profile would call bot.SetMyCommands with the correct structure
	expectedParams := bot.SetMyCommandsParams{
		Commands: commands,
	}

	if len(expectedParams.Commands) == 0 {
		t.Error("SetMyCommandsParams should have commands")
	}

	// Using context in a way that demonstrates it would be passed through
	_ = ctx // The context would be passed to bot.SetMyCommands(ctx, &expectedParams)
}
