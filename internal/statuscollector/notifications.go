package statuscollector

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
)

// sendGeneralQueueStatusUpdatePush sends a notification via Pushover API
func sendGeneralQueueStatusUpdatePush(queueName string, queueEnabled bool, actualTicket string, numberOfTicketsLeft int) error {

	url := "https://api.pushover.net/1/messages.json" //TODO: move to Config

	var reqBuf bytes.Buffer
	writer := multipart.NewWriter(&reqBuf)

	_ = writer.WriteField("token", "aay6otxvgv5zwkwck6r6r6bch4qucs")
	_ = writer.WriteField("user", "uun179bk9o34gn7tg3qk8s4jt8d4i5")
	_ = writer.WriteField("message", getMessage(queueName, queueEnabled, actualTicket, numberOfTicketsLeft))

	writer.Close()

	req, err := http.NewRequest("POST", url, &reqBuf)
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

	fmt.Println("Notification sent")
	return nil
}

func getMessage(queueName string, queueEnabled bool, actualTicket string, numberOfTicketsLeft int) string {
	const availableMsgTmpl = "Queue %s is available! Actual ticket: %s. Number of tickets left: %d."
	const unavailableMsgTmpl = "Queue %s is unavailable."

	if !queueEnabled {
		return fmt.Sprintf(unavailableMsgTmpl, queueName)
	}

	return fmt.Sprintf(availableMsgTmpl, queueName, actualTicket, numberOfTicketsLeft)
}
