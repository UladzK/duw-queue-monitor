package handlers

import (
	"context"
	"testing"
	"uladzk/duw_kolejka_checker/internal/logger"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/google/go-cmp/cmp"
)

type mockHandlerRegistry struct {
	commands []models.BotCommand
}

func (m *mockHandlerRegistry) GetAvailableCommands() []models.BotCommand {
	return m.commands
}

type mockReplyRegistry struct {
	handlers map[string]ReplyHandler
}

func (m *mockReplyRegistry) RegisterReplyHandler(handler ReplyHandler) {
	for _, pattern := range handler.GetReplyPatterns() {
		m.handlers[pattern] = handler
	}
}

func (m *mockReplyRegistry) FindHandler(replyText string) ReplyHandler {
	return m.handlers[replyText]
}

func TestBuildMenuMessage_WithMultipleCommands_BuildsCorrectMenuMessage(t *testing.T) {
	// Arrange
	mockHandlerRegistry := &mockHandlerRegistry{
		commands: []models.BotCommand{
			{Command: "feedback", Description: "Send feedback"},
			{Command: "status", Description: "Check status"},
			{Command: "help", Description: "Show help"},
		},
	}

	expectedMessage := "Witaj!\n\n<b>Dostępne komendy</b>\n/feedback - Send feedback\n/status - Check status\n/help - Show help\n\nUżyj /start aby zobaczyć to menu ponownie\n"

	// Act
	result := buildMenuMessage(mockHandlerRegistry)

	// Assert
	if result != expectedMessage {
		t.Errorf("Expected menu message:\n%s\nGot:\n%s", expectedMessage, result)
	}
}

func TestBuildMenuMessage_WithNoCommands_BuildsMenuWithNoCommands(t *testing.T) {
	// Arrange
	mockHandlerRegistry := &mockHandlerRegistry{
		commands: []models.BotCommand{},
	}

	expectedMessage := "Witaj!\n\n<b>Dostępne komendy</b>\n\n\nUżyj /start aby zobaczyć to menu ponownie\n"

	// Act
	result := buildMenuMessage(mockHandlerRegistry)

	// Assert
	if result != expectedMessage {
		t.Errorf("Expected menu message:\n%s\nGot:\n%s", expectedMessage, result)
	}
}

func TestDefaultHandler_NewDefaultHandler_InitializesCorrectly(t *testing.T) {
	// Arrange
	logger := logger.NewLogger(&logger.Config{Level: "error"})
	mockHandlerRegistry := &mockHandlerRegistry{
		commands: []models.BotCommand{
			{Command: "feedback", Description: "feedback"},
		},
	}
	mockReplyRegistry := &mockReplyRegistry{handlers: make(map[string]ReplyHandler)}

	expectedMenuMessage := "Witaj!\n\n<b>Dostępne komendy</b>\n/feedback - feedback\n\nUżyj /start aby zobaczyć to menu ponownie\n"

	// Act
	sut := NewDefaultHandler(logger, mockReplyRegistry, mockHandlerRegistry)

	// Assert
	if sut.menuMessage != expectedMenuMessage {
		t.Errorf("Expected menu message to be initialized correctly:\n%s\nGot:\n%s", expectedMenuMessage, sut.menuMessage)
	}

	if sut.replyRegistry != mockReplyRegistry {
		t.Error("Expected reply registry to be set correctly")
	}

	if sut.handlerRegistry != mockHandlerRegistry {
		t.Error("Expected handler registry to be set correctly")
	}
}

func TestDefaultHandler_HandleReplyMessage_WhenUpdateIsNil_ReturnsFalse(t *testing.T) {
	// Arrange
	logger := logger.NewLogger(&logger.Config{Level: "error"})
	mockHandlerRegistry := &mockHandlerRegistry{commands: []models.BotCommand{}}
	mockReplyRegistry := &mockReplyRegistry{handlers: make(map[string]ReplyHandler)}

	sut := NewDefaultHandler(logger, mockReplyRegistry, mockHandlerRegistry)

	// Act
	result := sut.handleReplyMessage(nil, nil, nil)

	// Assert
	if result != false {
		t.Error("Expected handleReplyMessage to return false when update is nil")
	}
}

func TestDefaultHandler_HandleReplyMessage_WhenMessageIsNil_ReturnsFalse(t *testing.T) {
	// Arrange
	logger := logger.NewLogger(&logger.Config{Level: "error"})
	mockHandlerRegistry := &mockHandlerRegistry{commands: []models.BotCommand{}}
	mockReplyRegistry := &mockReplyRegistry{handlers: make(map[string]ReplyHandler)}

	sut := NewDefaultHandler(logger, mockReplyRegistry, mockHandlerRegistry)
	update := &models.Update{Message: nil}

	// Act
	result := sut.handleReplyMessage(nil, nil, update)

	// Assert
	if result != false {
		t.Error("Expected handleReplyMessage to return false when message is nil")
	}
}

