package chatgpt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/satori/go.uuid"
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

func (ctx *Chat) sendRequest(accessToken string, body []byte, response any) chromedp.EmulateAction {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		return chromedp.EvaluateAsDevTools(fmt.Sprintf(`new Promise((resolve) => {
  fetch("https://chat.openai.com/backend-api/conversation", {
    "headers": {
      "authorization": "Bearer %s",
      "content-type": "application/json",
      "x-openai-assistant-app-id": ""
    },
    "body": JSON.stringify(%s),
    "method": "POST",
    "mode": "cors",
    "credentials": "include"
  })
  .then(context => context.text())
  .then(content => {
    let arr = content.split("\n\n");
    let len = arr.length;
    let index=len-3;
    if(index > -1 && index < len){
      content = arr[index];
      arr = content.split("data: ");
      if(arr.length > 1){
        resolve(JSON.parse(arr[1]));
        return;
      }
    }
    resolve(content);
  })
  .catch(e => resolve(e.toString()));
})`, accessToken, string(body)), &response, func(p *runtime.EvaluateParams) *runtime.EvaluateParams {
			return p.WithAwaitPromise(true)
		}).Do(ctx)
	})
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
		ctx.sendRequest(ctx.session.accessToken, requestBytes, &response),
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
