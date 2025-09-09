package queuemonitor

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// StatusCollector is responsible for collecting the status of a specific queue from the DUW API
// Note: only Wrocław city is supported for now, and only the queue with ID 24 (Odbiór Kart) is supported
type StatusCollector struct {
	cfg        *QueueMonitorConfig
	httpClient *http.Client
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

// TODO: product improvement: support multiple queues and cities by passing them in the config
const (
	odbiorKartyQueueId = 24        // ID of the queue we are interested in
	wroclawCityName    = "Wrocław" // City name for the queue we are interested in
)

func NewStatusCollector(cfg *QueueMonitorConfig, httpClient *http.Client) *StatusCollector {
	return &StatusCollector{
		cfg:        cfg,
		httpClient: httpClient,
	}
}

func (s *StatusCollector) GetQueueStatus() (queueStatus *Queue, err error) {
	req, err := http.NewRequest("GET", s.cfg.StatusApiUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("User-Agent", "") // needed because otherwise DUW's API does not return data

	//TODO: add retries with exponential backoff. including non-OK status codes
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errRespBody, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			errRespBody = []byte(fmt.Sprintf("failed to read response body: %v", readErr))
		}

		return nil, fmt.Errorf("failed to get queue status, status code: %d. response: %v", resp.StatusCode, string(errRespBody))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: \"%w\"", err)
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response body: \"%w\". body text: %v", err, string(body))
	}

	for _, queue := range response.Result[wroclawCityName] {
		if queue.ID == odbiorKartyQueueId {
			return &queue, nil
		}
	}

	return nil, fmt.Errorf("failed to find the queue status for the queue with id: %v", odbiorKartyQueueId)
}
