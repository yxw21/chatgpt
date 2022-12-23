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

type OpenAI struct {
	username string
	password string
	key      string
}

func (ctx *OpenAI) waitResolveReCaptcha(timeout time.Duration) chromedp.EmulateAction {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		ctxWithTimeout, cancel := context.WithTimeout(context.TODO(), timeout)
		defer cancel()
		for {
			select {
			case <-ctxWithTimeout.Done():
				screenshot(ctx, "recaptcha.png")
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

func (ctx *OpenAI) setExtension() (string, error) {
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
	addBytes := []byte(fmt.Sprintf(`Settings.set({id: "key", value: "%s"});`, ctx.key))
	contentBytes = append(contentBytes, addBytes...)
	err = os.WriteFile(background, contentBytes, 0777)
	if err != nil {
		return dist, errors.New("error updating js file: " + err.Error())
	}
	return dist, nil
}

func (ctx *OpenAI) GetPassport() (*Passport, error) {
	var data = &Passport{}
	dist, err := ctx.setExtension()
	if err != nil {
		return data, err
	}
	chromeCtx, cancel, err := chromedpundetected.New(chromedpundetected.NewConfig(
		chromedpundetected.WithHeadless(),
		chromedpundetected.WithTimeout(2*time.Minute),
		chromedpundetected.WithChromeFlags(chromedp.Flag("disable-extensions-except", dist+"chrome")),
	))
	if err != nil {
		return data, errors.New("error creating chrome context: " + err.Error())
	}
	defer cancel()
	if err = chromedp.Run(chromeCtx,
		chromedp.Navigate("https://chat.openai.com/auth/login"),
		waitElement(".btn:nth-child(1)", time.Minute),
		// Avoid ajax request failure and the page will not jump.
		clickElement(".btn:nth-child(1)", 5),
		waitElement("#username", time.Minute),
		chromedp.SetValue("#username", ctx.username),
		ctx.waitResolveReCaptcha(time.Minute),
		chromedp.Sleep(2*time.Second),
		chromedp.Click("button[type='submit']"),
		waitElement("#password", time.Minute),
		chromedp.SetValue("#password", ctx.password),
		waitElement("button[type='submit']", time.Minute),
		chromedp.Click("button[type='submit']"),
		waitElement("#__next", time.Minute),
		chromedp.Navigate("https://chat.openai.com/api/auth/session"),
		waitElement("pre", time.Minute),
		setClearance(&data.Clearance),
		chromedp.EvaluateAsDevTools(`navigator.userAgent`, &data.Useragent),
		chromedp.EvaluateAsDevTools(`JSON.parse(document.querySelector("pre").innerHTML).accessToken`, &data.AccessToken),
	); err != nil {
		return data, errors.New("login to chatgpt failed: " + err.Error())
	}
	return data, nil
}

func NewOpenAI(username, password, key string) *OpenAI {
	return &OpenAI{username: username, password: password, key: key}
}
