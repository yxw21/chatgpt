package chatgpt

import (
	"log"
	"testing"
	"time"
)

func startChat(chat *Chat, message string) {
	for {
		if err := chat.Check(); err != nil {
			log.Println(err)
		} else {
			break
		}
		time.Sleep(5 * time.Second)
	}
	res, err := chat.Send(message)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(res.Message.Content.Parts)
}

func TestChat(t *testing.T) {
	browser, closeBrowser, err := NewBrowser(BrowserOptions{
		ExtensionKey: "I-ABCDEFGHIJKL",
		//Proxy:        "socks5://38.91.107.224:36699",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer closeBrowser()
	// Make sure the session is initialized once
	session := (&Session{
		Browser:     browser,
		Username:    "example@gmail.com",
		Password:    "password",
		AccessToken: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6Ik1UaEVOVUpHTkVNMVFURTRNMEZCTWpkQ05UZzVNRFUxUlRVd1FVSkRNRU13UmtGRVFrRXpSZyJ9",
	}).AutoRefresh()
	chat := NewChat(browser, session)
	startChat(chat, "hi")
	startChat(chat, "hello")
	startChat(chat, "hey")
}
