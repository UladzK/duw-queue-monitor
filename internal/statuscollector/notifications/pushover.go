package notifications

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"uladzk/duw_kolejka_checker/internal/logger"
)

// PushOverNotifier provides methods to send notifications about queue status updates using the Pushover API. See API docs: https://pushover.net/api
type PushOverNotifier struct {
	cfg        *PushOverConfig
	log        *logger.Logger
	httpClient *http.Client
}

func NewPushOverNotifier(cfg *PushOverConfig, log *logger.Logger, httpClient *http.Client) *PushOverNotifier {
	return &PushOverNotifier{
		cfg:        cfg,
		log:        log,
		httpClient: httpClient,
	}
}

func (s *PushOverNotifier) SendGeneralQueueStatusUpdateNotification(queueName string, queueEnabled bool, actualTicket string, numberOfTicketsLeft int) error {
	req := url.Values{}
	req.Set("token", s.cfg.Token)
	req.Set("user", s.cfg.User)
	req.Set("message", buildQueueAvailableMsg(queueName, queueEnabled, actualTicket, numberOfTicketsLeft))

	resp, err := s.httpClient.PostForm(s.cfg.ApiUrl, req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		respTxt, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to send notification to PushOverApi. got unsuccessful status code: %d", resp.StatusCode)
		}

		return fmt.Errorf("failed to send notification to PushOverApi. got unsuccessful status code: %d, api response: \"%s\"", resp.StatusCode, respTxt)
	}

	s.log.Info("General queue status update notification sent successfully to PushOverApi.")
	defer resp.Body.Close()

	return nil
}
