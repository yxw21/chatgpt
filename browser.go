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
)

type Browser struct {
	Context context.Context
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
	config := chromedpundetected.NewConfig(
		chromedpundetected.WithHeadless(),
		chromedpundetected.WithChromeFlags(chromedp.Flag("disable-dev-shm-usage", true)),
	)
	browser := &Browser{}
	if extensionKey != "" {
		dist, err := browser.setExtension(extensionKey)
		if err != nil {
			return nil, nil, err
		}
		config = chromedpundetected.NewConfig(
			chromedpundetected.WithHeadless(),
			chromedpundetected.WithChromeFlags(
				chromedp.Flag("disable-extensions-except", dist+"chrome"),
				chromedp.Flag("disable-dev-shm-usage", true),
			),
		)
	}
	chromeCtx, cancel, err := chromedpundetected.New(config)
	if err != nil {
		return nil, nil, errors.New("error creating chrome context: " + err.Error())
	}
	browser.Context = chromeCtx
	err = chromedp.Run(browser.Context, chromedp.Navigate("https://chat.openai.com/api/auth/session"))
	if err != nil {
		return nil, nil, err
	}
	return browser, cancel, nil
}
