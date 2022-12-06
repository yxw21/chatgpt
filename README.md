# chatgpt

chatgpt api

# usage

```
package main

import (
	"fmt"
	"github.com/yxw21/chatgpt"
)

func main() {
	chat := chatgpt.NewChat("{jwt token}")
	res, err := chat.Send("hi")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(res)
}
```
