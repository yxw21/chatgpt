package chatgpt

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Davincible/chromedp-undetected"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type Chat struct {
	client       *http.Client
	username     string
	password     string
	key          string
	cfClearance  string
	useragent    string
	sessionToken string
	session      *Session
	cid          uuid.UUID
	pid          uuid.UUID
}

func (ctx *Chat) RefreshSession() error {
	req, err := http.NewRequest("GET", "https://chat.openai.com/api/auth/session", nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", ctx.useragent)
	req.Header.Set("Cookie", fmt.Sprintf("__Secure-next-auth.session-token=%s;cf_clearance=%s", ctx.sessionToken, ctx.cfClearance))
	resp, err := ctx.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err = json.NewDecoder(resp.Body).Decode(&ctx.session); err != nil {
		return err
	}
	return nil
}

func (ctx *Chat) CheckRefreshSession() error {
	if !ctx.session.IsInvalid() {
		return nil
	}
	if ctx.username != "" && ctx.password != "" && ctx.key != "" {
		openai := NewOpenAI(ctx.username, ctx.password, ctx.key)
		sessionToken, err := openai.GetToken()
		if err != nil {
			return err
		}
		ctx.sessionToken = sessionToken
	}
	if err := ctx.RefreshSession(); err != nil {
		return err
	}
	if ctx.session.Error != "" {
		return errors.New(ctx.session.Error)
	}
	return nil
}

func (ctx *Chat) AutoRefreshSession() {
	go func() {
		for {
			ctx.CheckRefreshSession()
			time.Sleep(time.Second * 10)
		}
	}()
}

func (ctx *Chat) CheckCfClearance() (string, error) {
	var cfClearance string
	chromeContext, cancel, err := chromedpundetected.New(chromedpundetected.NewConfig(
		chromedpundetected.WithHeadless(),
		chromedpundetected.WithTimeout(20*time.Second),
	))
	if err != nil {
		return cfClearance, err
	}
	defer cancel()
	if err = chromedp.Run(chromeContext,
		chromedp.Navigate("https://chat.openai.com/chat"),
		chromedp.WaitVisible(`//div[@id="__next"]`),
		chromedp.Evaluate(`navigator.userAgent`, &ctx.useragent),
		chromedp.ActionFunc(func(browserCtx context.Context) error {
			cookies, err := network.GetAllCookies().Do(browserCtx)
			if err != nil {
				return err
			}
			for _, cookie := range cookies {
				if cookie.Name == "cf_clearance" && cookie.Domain == ".chat.openai.com" {
					ctx.cfClearance = cookie.Value
					break
				}
			}
			return nil
		}),
	); err != nil {
		return cfClearance, err
	}
	return cfClearance, nil
}

func (ctx *Chat) AutoRefreshCF() {
	go func() {
		for {
			_, _ = ctx.CheckCfClearance()
			if ctx.cfClearance != "" {
				ctx.AutoRefreshSession()
			}
			rand.Seed(time.Now().UnixNano())
			t := time.Duration(rand.Intn(300) + 60)
			time.Sleep(t * time.Second)
		}
	}()
}

func (ctx *Chat) Send(word string) (*Response, error) {
	var (
		cid *uuid.UUID
		pid *uuid.UUID
	)
	if err := ctx.CheckRefreshSession(); err != nil {
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
	if err := ctx.CheckRefreshSession(); err != nil {
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
	req.Header.Set("User-Agent", ctx.useragent)
	req.Header.Set("Authorization", "Bearer "+ctx.session.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "cf_clearance="+ctx.cfClearance)
	resp, err := ctx.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	arr := strings.Split(string(responseBytes), "\n\n")
	index := len(arr) - 3
	if index >= len(arr) || index < 1 {
		return nil, errors.New(string(responseBytes))
	}
	arr = strings.Split(arr[index], "data: ")
	if len(arr) < 2 {
		return nil, errors.New(string(responseBytes))
	}
	if err = json.Unmarshal([]byte(arr[1]), &chatResponse); err != nil {
		return nil, err
	}
	return chatResponse, nil
}

func NewChat(username, password, key string) *Chat {
	chat := &Chat{client: &http.Client{}, username: username, password: password, key: key, session: &Session{}}
	chat.AutoRefreshCF()
	return chat
}

func NewChatWithSessionToken(sessionToken string) *Chat {
	chat := &Chat{client: &http.Client{}, sessionToken: sessionToken, session: &Session{}}
	chat.AutoRefreshCF()
	return chat
}
