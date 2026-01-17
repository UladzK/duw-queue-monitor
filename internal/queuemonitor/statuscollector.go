package queuemonitor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"uladzk/duw_kolejka_checker/internal/logger"

	"github.com/avast/retry-go/v4"
)

// StatusCollector is responsible for collecting the status of a specific queue from the DUW API
type StatusCollector struct {
	cfg        *QueueMonitorConfig
	httpClient *http.Client
	log        *logger.Logger
}

// Response represents the top-level structure of the response from the DUW API
// It contains a map of city names to their respective queue states
type Response struct {
	Result map[string][]Queue `json:"result"`
}

// Queue represents a queue state retrieved from the DUW API
type Queue struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Enabled     bool   `json:"enabled"`
	Active      bool   `json:"active"`
	TicketValue string `json:"ticket_value"`
	TicketsLeft int    `json:"tickets_left"`
}

func NewStatusCollector(cfg *QueueMonitorConfig, httpClient *http.Client, log *logger.Logger) *StatusCollector {
	return &StatusCollector{
		cfg:        cfg,
		httpClient: httpClient,
		log:        log,
	}
}

func (s *StatusCollector) GetQueueStatus() (queueStatus *Queue, err error) {
	req, err := http.NewRequest("GET", s.cfg.StatusApiUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("User-Agent", "") // needed because otherwise DUW's API does not return data

	response, err := s.getStatusWithRetries(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue status after retries: %w", err)
	}

	for _, queue := range response.Result[s.cfg.StatusMonitoredQueueCity] {
		if queue.ID == s.cfg.StatusMonitoredQueueId {
			return &queue, nil
		}
	}

	s.log.Debug("Queue not found in API response",
		"queueId", s.cfg.StatusMonitoredQueueId,
		"city", s.cfg.StatusMonitoredQueueCity,
		"response", response)

	return nil, fmt.Errorf("failed to find the queue status for the queue with id: %v", s.cfg.StatusMonitoredQueueId)
}

func (s *StatusCollector) getStatusWithRetries(req *http.Request) (*Response, error) {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Duration(s.cfg.StatusCheckTimeoutMs)*time.Millisecond)
	defer cancel()

	return retry.DoWithData(
		func() (*Response, error) {
			resp, err := s.httpClient.Do(req)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
			}

			var response Response
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				return nil, fmt.Errorf("failed to parse response body: %w", err)
			}

			return &response, nil
		},
		retry.Attempts(s.cfg.StatusCheckMaxAttempts),
		retry.Delay(time.Duration(s.cfg.StatusCheckAttemptDelayMs)*time.Millisecond),
		retry.DelayType(retry.FixedDelay),
		retry.Context(timeoutCtx),
	)
}
