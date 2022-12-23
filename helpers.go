package chatgpt

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func GetUserHomeDir() (string, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return "", errors.New("error getting user home folder: " + err.Error())
	}
	separator := string(filepath.Separator)
	if !strings.HasSuffix(userHomeDir, separator) {
		userHomeDir += separator
	}
	return userHomeDir, nil
}

func GetTempDir() string {
	tempDir := os.TempDir()
	separator := string(filepath.Separator)
	if !strings.HasSuffix(tempDir, separator) {
		tempDir += separator
	}
	return tempDir
}

func screenshot(ctx context.Context, filename string) {
	var content []byte
	filename = GetTempDir() + "chatgpt_error_" + filename
	err := chromedp.FullScreenshot(&content, 90).Do(ctx)
	if err != nil {
		return
	}
	_ = os.WriteFile(filename, content, 0777)
}

func waitElement(sel string, timeout time.Duration) chromedp.EmulateAction {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		ctxWithTimeout, cancel := context.WithTimeout(context.TODO(), timeout)
		defer cancel()
		for {
			select {
			case <-ctxWithTimeout.Done():
				name := base64.StdEncoding.EncodeToString([]byte(sel))
				name = strings.Replace(name, "=", "", -1)
				screenshot(ctx, name+".png")
				return errors.New("wait for " + sel + " element error: " + ctxWithTimeout.Err().Error())
			default:
				var isValid bool
				_ = chromedp.EvaluateAsDevTools(fmt.Sprintf(`document.querySelector("%s") !== null`, sel), &isValid).Do(ctx)
				if isValid {
					return nil
				}
			}
			time.Sleep(500 * time.Millisecond)
		}
	})
}

func clickElement(sel string, number int) chromedp.EmulateAction {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		var isValid bool
		_ = chromedp.EvaluateAsDevTools(fmt.Sprintf(`(function(sel, count){
for(let i = 1; i <= count; i++){
    var evt = document.createEvent('Event');
    evt.initEvent('click',true,true);
    document.querySelector(sel).dispatchEvent(evt);
}
})("%s", %d)==false`, sel, number), &isValid).Do(ctx)
		return nil
	})
}

func setClearance(clearance *string) chromedp.EmulateAction {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		ctxWithTimeout, cancel := context.WithTimeout(context.TODO(), time.Second*10)
		defer cancel()
		for {
			select {
			case <-ctxWithTimeout.Done():
				screenshot(ctx, "clearance.png")
				return errors.New("error setting clearance: " + ctxWithTimeout.Err().Error())
			default:
				cookies, err := network.GetCookies().Do(ctx)
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
