package chatgpt

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

type OpenAI struct {
	client   *http.Client
	Username string
	Password string
}

func (ctx *OpenAI) getStateFromHeader(header http.Header) string {
	location := header.Get("Location")
	sli := strings.Split(location, "=")
	if len(sli) < 2 {
		return ""
	}
	return sli[1]
}

func (ctx *OpenAI) request(method, url string, header map[string]string, body io.Reader) (http.Header, []byte, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, nil, err
	}
	for key, value := range header {
		req.Header.Set(key, value)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:107.0) Gecko/20100101 Firefox/107.0")
	resp, err := ctx.client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	contentBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	return resp.Header, contentBytes, nil
}

func (ctx *OpenAI) getAuthorizeUrl() (string, error) {
	var m = make(map[string]any)
	_, data, err := ctx.request("POST", "https://chat.openai.com/api/auth/signin/auth0?prompt=login", map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}, bytes.NewBufferString("callbackUrl=%2F&csrfToken=c9fc955397dc53bad7ea1025ac2d57647059ec13c1d73da3dd5b8fa1552bb82b&json=true"))
	if err != nil {
		return "", err
	}
	if err = json.Unmarshal(data, &m); err != nil {
		return "", err
	}
	if u, ok := m["url"].(string); ok {
		return u, nil
	}
	return "", errors.New("getAuthorizeUrl error")
}

func (ctx *OpenAI) getIdentifierState(authorizeUrl string) (string, error) {
	header, _, err := ctx.request("GET", authorizeUrl, nil, nil)
	if err != nil {
		return "", err
	}
	state := ctx.getStateFromHeader(header)
	if state == "" {
		return "", errors.New("getStateFromHeader error")
	}
	return state, nil
}

func (ctx *OpenAI) getPasswordState(identifierState string) (string, error) {
	header, _, err := ctx.request("POST", "https://auth0.openai.com/u/login/identifier?state="+identifierState, map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}, bytes.NewBufferString(fmt.Sprintf("username=%s&js-available=true&webauthn-available=true&is-brave=false&webauthn-platform-available=false&action=default&state=%s", ctx.Username, identifierState)))
	if err != nil {
		return "", err
	}
	state := ctx.getStateFromHeader(header)
	if state == "" {
		return "", errors.New("getStateFromHeader error")
	}
	return state, nil
}

func (ctx *OpenAI) getResumeState(passwordState string) (string, error) {
	header, _, err := ctx.request("POST", "https://auth0.openai.com/u/login/password?state="+passwordState, map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}, bytes.NewBufferString(fmt.Sprintf("username=%s&password=%s&action=default&state=%s", ctx.Username, ctx.Password, passwordState)))
	if err != nil {
		return "", err
	}
	state := ctx.getStateFromHeader(header)
	if state == "" {
		return "", errors.New("getStateFromHeader error")
	}
	return state, nil
}

func (ctx *OpenAI) getResumeLocation(resumeState string) (string, error) {
	header, _, err := ctx.request("GET", "https://auth0.openai.com/authorize/resume?state="+resumeState, nil, nil)
	if err != nil {
		return "", err
	}
	location := header.Get("Location")
	if location == "" {
		return "", errors.New("getResumeLocation error")
	}
	return location, nil
}

func (ctx *OpenAI) GetSessionToken() (string, error) {
	authorizeUrl, err := ctx.getAuthorizeUrl()
	if err != nil {
		return "", err
	}
	identifierState, err := ctx.getIdentifierState(authorizeUrl)
	if err != nil {
		return "", err
	}
	passwordState, err := ctx.getPasswordState(identifierState)
	if err != nil {
		return "", err
	}
	resumeState, err := ctx.getResumeState(passwordState)
	if err != nil {
		return "", err
	}
	location, err := ctx.getResumeLocation(resumeState)
	if err != nil {
		return "", err
	}
	_, _, err = ctx.request("GET", location, nil, nil)
	if err != nil {
		return "", err
	}
	cookies := ctx.client.Jar.Cookies(&url.URL{Scheme: "https", Host: "chat.openai.com"})
	for _, cookie := range cookies {
		if cookie.Name == "__Secure-next-auth.session-token" {
			return cookie.Value, nil
		}
	}
	return "", errors.New("not found __Secure-next-auth.session-token")
}

func NewOpenAI(username, password string) *OpenAI {
	jar, _ := cookiejar.New(nil)
	jar.SetCookies(&url.URL{Scheme: "https", Host: "chat.openai.com"}, []*http.Cookie{
		{Name: "__Host-next-auth.csrf-token", Value: "c9fc955397dc53bad7ea1025ac2d57647059ec13c1d73da3dd5b8fa1552bb82b%7C29dede65e8c76988e6ca0818cef992dcf2bfb5e008e6845d17c0e22361b1579e"},
		{Name: " __Secure-next-auth.callback-url", Value: "https%3A%2F%2Fchat.openai.com%2F"},
	})
	client := &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	return &OpenAI{
		client:   client,
		Username: username,
		Password: password,
	}
}
