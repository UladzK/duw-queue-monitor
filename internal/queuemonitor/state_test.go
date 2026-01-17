package queuemonitor

import (
	"context"
	"fmt"
	"testing"
	"uladzk/duw_kolejka_checker/internal/logger"
)

type testNotifier struct {
	sendMessageCalled bool
	shouldFail        bool
}

func (n *testNotifier) SendMessage(ctx context.Context, chatID, text string) error {
	n.sendMessageCalled = true
	if n.shouldFail {
		return fmt.Errorf("notification failed")
	}
	return nil
}

func createTestMonitor(notifier Notifier) *DefaultQueueMonitor {
	cfg := &Config{
		BroadcastChannelName: "test-channel",
	}
	log := logger.NewLogger(&logger.Config{Level: "error"})
	return &DefaultQueueMonitor{
		cfg:      cfg,
		log:      log,
		notifier: notifier,
		state:    &UninitializedState{},
	}
}

// =============================================================================
// UninitializedState Tests
// =============================================================================

func TestUninitializedState_Handle_WhenQueueInactive_TransitionsToInactive(t *testing.T) {
	// Arrange
	state := &UninitializedState{}
	notifier := &testNotifier{}
	monitor := createTestMonitor(notifier)
	queue := &Queue{Active: false, Enabled: false}

	// Act
	newState, err := state.Handle(context.Background(), monitor, queue)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newState.Name() != "Inactive" {
		t.Errorf("expected Inactive state, got %s", newState.Name())
	}
	if notifier.sendMessageCalled {
		t.Error("expected no notification for inactive queue")
	}
}

func TestUninitializedState_Handle_WhenQueueActiveEnabled_TransitionsToActiveEnabled(t *testing.T) {
	// Arrange
	state := &UninitializedState{}
	notifier := &testNotifier{}
	monitor := createTestMonitor(notifier)
	queue := &Queue{Active: true, Enabled: true, TicketsLeft: 10}

	// Act
	newState, err := state.Handle(context.Background(), monitor, queue)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newState.Name() != "ActiveEnabled" {
		t.Errorf("expected ActiveEnabled state, got %s", newState.Name())
	}
	if !notifier.sendMessageCalled {
		t.Error("expected notification for first active queue")
	}
	if newState.TicketsLeft() != 10 {
		t.Errorf("expected ticketsLeft=10, got %d", newState.TicketsLeft())
	}
}

func TestUninitializedState_Handle_WhenQueueActiveDisabled_TransitionsToActiveDisabled(t *testing.T) {
	// Arrange
	state := &UninitializedState{}
	notifier := &testNotifier{}
	monitor := createTestMonitor(notifier)
	queue := &Queue{Active: true, Enabled: false}

	// Act
	newState, err := state.Handle(context.Background(), monitor, queue)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newState.Name() != "ActiveDisabled" {
		t.Errorf("expected ActiveDisabled state, got %s", newState.Name())
	}
	if !notifier.sendMessageCalled {
		t.Error("expected notification for first active queue")
	}
}

// =============================================================================
// InactiveState Tests
// =============================================================================

func TestInactiveState_Handle_WhenQueueStaysInactive_StaysInInactive(t *testing.T) {
	// Arrange
	state := &InactiveState{}
	notifier := &testNotifier{}
	monitor := createTestMonitor(notifier)
	queue := &Queue{Active: false}

	// Act
	newState, err := state.Handle(context.Background(), monitor, queue)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newState.Name() != "Inactive" {
		t.Errorf("expected Inactive state, got %s", newState.Name())
	}
	if notifier.sendMessageCalled {
		t.Error("expected no notification when staying inactive")
	}
}

func TestInactiveState_Handle_WhenQueueBecomesActiveEnabled_TransitionsToActiveEnabled(t *testing.T) {
	// Arrange
	state := &InactiveState{}
	notifier := &testNotifier{}
	monitor := createTestMonitor(notifier)
	queue := &Queue{Active: true, Enabled: true, TicketsLeft: 5}

	// Act
	newState, err := state.Handle(context.Background(), monitor, queue)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newState.Name() != "ActiveEnabled" {
		t.Errorf("expected ActiveEnabled state, got %s", newState.Name())
	}
	if !notifier.sendMessageCalled {
		t.Error("expected notification when queue becomes active")
	}
}

