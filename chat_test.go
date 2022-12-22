package chatgpt

import (
	"fmt"
	"github.com/yxw21/chatgpt/session"
	"testing"
)

func TestChatWithCredential(t *testing.T) {
	// use username and password
	chat := NewChat(chatgpt.NewSessionWithCredential("example@gmail.com", "123456", "I-1123123KASD").AutoRefresh())
	res, err := chat.Send("hi")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(res.Message.Content.Parts)
}

func TestChatWithAccessToken(t *testing.T) {
	// use access token
	chat := NewChat(chatgpt.NewSessionWithAccessToken("jwt").AutoRefresh())
	res, err := chat.Send("hi")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(res.Message.Content.Parts)
}
