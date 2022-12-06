package chatgpt

import (
	"github.com/satori/go.uuid"
	"time"
)

type RESContent struct {
	ContentType string   `json:"content_type"`
	Parts       []string `json:"parts"`
}

type RESMessage struct {
	Content    RESContent     `json:"content"`
	CreateTime time.Time      `json:"create_time"`
	EndTurn    any            `json:"end_turn"`
	ID         uuid.UUID      `json:"id"`
	Metadata   map[string]any `json:"metadata"`
	Recipient  string         `json:"recipient"`
	Role       string         `json:"role"`
	UpdateTime time.Time      `json:"update_time"`
	User       any            `json:"user"`
	Weight     float64        `json:"weight"`
}

type Response struct {
	Message        RESMessage `json:"message"`
	ConversationId uuid.UUID  `json:"conversation_id"`
	Error          any        `json:"error"`
}
