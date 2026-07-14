package matrix

import (
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/jhachmer/happahappa/pkg/config"
)

type Command interface {
	Name() string
	Execute(roomID string, args []string) error
}

type CommandFunc func(roomID string, args []string) error

type CommandHandler struct {
	client Client

	commandMap map[string]CommandFunc
}

func NewCommandHandler(cfg *config.Config, client Client) (*CommandHandler, error) {
	if cfg.Matrix.RoomID == "" {
		return nil, errors.New("no room id was given")
	}
	return &CommandHandler{
		client:     client,
		commandMap: make(map[string]CommandFunc),
	}, nil
}

func (c CommandHandler) Register(command Command) {
	c.commandMap[command.Name()] = command.Execute
}

func (c CommandHandler) HandleCommands() {
	slog.Info("Listening for commands...")
	since := ""
	// sync once to avoid sending messages for old events
	syncResp, err := c.client.Sync(since)
	if err != nil {
		slog.Error("could not sync chat state", "since", since, "err", err)
	}
	since = syncResp.NextBatch
	for {
		syncResp, err := c.client.Sync(since)
		if err != nil {
			slog.Error("could not sync chat state", "since", since, "err", err)
			time.Sleep(10 * time.Second)
			continue
		}
		since = syncResp.NextBatch

		for roomID, room := range syncResp.Rooms.Join {
			for _, event := range room.Timeline.Events {
				if event.Type != MessageEventType {
					continue
				}
				body := strings.TrimSpace(event.Content.Body)
				if !strings.HasPrefix(body, "!") {
					continue
				}
				// TODO: probably still wanna allow commands that do not need args
				messageParts := strings.SplitN(body, " ", 2)
				if len(messageParts) != 2 {
					slog.Warn("not enough parts in body", "body", body)
					continue
				}
				command := strings.TrimPrefix(messageParts[0], "!")
				args := strings.Split(messageParts[1], ",")
				handler, ok := c.commandMap[command]
				if !ok {
					continue
				}

				if err := handler(roomID, args); err != nil {
					slog.Error(
						"command failed",
						"command", command,
						"err", err,
					)
				}
			}
		}
	}
}