func TestInactiveState_Handle_WhenQueueBecomesActiveDisabled_TransitionsToActiveDisabled(t *testing.T) {
	// Arrange
	state := &InactiveState{}
	notifier := &testNotifier{}
	monitor := createTestMonitor(notifier)
	queue := &Queue{Active: true, Enabled: false}

	// Act
	newState, err := state.Handle(context.Background(), monitor, queue)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newState.Name() != "ActiveDisabled" {
		t.Errorf("expected ActiveDisabled state, got %s", newState.Name())
	}
	if !notifier.sendMessageCalled {
		t.Error("expected notification when queue becomes active")
	}
}

// =============================================================================
// ActiveDisabledState Tests
// =============================================================================

func TestActiveDisabledState_Handle_WhenQueueBecomesInactive_TransitionsToInactive(t *testing.T) {
	// Arrange
	state := &ActiveDisabledState{}
	notifier := &testNotifier{}
	monitor := createTestMonitor(notifier)
	queue := &Queue{Active: false}

	// Act
	newState, err := state.Handle(context.Background(), monitor, queue)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newState.Name() != "Inactive" {
		t.Errorf("expected Inactive state, got %s", newState.Name())
	}
	if notifier.sendMessageCalled {
		t.Error("expected no notification when going to inactive")
	}
}

func TestActiveDisabledState_Handle_WhenQueueBecomesEnabled_TransitionsToActiveEnabled(t *testing.T) {
	// Arrange
	state := &ActiveDisabledState{}
	notifier := &testNotifier{}
	monitor := createTestMonitor(notifier)
	queue := &Queue{Active: true, Enabled: true, TicketsLeft: 15}

	// Act
	newState, err := state.Handle(context.Background(), monitor, queue)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newState.Name() != "ActiveEnabled" {
		t.Errorf("expected ActiveEnabled state, got %s", newState.Name())
	}
	if !notifier.sendMessageCalled {
		t.Error("expected notification when queue becomes enabled")
	}
	if newState.TicketsLeft() != 15 {
		t.Errorf("expected ticketsLeft=15, got %d", newState.TicketsLeft())
	}
}

func TestActiveDisabledState_Handle_WhenQueueStaysDisabled_StaysInActiveDisabled(t *testing.T) {
	// Arrange
	state := &ActiveDisabledState{}
	notifier := &testNotifier{}
	monitor := createTestMonitor(notifier)
	queue := &Queue{Active: true, Enabled: false}

	// Act
	newState, err := state.Handle(context.Background(), monitor, queue)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newState.Name() != "ActiveDisabled" {
		t.Errorf("expected ActiveDisabled state, got %s", newState.Name())
	}
	if notifier.sendMessageCalled {
		t.Error("expected no notification when staying disabled")
	}
}

// =============================================================================
// ActiveEnabledState Tests
// =============================================================================

func TestActiveEnabledState_Handle_WhenQueueBecomesInactive_TransitionsToInactive(t *testing.T) {
	// Arrange
	state := &ActiveEnabledState{ticketsLeft: 10}
	notifier := &testNotifier{}
	monitor := createTestMonitor(notifier)
	queue := &Queue{Active: false}

	// Act
	newState, err := state.Handle(context.Background(), monitor, queue)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newState.Name() != "Inactive" {
		t.Errorf("expected Inactive state, got %s", newState.Name())
	}
	if notifier.sendMessageCalled {
		t.Error("expected no notification when going to inactive")
	}
}

func TestActiveEnabledState_Handle_WhenQueueBecomesDisabled_TransitionsToActiveDisabled(t *testing.T) {
	// Arrange
	state := &ActiveEnabledState{ticketsLeft: 10}
	notifier := &testNotifier{}
	monitor := createTestMonitor(notifier)
	queue := &Queue{Active: true, Enabled: false}

	// Act
	newState, err := state.Handle(context.Background(), monitor, queue)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newState.Name() != "ActiveDisabled" {
		t.Errorf("expected ActiveDisabled state, got %s", newState.Name())
	}
	if !notifier.sendMessageCalled {
		t.Error("expected notification when queue becomes disabled")
	}
}

