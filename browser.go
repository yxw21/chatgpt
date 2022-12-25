package chatgpt

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Davincible/chromedp-undetected"
	"github.com/chromedp/chromedp"
	"github.com/mholt/archiver/v3"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Browser struct {
	Context context.Context
}

func (ctx *Browser) waitResolveReCaptcha(timeout time.Duration) chromedp.EmulateAction {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		ctxWithTimeout, cancel := context.WithTimeout(context.TODO(), timeout)
		defer cancel()
		for {
			select {
			case <-ctxWithTimeout.Done():
				return errors.New("solve the reCAPTCHA error: " + ctxWithTimeout.Err().Error())
			default:
				var value string
				_ = chromedp.Value("#g-recaptcha-response", &value).Do(ctx)
				if value != "" {
					return nil
				}
			}
			time.Sleep(time.Second)
		}
	})
}

func (ctx *Browser) GetChatGPTPassport(username, password string) (*Passport, error) {
	var passport = &Passport{}
	if err := chromedp.Run(ctx.Context,
		chromedp.Navigate("https://chat.openai.com/auth/login"),
		waitElement(".btn:nth-child(1)", 30*time.Second),
		chromedp.Click(".btn:nth-child(1)"),
		waitElement("#username", 30*time.Second),
		chromedp.SetValue("#username", username),
		ctx.waitResolveReCaptcha(time.Minute),
		chromedp.Sleep(2*time.Second),
		chromedp.Click("button[type='submit']"),
		waitElement("#password", 30*time.Second),
		chromedp.SetValue("#password", password),
		waitElement("button[type='submit']", 30*time.Second),
		chromedp.Click("button[type='submit']"),
		waitElement("#__next", 30*time.Second),
		chromedp.Navigate("https://chat.openai.com/api/auth/session"),
		waitElement("pre", 30*time.Second),
		readClearance(&passport.Clearance),
		chromedp.EvaluateAsDevTools(`navigator.userAgent`, &passport.Useragent),
		chromedp.EvaluateAsDevTools(`JSON.parse(document.querySelector("pre").innerHTML).accessToken`, &passport.AccessToken),
	); err != nil {
		return passport, errors.New("login to chatgpt failed: " + err.Error())
	}
	return passport, nil
}

func (ctx *Browser) setExtension(key string) (string, error) {
	var release []Release
	separator := string(filepath.Separator)
	tempDir := GetTempDir()
	userHomeDir, err := GetUserHomeDir()
	if err != nil {
		return "", err
	}
	gz := tempDir + "dist.tar.gz"
	dest := userHomeDir
	dist := dest + "dist" + separator
	background := dist + fmt.Sprintf("chrome%sbackground.js", separator)
	_, err = os.Stat(dist)
	if err == nil {
		return dist, nil
	}
	resp, err := http.Get("https://api.github.com/repos/yxw21/nopecha-extension/releases")
	if err != nil {
		return "", errors.New("error getting nopecha-extension information: " + err.Error())
	}
	if err = json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", errors.New("json parsing github response error: " + err.Error())
	}
	resp.Body.Close()
	downloadUrl := release[0].Assets[0].BrowserDownloadUrl
	resp, err = http.Get(downloadUrl)
	if err != nil {
		return "", errors.New("error downloading nopecha-extension: " + err.Error())
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	if _, err = buf.ReadFrom(bufio.NewReader(resp.Body)); err != nil {
		return "", errors.New("error reading buffer from response content: " + err.Error())
	}
	if err = os.WriteFile(gz, buf.Bytes(), 0777); err != nil {
		return "", errors.New("error writing extension to temporary directory: " + err.Error())
	}
	_ = os.RemoveAll(dist)
	if err = archiver.Unarchive(gz, dest); err != nil {
		return "", errors.New("error unpacking extension: " + err.Error())
	}
	contentBytes, err := os.ReadFile(background)
	if err != nil {
		return "", errors.New("error reading js file: " + err.Error())
	}
	addBytes := []byte(fmt.Sprintf(`Settings.set({id: "key", value: "%s"});`, key))
	contentBytes = append(contentBytes, addBytes...)
	err = os.WriteFile(background, contentBytes, 0777)
	if err != nil {
		return dist, errors.New("error updating js file: " + err.Error())
	}
	return dist, nil
}

func NewBrowser(extensionKey string) (*Browser, context.CancelFunc, error) {
	config := chromedpundetected.NewConfig(chromedpundetected.WithHeadless())
	browser := &Browser{}
	if extensionKey != "" {
		dist, err := browser.setExtension(extensionKey)
		if err != nil {
			return nil, nil, err
		}
		config = chromedpundetected.NewConfig(
			chromedpundetected.WithHeadless(),
			chromedpundetected.WithChromeFlags(chromedp.Flag("disable-extensions-except", dist+"chrome")),
		)
	}
	chromeCtx, cancel, err := chromedpundetected.New(config)
	if err != nil {
		return nil, nil, errors.New("error creating chrome context: " + err.Error())
	}
	browser.Context = chromeCtx
	err = chromedp.Run(browser.Context, chromedp.Navigate("https://chat.openai.com/auth/login"))
	if err != nil {
		return nil, nil, err
	}
	return browser, cancel, nil
}
