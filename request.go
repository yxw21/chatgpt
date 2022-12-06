package chatgpt

import "github.com/satori/go.uuid"

type REQContent struct {
	ContentType string   `json:"content_type"`
	Parts       []string `json:"parts"`
}

type REQMessage struct {
	ID      uuid.UUID  `json:"id"`
	Role    string     `json:"role"`
	Content REQContent `json:"content"`
}

type Request struct {
	Action          string       `json:"action"`
	ConversationId  *uuid.UUID   `json:"conversation_id"`
	Messages        []REQMessage `json:"messages"`
	ParentMessageId *uuid.UUID   `json:"parent_message_id"`
	Model           string       `json:"model"`
}

func NewRequest(word string, cid, pid *uuid.UUID) *Request {
	return &Request{
		Action:         "next",
		ConversationId: cid,
		Messages: []REQMessage{
			{
				ID:   uuid.NewV4(),
				Role: "user",
				Content: REQContent{
					ContentType: "text",
					Parts:       []string{word},
				},
			},
		},
		ParentMessageId: pid,
		Model:           "text-davinci-002-render",
	}
}
