# Session token (expires in about a month)
<img width="945" alt="image" src="https://user-images.githubusercontent.com/16237562/206679314-7d412b03-98fc-422d-92bb-2d4a19f375b8.png">

# Usage
```golang
package main

import (
	"fmt"
	"github.com/yxw21/chatgpt"
)

func main() {
	chat := chatgpt.NewChat("{session token}")
	res, err := chat.Send("hi")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(res.Message.Content.Parts)
}
```
