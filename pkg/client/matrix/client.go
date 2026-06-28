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

	"github.com/google/uuid"
)

// TODO: add leading slash to API Routes

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

	URL := mc.BaseURL + fmt.Sprintf("_matrix/client/v3/rooms/%s/send/%s/%s", mm.RoomID, eventType, generateUUID())
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
	resp, err := mc.MakeRequest("POST", URL, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		slog.Error("join room error", "status", resp.Status)
		return errors.New("join room error")
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

// TODO: Ignore initial messages on first sync
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