func TestActiveEnabledState_Handle_WhenTicketsChange_NotifiesAndStaysEnabled(t *testing.T) {
	// Arrange
	state := &ActiveEnabledState{ticketsLeft: 10}
	notifier := &testNotifier{}
	monitor := createTestMonitor(notifier)
	queue := &Queue{Active: true, Enabled: true, TicketsLeft: 5}

	// Act
	newState, err := state.Handle(context.Background(), monitor, queue)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newState.Name() != "ActiveEnabled" {
		t.Errorf("expected ActiveEnabled state, got %s", newState.Name())
	}
	if !notifier.sendMessageCalled {
		t.Error("expected notification when tickets change")
	}
	if newState.TicketsLeft() != 5 {
		t.Errorf("expected ticketsLeft=5, got %d", newState.TicketsLeft())
	}
}

func TestActiveEnabledState_Handle_WhenNoChange_StaysEnabledWithoutNotification(t *testing.T) {
	// Arrange
	state := &ActiveEnabledState{ticketsLeft: 10}
	notifier := &testNotifier{}
	monitor := createTestMonitor(notifier)
	queue := &Queue{Active: true, Enabled: true, TicketsLeft: 10}

	// Act
	newState, err := state.Handle(context.Background(), monitor, queue)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newState.Name() != "ActiveEnabled" {
		t.Errorf("expected ActiveEnabled state, got %s", newState.Name())
	}
	if notifier.sendMessageCalled {
		t.Error("expected no notification when nothing changes")
	}
}

// =============================================================================
// StateFromPersistence Tests
// =============================================================================

func TestStateFromPersistence_WithNil_ReturnsUninitializedState(t *testing.T) {
	// Arrange / Act
	state := StateFromPersistence(nil)

	// Assert
	if state.Name() != "Uninitialized" {
		t.Errorf("expected Uninitialized state, got %s", state.Name())
	}
}

func TestStateFromPersistence_WithNewFormat_UsesStateName(t *testing.T) {
	testCases := []struct {
		name          string
		input         *MonitorState
		expectedState string
		expectedTL    int
	}{
		{
			"Inactive state",
			&MonitorState{StateName: "Inactive"},
			"Inactive",
			0,
		},
		{
			"ActiveDisabled state",
			&MonitorState{StateName: "ActiveDisabled"},
			"ActiveDisabled",
			0,
		},
		{
			"ActiveEnabled state",
			&MonitorState{StateName: "ActiveEnabled", TicketsLeft: 10},
			"ActiveEnabled",
			10,
		},
		{
			"Uninitialized state",
			&MonitorState{StateName: "Uninitialized"},
			"Uninitialized",
			0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			state := StateFromPersistence(tc.input)

			// Assert
			if state.Name() != tc.expectedState {
				t.Errorf("expected %s state, got %s", tc.expectedState, state.Name())
			}
			if state.TicketsLeft() != tc.expectedTL {
				t.Errorf("expected ticketsLeft=%d, got %d", tc.expectedTL, state.TicketsLeft())
			}
		})
	}
}

func TestStateFromPersistence_WithLegacyFormat_DerivesStateFromBooleans(t *testing.T) {
	testCases := []struct {
		name          string
		input         *MonitorState
		expectedState string
		expectedTL    int
	}{
		{
			"Inactive (QueueActive=false)",
			&MonitorState{QueueActive: false, QueueEnabled: false},
			"Inactive",
			0,
		},
		{
			"ActiveDisabled (QueueActive=true, QueueEnabled=false)",
			&MonitorState{QueueActive: true, QueueEnabled: false},
			"ActiveDisabled",
			0,
		},
		{
			"ActiveEnabled (QueueActive=true, QueueEnabled=true)",
			&MonitorState{QueueActive: true, QueueEnabled: true, TicketsLeft: 5},
			"ActiveEnabled",
			5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			state := StateFromPersistence(tc.input)

			// Assert
			if state.Name() != tc.expectedState {
				t.Errorf("expected %s state, got %s", tc.expectedState, state.Name())
			}
			if state.TicketsLeft() != tc.expectedTL {
				t.Errorf("expected ticketsLeft=%d, got %d", tc.expectedTL, state.TicketsLeft())
			}
		})
	}
}

