package mattermost

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"

	"lab.it.hs-hannover.de/8mg-y3w-u2/happahappa/pkg/config"
)

type Message interface {
	Send(url string) error
}

type MattermostMessage struct {
	Text      string `json:"text"`
	Username  string `json:"username,omitempty"`
	IconEmoji string `json:"icon_emoji,omitempty"`
	Channel   string `json:"channel,omitempty"`
}

func NewMattermostMessage(text string, config *config.Config) MattermostMessage {
	return MattermostMessage{
		Text:      text,
		Username:  config.Mattermost.Username,
		IconEmoji: config.Mattermost.IconEmoji,
		Channel:   config.Mattermost.Channel,
	}
}

func (m MattermostMessage) Send(webhookUrl string) error {
	payload, err := json.Marshal(m)
	if err != nil {
		return err
	}
	body := bytes.NewReader(payload)

	req, err := http.NewRequest("POST", webhookUrl, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	slog.Info("Send Mattermost message", "status", resp.Status)
	resp.Body.Close()

	return nil
}
