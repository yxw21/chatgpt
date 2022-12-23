package chatgpt

import (
	"fmt"
	"testing"
)

func TestOpenAI(t *testing.T) {
	retry := 3
	openai := NewOpenAI("example@gmail.com", "password", "I-ASDA123ASDA")
	for i := 0; i < retry; i++ {
		data, err := openai.GetPassport()
		if err == nil {
			fmt.Println(data)
			break
		}
		if i == retry-1 {
			t.Fatal(err)
		}
	}
}
