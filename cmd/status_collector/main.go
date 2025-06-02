package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"

	"io"
	"mime/multipart"
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

// sendPushoverNotification sends a notification via Pushover API
func sendPushoverNotification(message string) error {
	url := "https://api.pushover.net/1/messages.json"

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	_ = writer.WriteField("token", "aay6otxvgv5zwkwck6r6r6bch4qucs")
	_ = writer.WriteField("user", "uun179bk9o34gn7tg3qk8s4jt8d4i5")
	_ = writer.WriteField("message", message)

	writer.Close()

	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println("Notification sent: Queue is available!")
	return nil
}

func main() {
	sendPushesAlways := os.Getenv("SEND_PUSHES_ALWAYS")
	firstPushSentAlready := false
	lastTicketProcessedInQueue := ""
	for {
		req, err := http.NewRequest("GET", "https://rezerwacje.duw.pl/status_kolejek/query.php?status=", nil)
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			time.Sleep(10 * time.Second)
			continue
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
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			fmt.Printf("got %v status code\n", resp.StatusCode)
			break
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error reading response body: %v\n", err)
			time.Sleep(10 * time.Second)
			continue
		}

		var response Response
		if err := json.Unmarshal(body, &response); err != nil {
			fmt.Printf("Error parsing JSON: %v\n", err)
			time.Sleep(10 * time.Second)
			continue
		}

		found := false
		for _, queue := range response.Result["Wrocław"] {
			if queue.ID == 24 {
				found = true

				fmt.Println("Found queue with id 24")
				fmt.Printf("Queue name: %s\n", queue.Name)
				fmt.Printf("Active: %t\n", queue.Active)
				fmt.Printf("Enabled: %t\n", queue.Enabled)
				fmt.Printf("Tickets left: %d\n", queue.TicketsLeft)
				fmt.Printf("Current ticket: %s\n", queue.TicketValue)
				fmt.Printf("Tickets in queue: %d\n", queue.TicketCount)

				actualTicketInQueue := queue.TicketValue
				if (queue.Active && queue.Enabled) || sendPushesAlways == "true" {

					pushMessage := fmt.Sprintf("Kolejka %s jest dostępna. Aktualny numer biletu to %s. Liczba biletów w kolejce to %d", queue.Name, queue.TicketValue, queue.TicketCount)

					fmt.Println(pushMessage)
					if !firstPushSentAlready {
						fmt.Println("Sending primary notification...")
						if err := sendPushoverNotification(pushMessage); err != nil {
							fmt.Printf("Error sending notification: %v\n", err)
						} else {
							firstPushSentAlready = true
							lastTicketProcessedInQueue = actualTicketInQueue
						}
					} else {
						if actualTicketInQueue != lastTicketProcessedInQueue {
							fmt.Println("Sending secondary notification...")
							if err := sendPushoverNotification(pushMessage); err != nil {
								fmt.Printf("Error sending notification: %v\n", err)
							} else {
								lastTicketProcessedInQueue = actualTicketInQueue
							}
						}
					}
				} else {
					fmt.Printf("Queue is not available (active: %t, enabled: %t)\n",
						queue.Active, queue.Enabled)
				}
				break
			}
		}

		if !found {
			fmt.Println("Queue with id 24 not found")
		}

		fmt.Printf("[%v] Checking again in 10 seconds...\n", time.Now())
		time.Sleep(10 * time.Second)
	}
}
