package statuscollector

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type MonitorState struct {
	QueueActive         bool   `json:"queue_active"`          // indicates if the queue is currently active
	QueueEnabled        bool   `json:"queue_enabled"`         // indicates if the queue is enabled
	LastTicketProcessed string `json:"last_ticket_processed"` // last ticket processed in the queue
	TicketsLeft         int    `json:"tickets_left"`          // number of tickets left in the queue
}

// MonitorStateRepository is responsible for storing and retrieving the queue monitor state in Redis.
// State is persisted to ensure that no doublicate notifications are sent in case of monitor restart or crash.
// TTL is used to avoid stale data.
type MonitorStateRepository struct {
	redisClient *redis.Client
	stateTtl    time.Duration
}

const (
	queueStateRedisKey = "monitor:state"
)

func NewMonitorStateRepository(redisClient *redis.Client, stateTtlSeconds int) *MonitorStateRepository {
	return &MonitorStateRepository{
		redisClient: redisClient,
		stateTtl:    time.Duration(stateTtlSeconds) * time.Second,
	}
}

func (r *MonitorStateRepository) Get(ctx context.Context) (*MonitorState, error) {
	stateData, err := r.redisClient.Get(ctx, queueStateRedisKey).Result()

	switch {
	case err == redis.Nil:
		return nil, nil // That's fine, the state is not initialized yet or expired
	case err != nil:
		return nil, fmt.Errorf("failed to get monitor state from Redis: \"%w\"", err)
	}

	var state MonitorState
	if err := json.Unmarshal([]byte(stateData), &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal monitor state from Redis: \"%w\"", err)
	}

	return &state, nil
}

func (r *MonitorStateRepository) Save(ctx context.Context, state *MonitorState) error {
	stateData, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal monitor state: \"%w\"", err)
	}

	if err := r.redisClient.Set(ctx, queueStateRedisKey, stateData, r.stateTtl).Err(); err != nil {
		return fmt.Errorf("failed to save monitor state to Redis: \"%w\"", err)
	}

	return nil
}
