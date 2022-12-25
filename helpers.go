package chatgpt

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"
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

func readClearance(clearance *string) chromedp.EmulateAction {
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

func setCookie(name, value, domain, path string, expires time.Time, httpOnly, secure bool, sameSite network.CookieSameSite) chromedp.EmulateAction {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		expr := cdp.TimeSinceEpoch(expires)
		return network.SetCookie(name, value).
			WithDomain(domain).
			WithPath(path).
			WithExpires(&expr).
			WithHTTPOnly(httpOnly).
			WithSecure(secure).
			WithSameSite(sameSite).
			Do(ctx)
	})
}

func sendRequest(accessToken string, body []byte, response any) chromedp.EmulateAction {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		return chromedp.EvaluateAsDevTools(fmt.Sprintf(`new Promise((resolve) => {
  fetch("https://chat.openai.com/backend-api/conversation", {
    "headers": {
      "authorization": "Bearer %s",
      "content-type": "application/json",
      "x-openai-assistant-app-id": ""
    },
    "body": JSON.stringify(%s),
    "method": "POST",
    "mode": "cors",
    "credentials": "include"
  })
  .then(context => context.text())
  .then(content => {
    let arr = content.split("\n\n");
    let len = arr.length;
    let index=len-3;
    if(index > -1 && index < len){
      content = arr[index];
      arr = content.split("data: ");
      if(arr.length > 1){
        resolve(JSON.parse(arr[1]));
        return;
      }
    }
    resolve(content);
  })
  .catch(e => resolve(e.toString()));
})`, accessToken, string(body)), &response, func(p *runtime.EvaluateParams) *runtime.EvaluateParams {
			return p.WithAwaitPromise(true)
		}).Do(ctx)
	})
}

func ConvertMapToStruct(data map[string]any, result any) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, &result)
}
