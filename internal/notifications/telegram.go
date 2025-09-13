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
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"` // needed to correctly format the message in Telegram
}

// TODO: add retries
func (s *TelegramNotifier) SendMessage(chatID, text string) error {
	botApiFullUrl := fmt.Sprintf("%s/bot%s/sendMessage", s.cfg.BaseApiUrl, s.cfg.BotToken)

	reqBody := SendMessageChannelRequest{
		ChatID:    chatID,
		Text:      text,
		ParseMode: parseMode,
	}
	b, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body when sending message to TelegramApi: %w", err)
	}

	resp, err := s.httpClient.Post(botApiFullUrl, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("failed to send message to TelegramApi: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respTxt, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body when sending message to TelegramApi. got unsuccessful status code: %d", resp.StatusCode)
		}

		return fmt.Errorf("sending message to TelegramApi failed. got unsuccessful status code: %d, api response: \"%s\"", resp.StatusCode, respTxt)
	}

	s.log.Info("Message sent successfully to TelegramApi.")
	return nil
}

func (s *TelegramNotifier) SendGeneralQueueStatusUpdateNotification(broadcastChannelName, queueName string, queueActive bool, queueEnabled bool, actualTicket string, numberOfTicketsLeft int) error {
	channelName := fmt.Sprintf("@%s", broadcastChannelName)
	message := buildQueueAvailableMsg(queueName, queueEnabled, actualTicket, numberOfTicketsLeft)

	if err := s.SendMessage(channelName, message); err != nil {
		return err
	}

	s.log.Info("General queue status update notification sent successfully to TelegramApi.")
	return nil
}
