package chatgpt

import (
	"fmt"
	"github.com/yxw21/chatgpt/session"
	"testing"
)

func TestChat(t *testing.T) {
	// use username and password
	chat := NewChat(chatgpt.NewSessionWithCredential("example@gmail.com", "123456", "I-1123123KASD").AutoRefresh())
	res, err := chat.Send("hi")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res.Message.Content.Parts)
}
