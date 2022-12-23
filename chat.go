package chatgpt

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/satori/go.uuid"
	"github.com/yxw21/chatgpt/session"
	"io"
	"log"
	"net/http"
	"strings"
)

type Chat struct {
	client  *http.Client
	session *chatgpt.Session
	cid     uuid.UUID
	pid     uuid.UUID
}

func (ctx *Chat) check() error {
	if ctx.session.IsInvalid() {
		return ctx.session.RefreshToken()
	}
	if ctx.session.GetClearance() == "" {
		return ctx.session.RefreshClearance()
	}
	return nil
}

func (ctx *Chat) Send(word string) (*Response, error) {
	var (
		cid *uuid.UUID
		pid *uuid.UUID
	)
	if err := ctx.check(); err != nil {
		return nil, err
	}
	if ctx.cid != uuid.Nil {
		cid = &ctx.cid
	}
	if ctx.pid == uuid.Nil {
		tid := uuid.NewV4()
		pid = &tid
	} else {
		pid = &ctx.pid
	}
	res, err := ctx.SendMessage(word, cid, pid)
	if err != nil {
		return nil, err
	}
	ctx.cid = res.ConversationId
	ctx.pid = res.Message.ID
	return res, nil
}

func (ctx *Chat) SendMessage(word string, cid, pid *uuid.UUID) (*Response, error) {
	var chatResponse *Response
	if err := ctx.check(); err != nil {
		return nil, err
	}
	chatRequest := NewRequest(word, cid, pid)
	requestBytes, err := json.Marshal(chatRequest)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", "https://chat.openai.com/backend-api/conversation", bytes.NewBuffer(requestBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", ctx.session.GetUserAgent())
	req.Header.Set("Authorization", "Bearer "+ctx.session.GetAccessToken())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "cf_clearance="+ctx.session.GetClearance())
	resp, err := ctx.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	arr := strings.Split(string(responseBytes), "\n\n")
	index := len(arr) - 3
	if index >= len(arr) || index < 1 {
		log.Println(ctx.session)
		return nil, errors.New(string(responseBytes))
	}
	arr = strings.Split(arr[index], "data: ")
	if len(arr) < 2 {
		log.Println(ctx.session)
		return nil, errors.New(string(responseBytes))
	}
	if err = json.Unmarshal([]byte(arr[1]), &chatResponse); err != nil {
		return nil, err
	}
	return chatResponse, nil
}

func NewChat(session *chatgpt.Session) *Chat {
	return &Chat{client: &http.Client{}, session: session}
}
