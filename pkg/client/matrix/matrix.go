package matrix

type PlainSender interface {
	Body() string
}

// HTMLSender is implemented by every type that is representable by plain text and formatted text (as HTML)
type HTMLSender interface {
	Body() string
	HTML() string
}

// Message is the type holding information about what and where a message is to be sent
type Message struct {
	Body          string `json:"body"`
	Format        string `json:"format,omitempty"`
	FormattedBody string `json:"formatted_body,omitempty"`
	Msgtype       string `json:"msgtype"`

	RoomID string `json:"-"`
}

func (m *Message) String() string {
	return m.Body
}

func NewPlainMessage(sender PlainSender, roomID string) *Message {
	return &Message{
		Body:    sender.Body(),
		Msgtype: "m.text",

		RoomID: roomID,
	}
}

func NewFormattedMessage(sender HTMLSender, roomID string) *Message {
	return &Message{
		Body:          sender.Body(),
		FormattedBody: sender.HTML(),
		Format:        "org.matrix.custom.html",
		Msgtype:       "m.text",

		RoomID: roomID,
	}
}
