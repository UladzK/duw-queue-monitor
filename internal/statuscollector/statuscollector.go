package statuscollector

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// StatusCollector is responsible for collecting the status of a specific queue from the DUW API
// Note: only Wrocław city is supported for now, and only the queue with ID 24 (Odbiór Kart) is supported
type StatusCollector struct {
	cfg        *StatusCollectorConfig
	httpClient *http.Client
}

// Response represents the top-level structure of the response from the DUW API
// It contains a map of city names to their respective queue states
type Response struct {
	Result map[string][]Queue `json:"result"`
}

// Queue represents a queue state retrieved from the DUW API
type Queue struct {
	ID                 int    `json:"id"`
	Name               string `json:"name"`
	TicketCount        int    `json:"ticket_count"`
	TicketsServed      int    `json:"tickets_served"`
	Workplaces         int    `json:"workplaces"`
	AverageWaitTime    int    `json:"average_wait_time"`
	AverageServiceTime int    `json:"average_service_time"`
	RegisteredTickets  int    `json:"registered_tickets"`
	MaxTickets         int    `json:"max_tickets"`
	TicketValue        string `json:"ticket_value"`
	Active             bool   `json:"active"`
	Location           string `json:"location"`
	TicketsLeft        int    `json:"tickets_left"`
	Enabled            bool   `json:"enabled"`
}

const (
	odbiorKartyQueueId = 24        // ID of the queue we are interested in
	wroclawCityName    = "Wrocław" // City name for the queue we are interested in
)

func NewStatusCollector(cfg *StatusCollectorConfig) *StatusCollector {
	return &StatusCollector{
		cfg: cfg,
		httpClient: &http.Client{
			Transport: &http.Transport{
				// needed because otherwise the TLS connection is not established when calling from inside the container. silly workaround which just works
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

func (s *StatusCollector) GetQueueStatus() (queueStatus *Queue, err error) {
	req, err := http.NewRequest("GET", s.cfg.StatusApiUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// needed because otherwise DUW's API does not return data
	req.Header.Set("User-Agent", "")

	//TODO: add retries with exponential backoff. including non-OK status codes
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get queue status, status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response body: %w", err)
	}

	for _, queue := range response.Result[wroclawCityName] {
		if queue.ID == odbiorKartyQueueId {
			return &queue, nil
		}
	}

	return nil, fmt.Errorf("failed to find the queue status for the queue with id: %v", odbiorKartyQueueId)
}
