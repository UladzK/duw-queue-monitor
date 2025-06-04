package statuscollector

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

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

func getQueueStatus() (queueStatus *Queue, err error) {

	const odbiorKartyQueueId = 24
	req, err := http.NewRequest("GET", "https://rezerwacje.duw.pl/status_kolejek/query.php?status=", nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		time.Sleep(10 * time.Second)
		return nil, err
	}

	// Add headers to make the request more browser-like
	// needed because otherwise urząd's API does not return data :( if they think that it should protect from bots, then they are not very smart ofc
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "application/json")

	// Create custom transport to skip certificate verification
	// needed because otherwise the TLS connection is not established when calling from inside the container. silly workaround which just works
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
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

			fmt.Println("Found queue with id 24")
			fmt.Printf("Queue name: %s\n", queue.Name)
			fmt.Printf("Active: %t\n", queue.Active)
			fmt.Printf("Enabled: %t\n", queue.Enabled)
			fmt.Printf("Tickets left: %d\n", queue.TicketsLeft)
			fmt.Printf("Current ticket: %s\n", queue.TicketValue)
			fmt.Printf("Tickets in queue: %d\n", queue.TicketCount)

			return &queue, nil
		}
	}

	return nil, fmt.Errorf("failed to find the queue status for the queue with id: %v", odbiorKartyQueueId)
}
