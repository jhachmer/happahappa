package matrix

type MatrixPlainSender interface {
	Body() string
}

// MatrixHTMLSender is implemented by every type that is representable by plain text and formatted text (as HTML)
type MatrixHTMLSender interface {
	Body() string
	HTML() string
}

// MatrixMessage is the type holding information about what and where a message is to be sent
type MatrixMessage struct {
	Body          string `json:"body"`
	Format        string `json:"format,omitempty"`
	FormattedBody string `json:"formatted_body,omitempty"`
	Msgtype       string `json:"msgtype"`

	RoomID string `json:"-"`
}

func NewFormattedMatrixMessage(body, formattedBody, roomID string) *MatrixMessage {
	return &MatrixMessage{
		Body:          body,
		FormattedBody: formattedBody,
		Format:        "org.matrix.custom.html",
		Msgtype:       "m.text",

		RoomID: roomID,
	}
}

func NewPlainMatrixMessage(body, roomID string) *MatrixMessage {
	return &MatrixMessage{
		Body:    body,
		Msgtype: "m.text",

		RoomID: roomID,
	}
}

func NewMatrixMessageFromSender(sender MatrixHTMLSender, roomID string) *MatrixMessage {
	return &MatrixMessage{
		Body:          sender.Body(),
		FormattedBody: sender.HTML(),
		Format:        "org.matrix.custom.html",
		Msgtype:       "m.text",

		RoomID: roomID,
	}
}
