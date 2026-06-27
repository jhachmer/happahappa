package matrix

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/jhachmer/happahappa/pkg/config"
	"github.com/jhachmer/happahappa/pkg/data/station"
)

const MessageEventType = "m.room.message"

var MessageBatch = make([]*MatrixMessage, 0)

type Client struct {
	BaseURL     string
	Credentials *Credentials
}

func NewMatrixClient(baseURL string, credentials *Credentials) Client {
	return Client{
		BaseURL:     baseURL,
		Credentials: credentials,
	}
}

func (mc Client) SendMessage(mm *MatrixMessage) error {
	const eventType = "m.room.message"

	payload, err := json.Marshal(mm)
	if err != nil {
		return err
	}

	URL := mc.BaseURL + fmt.Sprintf("/_matrix/client/v3/rooms/%s/send/%s/%s", mm.RoomID, eventType, generateUUID())
	reqBody := bytes.NewReader(payload)
	resp, err := mc.MakeRequest("PUT", URL, reqBody)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("matrix server responded with %d", resp.StatusCode)
	}
	slog.Info("Matrix response", "status", resp.Status)

	return nil
}

func (mc Client) SendMessageBatch(messages []*MatrixMessage) {
	for i, message := range messages {
		err := mc.SendMessage(message)
		if err != nil {
			slog.Error("failed message in batch", "err", err.Error())
		}
		slog.Info("send message from batch", "current", i, "total", len(messages), "message", message)
	}
}

func generateUUID() string {
	return uuid.New().String()
}

func (mc Client) JoinRoom(roomID string) error {
	URL := mc.BaseURL + fmt.Sprintf("_matrix/client/v3/rooms/%s/join", roomID)
	_, err := mc.MakeRequest("POST", URL, nil)
	if err != nil {
		return err
	}
	slog.Info("Added user to room", "roomID", roomID)
	return nil
}

type InviteRequest struct {
	UserID string `json:"user_id"`
}

func (mc Client) InviteToRoom(roomID, userID string) error {
	URL := mc.BaseURL + fmt.Sprintf("_matrix/client/v3/rooms/%s/join", roomID)
	inviteRequest := InviteRequest{UserID: userID}
	payload, err := json.Marshal(inviteRequest)
	if err != nil {
		return err
	}
	body := bytes.NewReader(payload)
	_, err = mc.MakeRequest("POST", URL, body)
	if err != nil {
		return err
	}
	slog.Info("Invited user to room", "roomID", roomID, "user_id", userID)
	return nil
}

func (mc Client) MakeRequest(method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+mc.Credentials.AccessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (mc Client) Register(message *MatrixMessage) {
	MessageBatch = append(MessageBatch, message)
}

func (mc Client) SendRegistered() {
	mc.SendMessageBatch(MessageBatch)
}

type SyncResponse struct {
	NextBatch string `json:"next_batch"`
	Rooms     struct {
		Join map[string]JoinedRoom `json:"join"`
	} `json:"rooms"`
}

type JoinedRoom struct {
	Timeline struct {
		Events []Event `json:"events"`
	} `json:"timeline"`
}

type Event struct {
	Type    string `json:"type"`
	Sender  string `json:"sender"`
	EventID string `json:"event_id"`
	Content struct {
		MsgType string `json:"msgtype"`
		Body    string `json:"body"`
	} `json:"content"`
}

func (mc Client) Sync(since string) (*SyncResponse, error) {
	URL := mc.BaseURL + "_matrix/client/v3/sync"
	if since != "" {
		URL += "?since=" + url.QueryEscape(since) + "&timeout=30000"
	} else {
		URL += "?timeout=30000"
	}

	resp, err := mc.MakeRequest("GET", URL, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var syncResponse SyncResponse
	err = json.NewDecoder(resp.Body).Decode(&syncResponse)
	if err != nil {
		return nil, err
	}
	return &syncResponse, nil
}

type CommandHandler struct {
	client         Client
	roomID         string
	stationScraper *station.StationScraper
}

func NewCommandHandler(cfg *config.Config, client Client) (*CommandHandler, error) {
	if cfg.Matrix.RoomID == "" {
		return nil, errors.New("no room id was given")
	}
	stationScraper, err := station.NewStationScraper(cfg)
	if err != nil {
		return nil, errors.New("could not construct station scraper")
	}
	return &CommandHandler{
		client:         client,
		roomID:         cfg.Matrix.RoomID,
		stationScraper: stationScraper,
	}, nil
}

func (c CommandHandler) HandleCommands() {
	since := ""

	for {
		syncResp, err := c.client.Sync(since)
		if err != nil {
			slog.Error("could not sync chat state", "since", since, "err", err)
			time.Sleep(10 * time.Second)
			continue
		}
		since = syncResp.NextBatch

		fmt.Println(syncResp)
		for roomID, room := range syncResp.Rooms.Join {
			for _, event := range room.Timeline.Events {
				if event.Type != MessageEventType {
					continue
				}
				fmt.Println(event.Content.Body)
				switch event.Content.Body {
				case "!abfahrt":
					fmt.Println("in", roomID)
				}
			}
		}
	}
}
