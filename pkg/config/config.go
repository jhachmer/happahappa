package config

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Canteen    *Canteen          `yaml:"canteen"`
	Weather    *Weather          `yaml:"weather"`
	Mattermost *MattermostConfig `yaml:"mattermost_server,omitempty"`
	Matrix     *MatrixConfig     `yaml:"matrix_server,omitempty"`
}

func (c Config) String() string {
	return fmt.Sprintf("Canteen: %s\nMattermost: %s", c.Canteen, c.Mattermost)
}

type Canteen struct {
	CanteenUrl string `yaml:"canteen_url"`
	PriceId    int    `yaml:"price_id"`
	PayId      int    `yaml:"pay_id"`
	CanteenId  int    `yaml:"canteen_id"`
}

func (c Canteen) String() string {
	return fmt.Sprintf("Canteen URL: %s\nPrice ID: %d\nPay ID: %d\nCanteen ID: %d", c.CanteenUrl, c.PriceId, c.PayId, c.CanteenId)
}

type Weather struct {
	Lat string `yaml:"lat"`
	Lon string `yaml:"lon"`
}

func (w Weather) String() string {
	return fmt.Sprintf("Lat: %s° Lon: %s°", w.Lat, w.Lon)
}

type MattermostConfig struct {
	WebhookUrl string `yaml:"webhook_url"`
	Username   string `yaml:"username,omitempty"`
	IconEmoji  string `yaml:"icon_emoji,omitempty"`
	Channel    string `yaml:"channel,omitempty"`
}

func (m MattermostConfig) String() string {
	return fmt.Sprintf("Webhook URL: %s\nUsername: %s\nIcon Emoji: %s\nChannel: %s", m.WebhookUrl, m.Username, m.IconEmoji, m.Channel)
}

type MatrixConfig struct {
	BaseUrl     string `yaml:"base_url"`
	UserID      string `yaml:"user_id"`
	AccessToken string `yaml:"access_token"`
	RoomID      string `yaml:"room_id"`
}

func (m MatrixConfig) String() string {
	return fmt.Sprintf("UserID: %s", m.UserID)
}

func ReadConfig(file string) (*Config, error) {
	var config Config
	yamlFile, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("error reading yaml file")
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling yaml contents")
	}
	return &config, nil
}
