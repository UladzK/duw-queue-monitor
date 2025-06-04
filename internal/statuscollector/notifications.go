package statuscollector

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
)

// sendPushoverNotification sends a notification via Pushover API
func sendPushoverNotification(message string) error {
	url := "https://api.pushover.net/1/messages.json" //TODO: move to Config

	var reqBuf bytes.Buffer
	writer := multipart.NewWriter(&reqBuf)

	_ = writer.WriteField("token", "aay6otxvgv5zwkwck6r6r6bch4qucs")
	_ = writer.WriteField("user", "uun179bk9o34gn7tg3qk8s4jt8d4i5")
	_ = writer.WriteField("message", message)

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

	fmt.Println("Notification sent: Queue is available!")
	return nil
}
