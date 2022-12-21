package chatgpt

import (
	"fmt"
	"testing"
	"time"
)

func TestChat(t *testing.T) {
	// use username and password
	chat := NewChat("username", "password", "key")
	// wait token
	time.Sleep(time.Minute)
	res, err := chat.Send("hi")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res.Message.Content.Parts)
	// use session token
	chat = NewChatWithSessionToken("session token")
	res, err = chat.Send("hi")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res.Message.Content.Parts)
}
