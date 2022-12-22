package chatgpt

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

func (ctx *OpenAI) setClearance(clearance *string) chromedp.EmulateAction {
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
					if cookie.Name == "cf_clearance" && cookie.Domain == ".chat.openai.com" {
						*clearance = cookie.Value
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
	separator := string(filepath.Separator)
	tmp := os.TempDir()
	dirname, err := os.UserHomeDir()
	if err != nil {
		return "", errors.New("get user home dir error: " + err.Error())
	}
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
	_, err = os.Stat(dist)
	if err == nil {
		return dist, nil
	}
	resp, err := http.Get("https://api.github.com/repos/yxw21/nopecha-extension/releases")
	if err != nil {
		return "", errors.New("getting nopecha-extension release error: " + err.Error())
	}
	if err = json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", errors.New("github json decode error: " + err.Error())
	}
	resp.Body.Close()
	downloadUrl := release[0].Assets[0].BrowserDownloadUrl
	resp, err = http.Get(downloadUrl)
	if err != nil {
		return "", errors.New("download nopecha-extension error: " + err.Error())
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
	return dist, errors.New("error updating js file: " + err.Error())
}

func (ctx *OpenAI) GetData() (*Data, error) {
	var data = &Data{}
	dist, err := ctx.setExtension()
	if err != nil {
		return data, err
	}
	chromeCtx, cancel, err := chromedpundetected.New(chromedpundetected.NewConfig(
		chromedpundetected.WithHeadless(),
		chromedpundetected.WithTimeout(5*time.Minute),
		chromedpundetected.WithChromeFlags(chromedp.Flag("disable-extensions-except", dist+"chrome")),
	))
	if err != nil {
		return data, errors.New("error starting chrome: " + err.Error())
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
		ctx.setClearance(&data.Clearance),
		chromedp.Navigate("https://chat.openai.com/api/auth/session"),
		chromedp.WaitVisible("pre"),
		chromedp.EvaluateAsDevTools(`navigator.userAgent`, &data.Useragent),
		chromedp.EvaluateAsDevTools(`JSON.parse(document.querySelector("pre").innerHTML).accessToken`, &data.AccessToken),
	); err != nil {
		return data, errors.New("get data error: " + err.Error())
	}
	return data, nil
}

func NewOpenAI(username, password, key string) *OpenAI {
	return &OpenAI{username: username, password: password, key: key}
}
