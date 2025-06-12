package statuscollector

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type MonitorState struct {
	QueueActive         bool   `json:"queue_active"`          // indicates if the queue is currently active
	QueueEnabled        bool   `json:"queue_enabled"`         // indicates if the queue is enabled
	LastTicketProcessed string `json:"last_ticket_processed"` // last ticket processed in the queue
	TicketsLeft         int    `json:"tickets_left"`          // number of tickets left in the queue
}

type MonitorStateRepository struct {
	redisClient *redis.Client
}

const (
	QueueStateKey = "queue-monitor-state:latest"
)

func NewMonitorStateRepository(redisClient *redis.Client) *MonitorStateRepository {
	return &MonitorStateRepository{
		redisClient: redisClient,
	}
}

func (r *MonitorStateRepository) GetState(ctx context.Context) (*MonitorState, error) {

	stateData, err := r.redisClient.Get(ctx, QueueStateKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get queue state from Redis: \"%w\"", err)
	}

	if stateData == "" {
		return nil, nil // That's fine, the state is not initialized yet or expired
	}

	var state MonitorState
	if err := json.Unmarshal([]byte(stateData), &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal queue state from Redis: \"%w\"", err)
	}

	return &state, nil
}
