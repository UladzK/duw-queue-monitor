package notifications

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"uladzk/duw_kolejka_checker/internal/logger"
)

type TelegramNotifier struct {
	cfg        *TelegramConfig
	log        *logger.Logger
	httpClient *http.Client
}

func NewTelegramNotifier(cfg *TelegramConfig, log *logger.Logger, httpClient *http.Client) *TelegramNotifier {
	return &TelegramNotifier{
		cfg:        cfg,
		log:        log,
		httpClient: httpClient,
	}
}

func (s *TelegramNotifier) SendGeneralQueueStatusUpdatePush(queueName string, queueEnabled bool, actualTicket string, numberOfTicketsLeft int) error {
	channelName := fmt.Sprintf("@%s", s.cfg.BroadcastChannelName)
	botApiUrl := fmt.Sprintf("%s/bot%s/sendMessage", s.cfg.ApiUrl, s.cfg.BotToken)

	req := url.Values{}
	req.Set("chat_id", channelName)
	req.Set("text", buildQueueAvailableMsg(queueName, queueEnabled, actualTicket, numberOfTicketsLeft))

	resp, err := s.httpClient.PostForm(botApiUrl, req) // TODO: application/json is better
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		respTxt, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to send notification to TelegramApi. got unsuccessful status code: %d", resp.StatusCode)
		}

		return fmt.Errorf("failed to send notification to TelegramApi. got unsuccessful status code: %d, api response: \"%s\"", resp.StatusCode, respTxt)
	}

	s.log.Info("General queue status update notification sent successfully to TelegramApi.")
	defer resp.Body.Close()

	return nil
}
