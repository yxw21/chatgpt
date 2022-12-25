package chatgpt

import (
	"encoding/json"
	"errors"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/satori/go.uuid"
	"time"
)

var (
	ErrorLogging   = errors.New("authorization information is being obtained")
	ErrorClearance = errors.New("clearance is being refreshed")
)

type Chat struct {
	browser *Browser
	session *Session
	cid     uuid.UUID
	pid     uuid.UUID
}

func (ctx *Chat) Check() error {
	if ctx.session.AccessTokenIsInvalid() {
		return ErrorLogging
	}
	if ctx.session.ClearanceIsInValid() {
		return ErrorClearance
	}
	return nil
}

func (ctx *Chat) Send(word string) (*Response, error) {
	var (
		cid *uuid.UUID
		pid *uuid.UUID
	)
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
	var (
		chatResponse *Response
		response     any
	)
	if err := ctx.Check(); err != nil {
		return nil, err
	}
	chatRequest := NewRequest(word, cid, pid)
	requestBytes, err := json.Marshal(chatRequest)
	if err != nil {
		return nil, err
	}
	if err = chromedp.Run(ctx.browser.Context,
		setCookie("cf_clearance", ctx.session.clearance, ".chat.openai.com", "/", time.Now().Add(7*24*time.Hour), true, true, network.CookieSameSiteNone),
		sendRequest(ctx.session.accessToken, requestBytes, &response),
	); err != nil {
		return nil, err
	}
	switch res := response.(type) {
	case string:
		return nil, errors.New(res)
	case map[string]any:
		if err = ConvertMapToStruct(res, &chatResponse); err != nil {
			return nil, errors.New("ConvertMapToStruct failed: " + err.Error())
		}
		return chatResponse, nil
	default:
		return nil, errors.New("unknown error")
	}
}

func NewChat(browser *Browser, session *Session) *Chat {
	return &Chat{browser: browser, session: session}
}
