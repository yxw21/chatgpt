package chatgpt

import (
	"fmt"
	"testing"
)

func TestChatWithCredential(t *testing.T) {
	retry := 3
	// use username and password
	session := NewSessionWithCredential("example@gmail.com", "password", "I-ASDA123ASDA").AutoRefresh()
	chat := NewChat(session)
	for i := 0; i < retry; i++ {
		res, err := chat.Send("hi")
		if err == nil {
			fmt.Println(res.Message.Content.Parts)
			break
		}
		if i == retry-1 {
			t.Fatal(err)
		}
	}
}

func TestChatWithAccessToken(t *testing.T) {
	retry := 3
	// use access token
	session := NewSessionWithAccessToken("jwt").AutoRefresh()
	chat := NewChat(session)
	for i := 0; i < retry; i++ {
		res, err := chat.Send("hi")
		if err == nil {
			fmt.Println(res.Message.Content.Parts)
			break
		}
		if i == retry-1 {
			t.Fatal(err)
		}
	}
}
