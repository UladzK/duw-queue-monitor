package statuscollector

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"time"
)

func Run() {
	sendPushesAlways, _ := strconv.ParseBool(os.Getenv("SEND_PUSHES_ALWAYS"))                 // TODO: follow up on err handling, TODO: move to Config
	statusCheckIntervalSeconds, _ := strconv.Atoi(os.Getenv("STATUS_CHECK_INTERVAL_SECONDS")) // TODO: follow up on err handling, TODO: move to Config

	firstPushSentAlready := false
	lastTicketProcessedInQueue := ""
	for {
		pushSent, lastTicket, err := checkAndProcessStatus(sendPushesAlways, firstPushSentAlready, lastTicketProcessedInQueue)
		if err != nil {
			fmt.Printf("err during collecting status and pushing notifications: %v\n", err)
		}

		firstPushSentAlready = pushSent
		lastTicketProcessedInQueue = lastTicket
		fmt.Printf("[%v] Checking again in %v seconds...\n", time.Now(), statusCheckIntervalSeconds)
		time.Sleep(time.Duration(statusCheckIntervalSeconds) * time.Second)
	}
}

func checkAndProcessStatus(sendPushesAlways bool, firstPushSentAlready bool, lastTicketProcessedInQueue string) (firstPushSent bool, lastTicketProcessed string, err error) {

	queueStatus, err := getQueueStatus()
	if err != nil {
		return firstPushSentAlready, lastTicketProcessedInQueue, err
	}

	//TODO: rewrite to work on status changes
	actualTicketInQueue := queueStatus.TicketValue
	if queueStatus.Enabled || sendPushesAlways {

		pushMessage := fmt.Sprintf("Kolejka %s jest dostępna. Aktualny numer biletu to %s. Liczba biletów w kolejce to %d", queueStatus.Name, queueStatus.TicketValue, queueStatus.TicketCount)

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
		fmt.Printf("Queue is not available (active: %t, enabled: %t)\n", queueStatus.Active, queueStatus.Enabled)
	}

	return firstPushSentAlready, "", errors.New("something went wrong. didn't quit from the main loop")
}
