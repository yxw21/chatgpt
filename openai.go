package chatgpt

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Davincible/chromedp-undetected"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/mholt/archiver/v3"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type OpenAI struct {
	username string
	password string
	key      string
}

func (ctx *OpenAI) waitResolveCapture() chromedp.EmulateAction {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		ctxWithTimeout, cancel := context.WithTimeout(context.TODO(), time.Minute)
		defer cancel()
		for {
			select {
			case <-ctxWithTimeout.Done():
				return ctxWithTimeout.Err()
			default:
				var value string
				chromedp.Value("#g-recaptcha-response", &value).Do(ctx)
				if value != "" {
					return nil
				}
			}
			time.Sleep(time.Second)
		}
	})
}

func (ctx *OpenAI) waitCookie(token *string) chromedp.EmulateAction {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		ctxWithTimeout, cancel := context.WithTimeout(context.TODO(), time.Minute)
		defer cancel()
		for {
			select {
			case <-ctxWithTimeout.Done():
				return ctxWithTimeout.Err()
			default:
				cookies, err := network.GetCookies().Do(ctx)
				if err != nil {
					return err
				}
				for _, cookie := range cookies {
					if cookie.Name == "__Secure-next-auth.session-token" && cookie.Domain == "chat.openai.com" {
						*token = cookie.Value
						return nil
					}
				}
			}
			time.Sleep(time.Second)
		}
	})
}

func (ctx *OpenAI) setExtension() (string, error) {
	var release []Release
	resp, err := http.Get("https://api.github.com/repos/yxw21/nopecha-extension/releases")
	if err != nil {
		return "", err
	}
	if err = json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}
	resp.Body.Close()
	downloadUrl := release[0].Assets[0].BrowserDownloadUrl
	resp, err = http.Get(downloadUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	if _, err = buf.ReadFrom(bufio.NewReader(resp.Body)); err != nil {
		return "", err
	}
	dirname, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	separator := string(filepath.Separator)
	tmp := os.TempDir()
	if !strings.HasSuffix(tmp, separator) {
		tmp += separator
	}
	if !strings.HasSuffix(dirname, separator) {
		dirname += separator
	}
	gz := tmp + "dist.tar.gz"
	dest := dirname
	dist := dest + "dist" + separator
	background := dist + fmt.Sprintf("chrome%sbackground.js", separator)
	if err = os.WriteFile(gz, buf.Bytes(), 0777); err != nil {
		return "", err
	}
	_ = os.RemoveAll(dist)
	if err = archiver.Unarchive(gz, dest); err != nil {
		return "", err
	}
	contentBytes, err := os.ReadFile(background)
	if err != nil {
		return "", err
	}
	addBytes := []byte(fmt.Sprintf(`Settings.set({id: "key", value: "%s"});`, ctx.key))
	contentBytes = append(contentBytes, addBytes...)
	err = os.WriteFile(background, contentBytes, 0777)
	return dist, err
}

func (ctx *OpenAI) GetToken() (string, error) {
	var token string
	dist, err := ctx.setExtension()
	if err != nil {
		return token, err
	}
	chromeCtx, cancel, err := chromedpundetected.New(chromedpundetected.NewConfig(
		chromedpundetected.WithHeadless(),
		chromedpundetected.WithTimeout(5*time.Minute),
		chromedpundetected.WithChromeFlags(chromedp.Flag("disable-extensions-except", dist+"chrome")),
	))
	if err != nil {
		return token, err
	}
	defer cancel()
	if err = chromedp.Run(chromeCtx,
		chromedp.Navigate("https://chat.openai.com/auth/login"),
		chromedp.WaitVisible(".btn:nth-child(1)"),
		chromedp.Click(".btn:nth-child(1)"),
		chromedp.WaitVisible("#username"),
		chromedp.SetValue("#username", ctx.username),
		ctx.waitResolveCapture(),
		chromedp.Sleep(time.Second),
		chromedp.Click("button[type='submit']"),
		chromedp.WaitVisible("#password"),
		chromedp.SetValue("#password", ctx.password),
		chromedp.WaitVisible("button[type='submit']"),
		chromedp.Click("button[type='submit']"),
		ctx.waitCookie(&token),
	); err != nil {
		return token, err
	}
	return token, nil
}

func NewOpenAI(username, password, key string) *OpenAI {
	return &OpenAI{username: username, password: password, key: key}
}
