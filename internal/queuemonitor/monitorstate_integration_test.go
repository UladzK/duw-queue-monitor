package queuemonitor

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func initDevContainer(ctx context.Context, t *testing.T) testcontainers.Container {
	req := testcontainers.ContainerRequest{
		Image:        "redis:latest",
		Name:         "monitor-state-redis-integration-test",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}

	redisC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start Redis container: \"%v\". Test cannot be executed", err)
	}

	return redisC
}

// TODO: add env vars to split up running unit and integration tests
func TestGetAndSave_WhenRedisIsAvailable_GetsAndSavesState(t *testing.T) {
	// Arrange
	ctx := context.Background()
	redisC := initDevContainer(ctx, t)
	defer testcontainers.CleanupContainer(t, redisC)

	endpoint, err := redisC.Endpoint(ctx, "")
	if err != nil {
		t.Fatalf("Failed to get Redis endpoint: \"%v\". Test cannot be executed", err)
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr: endpoint,
	})

	testState := &MonitorState{
		QueueActive:         true,
		QueueEnabled:        true,
		LastTicketProcessed: "K123",
		TicketsLeft:         5,
	}

	sut := NewMonitorStateRepository(redisClient, 120)

	// Act
	saveErr := sut.Save(ctx, testState)
	returnedState, getErr := sut.Get(ctx)

	// Assert
	if saveErr != nil {
		t.Errorf("Expected to save state successfully, but: \"%v\"", saveErr)
	}

	if getErr != nil {
		t.Errorf("Expected to get state successfully, but: \"%v\"", getErr)
	}

	if returnedState == nil {
		t.Error("Expected to get a non-nil state, but: got nil")
	}

	if diff := cmp.Diff(testState, returnedState); diff != "" {
		t.Errorf("Get state mismatch (-want +got):\n%s", diff)
	}
}

func TestGet_WhenNoStateFoundInRedis_ReturnsNilWithoutError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	redisC := initDevContainer(ctx, t)
	defer testcontainers.CleanupContainer(t, redisC)

	endpoint, err := redisC.Endpoint(ctx, "")
	if err != nil {
		panic(fmt.Errorf("Failed to get Redis endpoint: \"%v\". Test cannot be executed", err))
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr: endpoint,
	})

	sut := NewMonitorStateRepository(redisClient, 120)

	// Act
	returnedState, getErr := sut.Get(ctx)

	// Assert
	if getErr != nil {
		t.Errorf("Expected to get state successfully, but: \"%v\"", getErr)
	}

	if returnedState != nil {
		t.Errorf("Expected to get a nil state, but: got %v", returnedState)
	}
}
