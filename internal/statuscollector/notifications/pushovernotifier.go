package notifications

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	availableMsgTmpl   = "Queue %s is available! Actual ticket: %s. Number of tickets left: %d."
	unavailableMsgTmpl = "Queue %s is unavailable."
)

type PushOverNotifier struct {
	cfg        *PushOverConfig
	httpClient *http.Client
}

func NewPushOverNotifier(cfg *PushOverConfig) *PushOverNotifier {
	return &PushOverNotifier{
		cfg:        cfg,
		httpClient: &http.Client{},
	}
}

// sendGeneralQueueStatusUpdatePush sends a notification via Pushover API
func (s *PushOverNotifier) SendGeneralQueueStatusUpdatePush(queueName string, queueEnabled bool, actualTicket string, numberOfTicketsLeft int) error {
	fmt.Println(s.cfg.Token)
	fmt.Println(s.cfg.User)
	req := url.Values{}
	req.Set("token", s.cfg.Token)
	req.Set("user", s.cfg.User)
	req.Set("message", buildMsg(queueName, queueEnabled, actualTicket, numberOfTicketsLeft))

	resp, err := s.httpClient.PostForm(s.cfg.ApiUrl, req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		respTxt, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to send notification, status code: %d", resp.StatusCode)
		}

		return fmt.Errorf("failed to send notification, status code: %d, response; %s", resp.StatusCode, respTxt)
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