// =============================================================================
// StateToPersistence Tests
// =============================================================================

func TestStateToPersistence_IncludesStateName(t *testing.T) {
	testCases := []struct {
		name          string
		state         QueueState
		expectedName  string
		expectedAct   bool
		expectedEn    bool
		expectedTL    int
	}{
		{
			"UninitializedState",
			&UninitializedState{},
			"Uninitialized",
			false,
			false,
			0,
		},
		{
			"InactiveState",
			&InactiveState{},
			"Inactive",
			false,
			false,
			0,
		},
		{
			"ActiveDisabledState",
			&ActiveDisabledState{},
			"ActiveDisabled",
			true,
			false,
			0,
		},
		{
			"ActiveEnabledState",
			&ActiveEnabledState{ticketsLeft: 7},
			"ActiveEnabled",
			true,
			true,
			7,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			ms := StateToPersistence(tc.state, nil)

			// Assert
			if ms.StateName != tc.expectedName {
				t.Errorf("expected StateName=%s, got %s", tc.expectedName, ms.StateName)
			}
			if ms.QueueActive != tc.expectedAct {
				t.Errorf("expected QueueActive=%v, got %v", tc.expectedAct, ms.QueueActive)
			}
			if ms.QueueEnabled != tc.expectedEn {
				t.Errorf("expected QueueEnabled=%v, got %v", tc.expectedEn, ms.QueueEnabled)
			}
			if ms.TicketsLeft != tc.expectedTL {
				t.Errorf("expected TicketsLeft=%d, got %d", tc.expectedTL, ms.TicketsLeft)
			}
		})
	}
}

func TestStateToPersistence_WithQueueData_IncludesTicketInfo(t *testing.T) {
	// Arrange
	state := &ActiveEnabledState{ticketsLeft: 10}
	queue := &Queue{TicketValue: "K123", TicketsLeft: 8}

	// Act
	ms := StateToPersistence(state, queue)

	// Assert
	if ms.LastTicketProcessed != "K123" {
		t.Errorf("expected LastTicketProcessed=K123, got %s", ms.LastTicketProcessed)
	}
	if ms.TicketsLeft != 8 {
		t.Errorf("expected TicketsLeft=8 (from queue), got %d", ms.TicketsLeft)
	}
}

// =============================================================================
// Error Handling Tests
// =============================================================================

func TestState_Handle_WhenNotificationFails_ReturnsErrorAndKeepsState(t *testing.T) {
	testCases := []struct {
		name  string
		state QueueState
		queue *Queue
	}{
		{
			"UninitializedState with active queue",
			&UninitializedState{},
			&Queue{Active: true, Enabled: true},
		},
		{
			"InactiveState transitioning to active",
			&InactiveState{},
			&Queue{Active: true, Enabled: true},
		},
		{
			"ActiveDisabledState transitioning to enabled",
			&ActiveDisabledState{},
			&Queue{Active: true, Enabled: true},
		},
		{
			"ActiveEnabledState transitioning to disabled",
			&ActiveEnabledState{ticketsLeft: 10},
			&Queue{Active: true, Enabled: false},
		},
		{
			"ActiveEnabledState with ticket change",
			&ActiveEnabledState{ticketsLeft: 10},
			&Queue{Active: true, Enabled: true, TicketsLeft: 5},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			notifier := &testNotifier{shouldFail: true}
			monitor := createTestMonitor(notifier)

			// Act
			newState, err := tc.state.Handle(context.Background(), monitor, tc.queue)

			// Assert
			if err == nil {
				t.Error("expected error when notification fails")
			}
			if newState.Name() != tc.state.Name() {
				t.Errorf("expected state to remain %s, got %s", tc.state.Name(), newState.Name())
			}
		})
	}
}
