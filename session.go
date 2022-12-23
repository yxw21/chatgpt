package chatgpt

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/Davincible/chromedp-undetected"
	"github.com/chromedp/chromedp"
	"strings"
	"time"
)

const cfExpire = 120 * time.Minute

type Session struct {
	username         string
	password         string
	key              string
	accessToken      string
	clearance        string
	useragent        string
	clearanceCreated int64
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

func (ctx *Session) AccessTokenIsInvalid() bool {
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

func (ctx *Session) ClearanceIsInValid() bool {
	now := time.Now().Unix()
	val := now - ctx.clearanceCreated
	// Refresh token 10 minutes in advance ()
	t := cfExpire - (10 * time.Minute)
	return val > int64(t.Seconds())
}

func (ctx *Session) RefreshToken() error {
	if ctx.username != "" && ctx.password != "" && ctx.key != "" {
		passport, err := NewOpenAI(ctx.username, ctx.password, ctx.key).GetPassport()
		if err != nil {
			return err
		}
		ctx.accessToken = passport.AccessToken
		ctx.clearance = passport.Clearance
		ctx.useragent = passport.Useragent
		ctx.clearanceCreated = time.Now().Unix()
	}
	return nil
}

func (ctx *Session) RefreshClearance() error {
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
		waitElement(".btn:nth-child(1)", time.Minute),
		chromedp.Evaluate(`navigator.userAgent`, &ctx.useragent),
		setClearance(&ctx.clearance),
	); err != nil {
		return errors.New("error refreshing clearance: " + err.Error())
	}
	ctx.clearanceCreated = time.Now().Unix()
	return nil
}

func (ctx *Session) AutoRefresh() *Session {
	go func() {
		for {
			if ctx.AccessTokenIsInvalid() {
				_ = ctx.RefreshToken()
			} else {
				if ctx.ClearanceIsInValid() {
					_ = ctx.RefreshClearance()
				}
			}
			time.Sleep(time.Second * 10)
		}
	}()
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