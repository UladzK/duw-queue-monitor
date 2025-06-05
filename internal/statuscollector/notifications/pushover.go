package notifications

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
)

const (
	availableMsgTmpl   = "Queue %s is available! Actual ticket: %s. Number of tickets left: %d."
	unavailableMsgTmpl = "Queue %s is unavailable."
)

type PushOverService struct {
	cfg        *PushOverConfig
	httpClient *http.Client
}

func NewPushOverService(cfg *PushOverConfig) *PushOverService {
	return &PushOverService{
		cfg:        cfg,
		httpClient: &http.Client{},
	}
}

// sendGeneralQueueStatusUpdatePush sends a notification via Pushover API
func (s *PushOverService) SendGeneralQueueStatusUpdatePush(queueName string, queueEnabled bool, actualTicket string, numberOfTicketsLeft int) error {

	var reqBuf bytes.Buffer
	writer := multipart.NewWriter(&reqBuf)

	_ = writer.WriteField("token", s.cfg.Token)
	_ = writer.WriteField("user", s.cfg.User)
	_ = writer.WriteField("message", buildMsg(queueName, queueEnabled, actualTicket, numberOfTicketsLeft))

	writer.Close()

	req, err := http.NewRequest("POST", s.cfg.ApiUrl, &reqBuf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println("Notification sent")
	return nil
}

func buildMsg(queueName string, queueEnabled bool, actualTicket string, numberOfTicketsLeft int) string {
	if !queueEnabled {
		return fmt.Sprintf(unavailableMsgTmpl, queueName)
	}

	return fmt.Sprintf(availableMsgTmpl, queueName, actualTicket, numberOfTicketsLeft)
}
