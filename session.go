package chatgpt

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"log"
	"os"
	"strings"
	"time"
)

const cfExpire = 120 * time.Minute

type Session struct {
	browser          *Browser
	username         string
	password         string
	accessToken      string
	clearance        string
	clearanceCreated int64
}

func (ctx *Session) GetAccessToken() string {
	return ctx.accessToken
}

func (ctx *Session) GetClearance() string {
	return ctx.clearance
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
	return (time.Now().Unix() + int64((5 * time.Minute).Seconds())) >= time.Unix(user.Exp, 0).Unix()
}

func (ctx *Session) ClearanceIsInValid() bool {
	now := time.Now().Unix()
	val := now - ctx.clearanceCreated
	t := cfExpire - (10 * time.Minute)
	return val > int64(t.Seconds())
}

func (ctx *Session) waitResolveReCaptcha(timeout time.Duration) chromedp.EmulateAction {
	return chromedp.ActionFunc(func(tab context.Context) error {
		ctxWithTimeout, cancel := context.WithTimeout(context.TODO(), timeout)
		defer cancel()
		for {
			select {
			case <-ctxWithTimeout.Done():
				ctx.screenshot(tab, "recaptcha.png")
				return errors.New("solve the reCAPTCHA error: " + ctxWithTimeout.Err().Error())
			default:
				var value string
				_ = chromedp.Value("#g-recaptcha-response", &value).Do(tab)
				if value != "" {
					return nil
				}
			}
			time.Sleep(time.Second)
		}
	})
}

func (ctx *Session) screenshot(tab context.Context, filename string) {
	var content []byte
	filename = GetTempDir() + "chatgpt_error_" + filename
	err := chromedp.FullScreenshot(&content, 90).Do(tab)
	if err != nil {
		return
	}
	_ = os.WriteFile(filename, content, 0777)
}

func (ctx *Session) waitElement(sel string, timeout time.Duration) chromedp.EmulateAction {
	return chromedp.ActionFunc(func(tab context.Context) error {
		ctxWithTimeout, cancel := context.WithTimeout(context.TODO(), timeout)
		defer cancel()
		for {
			select {
			case <-ctxWithTimeout.Done():
				name := base64.StdEncoding.EncodeToString([]byte(sel))
				name = strings.Replace(name, "=", "", -1)
				ctx.screenshot(tab, name+".png")
				return errors.New("wait for " + sel + " element error: " + ctxWithTimeout.Err().Error())
			default:
				var isValid bool
				_ = chromedp.EvaluateAsDevTools(fmt.Sprintf(`document.querySelector("%s") !== null`, sel), &isValid).Do(tab)
				if isValid {
					return nil
				}
			}
			time.Sleep(500 * time.Millisecond)
		}
	})
}

func (ctx *Session) readClearance(clearance *string) chromedp.EmulateAction {
	return chromedp.ActionFunc(func(tab context.Context) error {
		ctxWithTimeout, cancel := context.WithTimeout(context.TODO(), time.Second*10)
		defer cancel()
		for {
			select {
			case <-ctxWithTimeout.Done():
				ctx.screenshot(tab, "clearance.png")
				return errors.New("error setting clearance: " + ctxWithTimeout.Err().Error())
			default:
				cookies, err := network.GetCookies().Do(tab)
				if err != nil {
					return errors.New("error getting cookie: " + err.Error())
				}
				for _, cookie := range cookies {
					if cookie.Name == "cf_clearance" && cookie.Domain == ".chat.openai.com" {
						*clearance = cookie.Value
						return nil
					}
				}
			}
			time.Sleep(500 * time.Millisecond)
		}
	})
}

func (ctx *Session) RefreshToken() error {
	if ctx.username != "" && ctx.password != "" {
		tab, closeTab := chromedp.NewContext(ctx.browser.Context)
		defer closeTab()
		tabTimeout, closeTabTimeout := context.WithTimeout(tab, 2*time.Minute)
		defer closeTabTimeout()
		if err := chromedp.Run(tabTimeout,
			chromedp.Navigate("https://chat.openai.com/auth/login"),
			ctx.waitElement(".btn:nth-child(1)", 30*time.Second),
			chromedp.Click(".btn:nth-child(1)"),
			ctx.waitElement("#username", 30*time.Second),
			chromedp.SetValue("#username", ctx.username),
			ctx.waitResolveReCaptcha(time.Minute),
			chromedp.Sleep(2*time.Second),
			chromedp.Click("button[type='submit']"),
			ctx.waitElement("#password", 30*time.Second),
			chromedp.SetValue("#password", ctx.password),
			ctx.waitElement("button[type='submit']", 30*time.Second),
			chromedp.Click("button[type='submit']"),
			ctx.waitElement("#__next", 30*time.Second),
			chromedp.Navigate("https://chat.openai.com/api/auth/session"),
			ctx.waitElement("pre", 30*time.Second),
			ctx.readClearance(&ctx.clearance),
			chromedp.EvaluateAsDevTools(`JSON.parse(document.querySelector("pre").innerHTML).accessToken`, &ctx.accessToken),
		); err != nil {
			return errors.New("login to chatgpt failed: " + err.Error())
		}
		ctx.clearanceCreated = time.Now().Unix()
		return nil
	}
	return nil
}

func (ctx *Session) RefreshClearance() error {
	tab, closeTab := chromedp.NewContext(ctx.browser.Context)
	defer closeTab()
	tabTimeout, closeTabTimeout := context.WithTimeout(tab, time.Minute)
	defer closeTabTimeout()
	if err := chromedp.Run(tabTimeout,
		chromedp.Navigate("https://chat.openai.com/auth/login"),
		ctx.waitElement(".btn:nth-child(1)", 30*time.Second),
		ctx.readClearance(&ctx.clearance),
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
				log.Println("start refresh token")
				if err := ctx.RefreshToken(); err != nil {
					log.Println("refresh token failed: " + err.Error())
				} else {
					log.Println("refresh token success")
				}
			} else {
				if ctx.ClearanceIsInValid() {
					_ = ctx.RefreshClearance()
				}
			}
			time.Sleep(time.Second * 5)
		}
	}()
	return ctx
}

func NewSessionWithCredential(browser *Browser, username, password string) *Session {
	return &Session{
		browser:  browser,
		username: username,
		password: password,
	}
}

func NewSessionWithAccessToken(browser *Browser, accessToken string) *Session {
	return &Session{
		browser:     browser,
		accessToken: accessToken,
	}
}
