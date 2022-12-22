package chatgpt

import (
	"fmt"
	"testing"
)

func TestOpenAI(t *testing.T) {
	openai := NewOpenAI("example@gmail.com", "123456", "I-1123123KASD")
	data, err := openai.GetData()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(data)
}
