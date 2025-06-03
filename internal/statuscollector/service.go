package statuscollector

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strconv"

	"mime/multipart"
	"net/http"
	"time"
)

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

func Run() {
	sendPushesAlways, _ := strconv.ParseBool(os.Getenv("SEND_PUSHES_ALWAYS"))                 // TODO: follow up on err handling, TODO: move to Config
	statusCheckIntervalSeconds, _ := strconv.Atoi(os.Getenv("STATUS_CHECK_INTERVAL_SECONDS")) // TODO: follow up on err handling, TODO: move to Config

	firstPushSentAlready := false
	lastTicketProcessedInQueue := ""
	for {
		pushSent, lastTicket, err := collectStatusAndPushNotifications(sendPushesAlways, firstPushSentAlready, lastTicketProcessedInQueue)
		if err != nil {
			fmt.Printf("err during collecting status and pushing notifications: %v\n", err)
		}

		firstPushSentAlready = pushSent
		lastTicketProcessedInQueue = lastTicket
		fmt.Printf("[%v] Checking again in %v seconds...\n", time.Now(), statusCheckIntervalSeconds)
		time.Sleep(time.Duration(statusCheckIntervalSeconds) * time.Second)
	}
}

func collectStatusAndPushNotifications(sendPushesAlways bool, firstPushSentAlready bool, lastTicketProcessedInQueue string) (firstPushSent bool, lastTicketProcessed string, err error) {

	queue, err := getQueueStatus()
	if err != nil {
		return firstPushSentAlready, lastTicketProcessedInQueue, err
	}

	actualTicketInQueue := queue.TicketValue
	if (queue.Active && queue.Enabled) || sendPushesAlways == true { // Q: why warning??

		pushMessage := fmt.Sprintf("Kolejka %s jest dostępna. Aktualny numer biletu to %s. Liczba biletów w kolejce to %d", queue.Name, queue.TicketValue, queue.TicketCount)

		fmt.Println(pushMessage)
		if !firstPushSentAlready {
			fmt.Println("Sending primary notification...")
			if err := sendPushoverNotification(pushMessage); err != nil {
				fmt.Printf("Error sending notification: %v\n", err)
				return firstPushSentAlready, lastTicketProcessedInQueue, err
			} else {
				return true, actualTicketInQueue, nil
			}
		} else {
			if actualTicketInQueue != lastTicketProcessedInQueue {
				fmt.Println("Sending secondary notification...")
				if err := sendPushoverNotification(pushMessage); err != nil {
					fmt.Printf("Error sending notification: %v\n", err)
					return firstPushSentAlready, lastTicketProcessedInQueue, err
				} else {
					return firstPushSentAlready, actualTicketInQueue, nil
				}
			}
		}
	} else {
		fmt.Printf("Queue is not available (active: %t, enabled: %t)\n", queue.Active, queue.Enabled)
	}

	return firstPushSentAlready, "", errors.New("something went wrong. didn't quit from the main loop")
}
