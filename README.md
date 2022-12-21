# Dependency
- Linux needs to install `xvfb`

On Ubuntu or Debian
```
apt update
apt install xvfb
```
On CentOS
```
yum update
yum install xorg-x11-server-Xvfb
```
- You need a key to crack the verification code, you can go to the website `nopecha.com` to register, it is very cheap.

```
https://nopecha.com
```
- chrome

On Ubuntu or Debian
```
wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb
apt install ./google-chrome-stable_current_amd64.deb
```
On CentOS
```
wget https://dl.google.com/linux/direct/google-chrome-stable_current_x86_64.rpm
yum localinstall -y google-chrome-stable_current_x86_64.rpm
```
On Alpine
```
apk add chromium
```

# Login
1. username and password
```golang
package main

import (
	"fmt"
	"time"
	"github.com/yxw21/chatgpt"
)

func main() {
	chat := chatgpt.NewChat("username", "password", "key")
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
