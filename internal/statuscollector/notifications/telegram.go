package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

type SendMessageChannelRequest struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

func (s *TelegramNotifier) SendGeneralQueueStatusUpdatePush(queueName string, queueEnabled bool, actualTicket string, numberOfTicketsLeft int) error {
	channelName := fmt.Sprintf("@%s", s.cfg.BroadcastChannelName)
	botApiFullUrl := fmt.Sprintf("%s/bot%s/sendMessage", s.cfg.BaseApiUrl, s.cfg.BotToken)

	reqBody := SendMessageChannelRequest{
		ChatID: channelName,
		Text:   buildQueueAvailableMsg(queueName, queueEnabled, actualTicket, numberOfTicketsLeft),
	}
	b, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	resp, err := s.httpClient.Post(botApiFullUrl, "application/json", bytes.NewBuffer(b))
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
