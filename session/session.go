package chatgpt

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/Davincible/chromedp-undetected"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/yxw21/chatgpt/openai"
	"strings"
	"time"
)

type Session struct {
	username    string
	password    string
	key         string
	accessToken string
	clearance   string
	cfTime      int64
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
	// Refresh token 5 minutes in advance
	return (time.Now().Unix() + 300) >= time.Unix(user.Exp, 0).Unix()
}

func (ctx *Session) RefreshToken() error {
	if ctx.username != "" && ctx.password != "" && ctx.key != "" {
		data, err := chatgpt.NewOpenAI(ctx.username, ctx.password, ctx.key).GetData()
		if err != nil {
			return err
		}
		ctx.accessToken = data.AccessToken
		ctx.clearance = data.Clearance
		ctx.useragent = data.Useragent
		ctx.cfTime = time.Now().Unix()
	}
	return nil
}

func (ctx *Session) RefreshClearance() error {
	var (
		cookies []*network.Cookie
		err     error
	)
	chromeContext, cancel, err := chromedpundetected.New(chromedpundetected.NewConfig(
		chromedpundetected.WithHeadless(),
		chromedpundetected.WithTimeout(20*time.Second),
	))
	if err != nil {
		return errors.New("error creating chrome context (cf): " + err.Error())
	}
	defer cancel()
	if err = chromedp.Run(chromeContext,
		chromedp.Navigate("https://chat.openai.com/auth/login"),
		chromedp.WaitVisible(`//div[@id="__next"]`),
		chromedp.Evaluate(`navigator.userAgent`, &ctx.useragent),
		chromedp.ActionFunc(func(browserCtx context.Context) error {
			cookies, err = network.GetAllCookies().Do(browserCtx)
			if err != nil {
				return errors.New("error getting cookie (cf): " + err.Error())
			}
			for _, cookie := range cookies {
				if cookie.Name == "cf_clearance" && cookie.Domain == ".chat.openai.com" {
					ctx.clearance = cookie.Value
					ctx.cfTime = time.Now().Unix()
					break
				}
			}
			return nil
		}),
	); err != nil {
		return errors.New("error refreshing clearance: " + err.Error())
	}
	return nil
}

func (ctx *Session) autoRefreshToken() {
	for {
		if ctx.IsInvalid() {
			_ = ctx.RefreshToken()
		}
		time.Sleep(time.Second * 10)
	}
}

func (ctx *Session) autoRefreshClearance() {
	for {
		now := time.Now().Unix()
		val := now - ctx.cfTime
		// https://support.cloudflare.com/hc/en-us/articles/360038470312#4C6RjJMNCGMUpBYm0vCYj1
		// Refresh token 3 minutes in advance ()
		t := (30 * time.Minute) - (3 * time.Minute)
		if val > int64(t) {
			_ = ctx.RefreshClearance()
		}
		time.Sleep(time.Second * 10)
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