func TestDefaultHandler_HandleReplyMessage_WhenReplyToMessageIsNil_ReturnsFalse(t *testing.T) {
	// Arrange
	logger := logger.NewLogger(&logger.Config{Level: "error"})
	mockHandlerRegistry := &mockHandlerRegistry{commands: []models.BotCommand{}}
	mockReplyRegistry := &mockReplyRegistry{handlers: make(map[string]ReplyHandler)}

	sut := NewDefaultHandler(logger, mockReplyRegistry, mockHandlerRegistry)
	update := &models.Update{
		Message: &models.Message{
			ReplyToMessage: nil,
		},
	}

	// Act
	result := sut.handleReplyMessage(nil, nil, update)

	// Assert
	if result != false {
		t.Error("Expected handleReplyMessage to return false when reply to message is nil")
	}
}

func TestDefaultHandler_HandleReplyMessage_WhenReplyToMessageTextIsEmpty_ReturnsFalse(t *testing.T) {
	// Arrange
	logger := logger.NewLogger(&logger.Config{Level: "error"})
	mockHandlerRegistry := &mockHandlerRegistry{commands: []models.BotCommand{}}
	mockReplyRegistry := &mockReplyRegistry{handlers: make(map[string]ReplyHandler)}

	sut := NewDefaultHandler(logger, mockReplyRegistry, mockHandlerRegistry)
	update := &models.Update{
		Message: &models.Message{
			ReplyToMessage: &models.Message{
				Text: "",
			},
		},
	}

	// Act
	result := sut.handleReplyMessage(nil, nil, update)

	// Assert
	if result != false {
		t.Error("Expected handleReplyMessage to return false when reply to message text is empty")
	}
}

func TestDefaultHandler_HandleReplyMessage_WhenNoHandlerFound_ReturnsFalse(t *testing.T) {
	// Arrange
	logger := logger.NewLogger(&logger.Config{Level: "error"})
	mockHandlerRegistry := &mockHandlerRegistry{commands: []models.BotCommand{}}
	mockReplyRegistry := &mockReplyRegistry{handlers: make(map[string]ReplyHandler)}

	sut := NewDefaultHandler(logger, mockReplyRegistry, mockHandlerRegistry)
	update := &models.Update{
		Message: &models.Message{
			ReplyToMessage: &models.Message{
				Text: "unknown pattern",
			},
		},
	}

	// Act
	result := sut.handleReplyMessage(nil, nil, update)

	// Assert
	if result != false {
		t.Error("Expected handleReplyMessage to return false when no handler found")
	}
}

type mockReplyHandler struct {
	patterns     []string
	handleCalled bool
	lastUpdate   *models.Update
}

func (m *mockReplyHandler) GetReplyPatterns() []string {
	return m.patterns
}

func (m *mockReplyHandler) HandleReply(ctx context.Context, b *bot.Bot, update *models.Update) {
	m.handleCalled = true
	m.lastUpdate = update
}

func TestDefaultHandler_HandleReplyMessage_WhenHandlerFound_CallsHandlerAndReturnsTrue(t *testing.T) {
	// Arrange
	logger := logger.NewLogger(&logger.Config{Level: "error"})
	mockHandlerRegistry := &mockHandlerRegistry{commands: []models.BotCommand{}}
	mockReplyRegistry := &mockReplyRegistry{handlers: make(map[string]ReplyHandler)}

	replyHandler := &mockReplyHandler{patterns: []string{"test reply pattern"}}
	mockReplyRegistry.RegisterReplyHandler(replyHandler)

	sut := NewDefaultHandler(logger, mockReplyRegistry, mockHandlerRegistry)
	update := &models.Update{
		Message: &models.Message{
			ReplyToMessage: &models.Message{
				Text: "test reply pattern",
			},
		},
	}

	// Act
	result := sut.handleReplyMessage(nil, nil, update)

	// Assert
	if result != true {
		t.Error("Expected handleReplyMessage to return true when handler is found and called")
	}

	if !replyHandler.handleCalled {
		t.Error("Expected reply handler to be called")
	}

	if diff := cmp.Diff(replyHandler.lastUpdate, update); diff != "" {
		t.Errorf("Reply handler called with wrong update (-want +got):\n%s", diff)
	}
}
