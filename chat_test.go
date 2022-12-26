package chatgpt

import (
	"fmt"
	"testing"
	"time"
)

func waitChat(chat *Chat) {
	for {
		if err := chat.Check(); err != nil {
			fmt.Println(err)
		} else {
			break
		}
		time.Sleep(5 * time.Second)
	}
}

func startChat(chat *Chat, message string, retry int) {
	waitChat(chat)
	for i := 0; i < retry; i++ {
		res, err := chat.Send(message)
		if err == nil {
			fmt.Println(res.Message.Content.Parts)
			break
		}
		if i == retry-1 {
			fmt.Println(err)
		}
	}
}

func TestChatWithCredential(t *testing.T) {
	retry := 3
	browser, closeBrowser, err := NewBrowser(BrowserOptions{
		ExtensionKey: "I-ABCDEFGHIJKL",
		//Proxy:        "socks5://38.91.107.224:36699",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer closeBrowser()
	// use username and password
	session := NewSessionWithCredential(browser, "example@gmail.com", "password").AutoRefresh()
	chat := NewChat(browser, session)
	startChat(chat, "hi", retry)
	startChat(chat, "hello", retry)
	startChat(chat, "hey", retry)
}

func TestChatWithAccessToken(t *testing.T) {
	retry := 3
	browser, closeBrowser, err := NewBrowser(BrowserOptions{
		//Proxy: "socks5://38.91.107.224:36699",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer closeBrowser()
	// use access token
	session := NewSessionWithAccessToken(browser, "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6Ik1UaEVOVUpHTkVNMVFURTRNMEZCTWpkQ05UZzVNRFUxUlRVd1FVSkRNRU13UmtGRVFrRXpSZyJ9").AutoRefresh()
	chat := NewChat(browser, session)
	startChat(chat, "hi", retry)
	startChat(chat, "hello", retry)
	startChat(chat, "hey", retry)
}
