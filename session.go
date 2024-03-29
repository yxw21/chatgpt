package chatgpt

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"log"
	"os"
	"strings"
	"time"
)

const cfExpire = 120 * time.Minute

var (
	ErrorAccessTokenIsEmpty     = errors.New("accessToken does not provide")
	ErrorAccessTokenGet         = errors.New("the accessToken is being obtained")
	ErrorAccessTokenFormatError = errors.New("the provided accessToken is in the wrong format")
	ErrorAccessTokenExpires     = errors.New("the provided accessToken has expired")
	ErrorClearanceExpires       = errors.New("clearance has expired")
)

type Session struct {
	Browser          *Browser
	Username         string
	Password         string
	AccessToken      string
	clearance        string
	clearanceCreated int64
}

func (ctx *Session) GetClearance() string {
	return ctx.clearance
}

func (ctx *Session) getExpFromAccessToken() (int64, error) {
	var user User
	if ctx.AccessToken == "" {
		if ctx.Username != "" && ctx.Password != "" {
			return 0, ErrorAccessTokenGet
		}
		return 0, ErrorAccessTokenIsEmpty
	}
	sli := strings.Split(ctx.AccessToken, ".")
	if len(sli) != 3 {
		return 0, ErrorAccessTokenFormatError
	}
	bs, err := base64.StdEncoding.DecodeString(sli[1])
	if err != nil {
		bs, err = base64.RawStdEncoding.DecodeString(sli[1])
		if err != nil {
			return 0, ErrorAccessTokenFormatError
		}
	}
	if err = json.Unmarshal(bs, &user); err != nil {
		return 0, ErrorAccessTokenFormatError
	}
	return time.Unix(user.Exp, 0).Unix(), nil
}

func (ctx *Session) accessTokenShouldRefresh() bool {
	exp, err := ctx.getExpFromAccessToken()
	if err != nil {
		return true
	}
	sixHour := int64((6 * time.Hour).Seconds())
	return (time.Now().Unix() + sixHour) >= exp
}

func (ctx *Session) CheckAccessToken() error {
	exp, err := ctx.getExpFromAccessToken()
	if err != nil {
		return err
	}
	isInvalid := time.Now().Unix() >= exp
	if isInvalid {
		if ctx.Username != "" && ctx.Password != "" {
			return ErrorAccessTokenGet
		}
		return ErrorAccessTokenExpires
	}
	return nil
}

func (ctx *Session) clearanceShouldRefresh() bool {
	now := time.Now().Unix()
	val := now - ctx.clearanceCreated
	t := cfExpire - (30 * time.Minute)
	return val >= int64(t.Seconds())
}

func (ctx *Session) CheckClearance() error {
	now := time.Now().Unix()
	val := now - ctx.clearanceCreated
	if val >= int64(cfExpire.Seconds()) {
		return ErrorClearanceExpires
	}
	return nil
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

func (ctx *Session) setClearance(clearance string) {
	tab, closeTab := chromedp.NewContext(ctx.Browser.Context)
	defer closeTab()
	tabTimeout, closeTabTimeout := context.WithTimeout(tab, 30*time.Second)
	defer closeTabTimeout()
	if err := chromedp.Run(tabTimeout,
		chromedp.ActionFunc(func(ctx context.Context) error {
			expr := cdp.TimeSinceEpoch(time.Now().Add(365 * 24 * time.Hour))
			return network.SetCookie("cf_clearance", clearance).
				WithDomain(".chat.openai.com").
				WithPath("/").
				WithExpires(&expr).
				WithHTTPOnly(true).
				WithSecure(true).
				WithSameSite(network.CookieSameSiteNone).
				Do(ctx)
		})); err != nil {
		log.Println("error setting cloudflare cookie: " + err.Error())
	}
}

func (ctx *Session) RefreshToken() error {
	if ctx.Username != "" && ctx.Password != "" {
		tab, closeTab := chromedp.NewContext(ctx.Browser.Context)
		defer closeTab()
		tabTimeout, closeTabTimeout := context.WithTimeout(tab, 2*time.Minute)
		defer closeTabTimeout()
		if err := chromedp.Run(tabTimeout,
			chromedp.Navigate("https://chat.openai.com/auth/login"),
			ctx.waitElement(".btn:nth-child(1)", 30*time.Second),
			chromedp.Click(".btn:nth-child(1)"),
			ctx.waitElement("#username", 30*time.Second),
			chromedp.SetValue("#username", ctx.Username),
			//ctx.waitResolveReCaptcha(time.Minute),
			//chromedp.Sleep(2*time.Second),
			chromedp.Click("button[type='submit']"),
			ctx.waitElement("#password", 30*time.Second),
			chromedp.SetValue("#password", ctx.Password),
			ctx.waitElement("button[type='submit']", 30*time.Second),
			chromedp.Click("button[type='submit']"),
			ctx.waitElement("#__next", 30*time.Second),
			chromedp.Navigate("https://chat.openai.com/api/auth/session"),
			ctx.waitElement("pre", 30*time.Second),
			ctx.readClearance(&ctx.clearance),
			chromedp.EvaluateAsDevTools(`JSON.parse(document.querySelector("pre").innerHTML).accessToken`, &ctx.AccessToken),
		); err != nil {
			return errors.New("login to chatgpt failed: " + err.Error())
		}
		ctx.clearanceCreated = time.Now().Unix()
		return nil
	}
	return errors.New("no username and password set")
}

func (ctx *Session) RefreshClearance() error {
	browser, closeBrowser, err := NewBrowser(ctx.Browser.browserOptions)
	if err != nil {
		return err
	}
	defer closeBrowser()
	tab, closeTab := context.WithTimeout(browser.Context, time.Minute)
	defer closeTab()
	if err = chromedp.Run(tab,
		chromedp.Navigate("https://chat.openai.com/auth/login"),
		ctx.waitElement(".btn:nth-child(1)", 30*time.Second),
		ctx.readClearance(&ctx.clearance),
	); err != nil {
		return errors.New("error refreshing clearance: " + err.Error())
	}
	ctx.clearanceCreated = time.Now().Unix()
	ctx.setClearance(ctx.clearance)
	return nil
}

func (ctx *Session) AutoRefresh() *Session {
	go func() {
		for {
			if ctx.accessTokenShouldRefresh() && ctx.Username != "" && ctx.Password != "" {
				log.Println("start refresh token")
				if err := ctx.RefreshToken(); err != nil {
					log.Println("refresh token failed: " + err.Error())
				} else {
					log.Println("refresh token success")
				}
			} else {
				if ctx.clearanceShouldRefresh() {
					_ = ctx.RefreshClearance()
				}
			}
			time.Sleep(5 * time.Second)
		}
	}()
	return ctx
}
