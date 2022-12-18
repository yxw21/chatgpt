# openai升级了安全策略
- 使用cloudflare保护chatgpt
- 用户名密码登录需要谷歌验证码(无法绕过)
- 聊天加入了请求限制(请求限制太死，可参考官网)

等后续官方API吧

<del>

# Login
1. username and password
```golang
package main

import (
	"fmt"
	"github.com/yxw21/chatgpt"
)

func main() {
	chat := chatgpt.NewChat("username", "password")
	res, err := chat.Send("hi")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(res.Message.Content.Parts)
}
```
2.session token login
```golang
package main

import (
	"fmt"
	"github.com/yxw21/chatgpt"
)

func main() {
	chat := chatgpt.NewChatWithSessionToken("{session token}")
	res, err := chat.Send("hi")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(res.Message.Content.Parts)
}
```
# Session token (expires in about a month)
<img width="945" alt="image" src="https://user-images.githubusercontent.com/16237562/206679314-7d412b03-98fc-422d-92bb-2d4a19f375b8.png">

</del>
