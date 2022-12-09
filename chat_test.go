package chatgpt

import (
	"fmt"
	"testing"
)

func TestChat(t *testing.T) {
	chat := NewChat("{session token}")
	res, err := chat.Send("hi")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res.Message.Content.Parts)
}
