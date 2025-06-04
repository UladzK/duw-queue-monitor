package statuscollector

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type StatusCollectorService struct {
	cfg        *StatusCollectorConfig
	httpClient *http.Client
}

// QueueOperation represents an operation within a queue
type QueueOperation struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

// Queue represents a queue from the API
type Queue struct {
	ID                 int              `json:"id"`
	Name               string           `json:"name"`
	Operations         []QueueOperation `json:"operations"`
	TicketCount        int              `json:"ticket_count"`
	TicketsServed      int              `json:"tickets_served"`
	Workplaces         int              `json:"workplaces"`
	AverageWaitTime    int              `json:"average_wait_time"`
	AverageServiceTime int              `json:"average_service_time"`
	RegisteredTickets  int              `json:"registered_tickets"`
	MaxTickets         int              `json:"max_tickets"`
	TicketValue        string           `json:"ticket_value"`
	Active             bool             `json:"active"`
	Location           string           `json:"location"`
	TicketsLeft        int              `json:"tickets_left"`
	Enabled            bool             `json:"enabled"`
}

// Response represents the overall API response structure
type Response struct {
	Result map[string][]Queue `json:"result"`
}

const (
	odbiorKartyQueueId = 24 // ID of the queue we are interested in
)

func NewStatusCollectorService(cfg *StatusCollectorConfig) *StatusCollectorService {
	return &StatusCollectorService{
		cfg: cfg,
		httpClient: &http.Client{
			Transport: &http.Transport{
				// needed because otherwise the TLS connection is not established when calling from inside the container. silly workaround which just works
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

func (s *StatusCollectorService) getQueueStatus() (queueStatus *Queue, err error) {

	fmt.Println(s.cfg.StatusApiUrl)
	req, err := http.NewRequest("GET", s.cfg.StatusApiUrl, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		time.Sleep(10 * time.Second)
		return nil, err
	}

	// Add headers to make the request more browser-like
	// needed because otherwise urząd's API does not return data :( if they think that it should protect from bots, then they are not very smart ofc
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		fmt.Printf("Error making HTTP request: %v\n", err)
		time.Sleep(10 * time.Second)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("got %v status code\n", resp.StatusCode)
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		time.Sleep(10 * time.Second)
		return nil, err
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		time.Sleep(10 * time.Second)
		return nil, err
	}
	for _, queue := range response.Result["Wrocław"] {
		if queue.ID == odbiorKartyQueueId {
			return &queue, nil
		}
	}

	return nil, fmt.Errorf("failed to find the queue status for the queue with id: %v", odbiorKartyQueueId)
}
