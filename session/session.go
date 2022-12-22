package chatgpt

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/Davincible/chromedp-undetected"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/yxw21/chatgpt/openai"
	"math/rand"
	"strings"
	"time"
)

type Session struct {
	username    string
	password    string
	key         string
	accessToken string
	clearance   string
	useragent   string
}

func (ctx *Session) GetAccessToken() string {
	return ctx.accessToken
}

func (ctx *Session) GetClearance() string {
	return ctx.clearance
}

func (ctx *Session) GetUserAgent() string {
	return ctx.useragent
}

func (ctx *Session) IsInvalid() bool {
	var user User
	if ctx.accessToken == "" {
		return true
	}
	sli := strings.Split(ctx.accessToken, ".")
	if len(sli) != 3 {
		return true
	}
	bs, err := base64.StdEncoding.DecodeString(sli[1])
	if err != nil {
		return true
	}
	if err = json.Unmarshal(bs, &user); err != nil {
		return true
	}
	return (time.Now().Unix() + 60) >= time.Unix(user.Exp, 0).Unix()
}

func (ctx *Session) RefreshToken() {
	if ctx.username != "" && ctx.password != "" && ctx.key != "" {
		data, err := chatgpt.NewOpenAI(ctx.username, ctx.password, ctx.key).GetData()
		if err == nil {
			ctx.accessToken = data.AccessToken
			ctx.clearance = data.Clearance
			ctx.useragent = data.Useragent
		}
	}
}

func (ctx *Session) refreshClearance() {
	var (
		cookies []*network.Cookie
		err     error
	)
	chromeContext, cancel, err := chromedpundetected.New(chromedpundetected.NewConfig(
		chromedpundetected.WithHeadless(),
		chromedpundetected.WithTimeout(20*time.Second),
	))
	if err != nil {
		return
	}
	defer cancel()
	if err = chromedp.Run(chromeContext,
		chromedp.Navigate("https://chat.openai.com/chat"),
		chromedp.WaitVisible(`//div[@id="__next"]`),
		chromedp.Evaluate(`navigator.userAgent`, &ctx.useragent),
		chromedp.ActionFunc(func(browserCtx context.Context) error {
			cookies, err = network.GetAllCookies().Do(browserCtx)
			if err != nil {
				return err
			}
			for _, cookie := range cookies {
				if cookie.Name == "cf_clearance" && cookie.Domain == ".chat.openai.com" {
					ctx.clearance = cookie.Value
					break
				}
			}
			return nil
		}),
	); err != nil {
		return
	}
	return
}

func (ctx *Session) autoRefreshToken() {
	for {
		if ctx.IsInvalid() {
			ctx.RefreshToken()
		}
		time.Sleep(time.Second * 10)
	}
}

func (ctx *Session) autoRefreshClearance() {
	for {
		ctx.refreshClearance()
		rand.Seed(time.Now().UnixNano())
		t := time.Duration(rand.Intn(300) + 60)
		time.Sleep(t * time.Second)
	}
}

func (ctx *Session) AutoRefresh() *Session {
	go ctx.autoRefreshToken()
	go ctx.autoRefreshClearance()
	return ctx
}

func NewSessionWithCredential(username, password, key string) *Session {
	return &Session{
		username: username,
		password: password,
		key:      key,
	}
}

func NewSessionWithAccessToken(accessToken string) *Session {
	return &Session{
		accessToken: accessToken,
	}
}
