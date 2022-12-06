# chatgpt

chatgpt api

jwt token

<img width="1672" alt="image" src="https://user-images.githubusercontent.com/16237562/205948470-86ccf237-c8cc-4017-8bc5-fcdae1f91cb3.png">


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
